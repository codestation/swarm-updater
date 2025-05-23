/*
Copyright 2025 codestation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/distribution/reference"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/credentials"
	_ "github.com/docker/cli/cli/connhelper"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

const (
	serviceLabel        string = "xyz.megpoid.swarm-updater"
	updateOnlyLabel     string = "xyz.megpoid.swarm-updater.update-only"
	enabledServiceLabel string = "xyz.megpoid.swarm-updater.enable"
)

// Swarm struct to handle all the service operations
type Swarm struct {
	client      DockerClient
	Blacklist   []*regexp.Regexp
	LabelEnable bool
	MaxThreads  int
	// used to protect the service update when ran from cron and http endpoint at the same time
	mu sync.Mutex
}

func (c *Swarm) validService(service swarm.Service) bool {
	if c.LabelEnable {
		label := service.Spec.Labels[enabledServiceLabel]

		return strings.ToLower(label) == "true"
	}

	serviceName := service.Spec.Name

	for _, entry := range c.Blacklist {
		if entry.MatchString(serviceName) {
			return false
		}
	}

	return true
}

// NewSwarm instantiates a new Docker swarm client
func NewSwarm(configDir string, opts ...client.Opt) (*Swarm, error) {
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	configFile, err := config.Load(configDir)
	if err != nil {
		// https://github.com/docker/cli/issues/5075
		slog.Warn("failed to load config", "err", err)
	}

	if !configFile.ContainsAuth() {
		configFile.CredentialsStore = credentials.DetectDefaultStore(configFile.CredentialsStore)
	}

	return &Swarm{client: &dockerClient{apiClient: cli, configFile: configFile}, MaxThreads: 1}, nil
}

func (c *Swarm) serviceList(ctx context.Context) ([]swarm.Service, error) {
	services, err := c.client.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("ServiceList failed: %w", err)
	}

	return services, nil
}

func (c *Swarm) updateServiceWithRetries(ctx context.Context, service swarm.Service) error {
	var err error
	for i := 0; i < 3; i++ {
		err = c.updateService(ctx, service)
		if err == nil {
			return nil
		}

		// check if error has "update out of sequence" in the message
		if strings.Contains(err.Error(), "update out of sequence") {
			slog.Debug("Service update out of sequence, retrying with updated version", "service", service.Spec.Name)

			// fetch a newer service version
			updatedService, _, err := c.client.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
			if err != nil {
				return fmt.Errorf("ServiceInspect failed: %w", err)
			}

			service.Version = updatedService.Version
		} else {
			return err
		}
	}

	return fmt.Errorf("failed to update service %s after retries", service.Spec.Name)
}

func (c *Swarm) updateService(ctx context.Context, service swarm.Service) error {
	image := service.Spec.TaskTemplate.ContainerSpec.Image
	updateOpts := types.ServiceUpdateOptions{}

	// get docker auth
	encodedAuth, err := c.client.RetrieveAuthTokenFromImage(image)
	if err != nil {
		return fmt.Errorf("cannot retrieve auth token from service's image: %w", err)
	}

	// do not set auth if is an empty json object
	if encodedAuth != "e30=" {
		updateOpts.EncodedRegistryAuth = encodedAuth
	}

	// remove image hash from name
	imageName := strings.Split(image, "@sha")[0]

	// fetch a newer image digest
	service.Spec.TaskTemplate.ContainerSpec.Image, err = c.getImageDigest(ctx, imageName, updateOpts.EncodedRegistryAuth)
	if err != nil {
		return fmt.Errorf("failed to get new image digest: %w", err)
	}

	if image == service.Spec.TaskTemplate.ContainerSpec.Image {
		slog.Debug("Service is already up to date", "service", service.Spec.Name)

		return nil
	}

	if strings.ToLower(service.Spec.Labels[updateOnlyLabel]) == "true" {
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			*service.Spec.Mode.Replicated.Replicas = 0
		}
	}

	slog.Debug("Updating service", "service", service.Spec.Name)
	response, err := c.client.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update service %s: %w", service.Spec.Name, err)
	}

	for _, warning := range response.Warnings {
		slog.Debug("Response with warnings", "warning", warning)
	}

	updatedService, _, err := c.client.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("cannot inspect service %s to check update status: %w", service.Spec.Name, err)
	}

	previous := updatedService.PreviousSpec.TaskTemplate.ContainerSpec.Image
	current := updatedService.Spec.TaskTemplate.ContainerSpec.Image

	if previous != current {
		slog.Info("Updated service", "service", service.Spec.Name, "image", current)
	} else {
		slog.Debug("Service is already up to date", "service", service.Spec.Name)
	}

	return nil
}

// UpdateServices updates all the services from a Docker swarm that matches the specified image name.
// If no images are passed then it updates all the services.
func (c *Swarm) UpdateServices(ctx context.Context, imageName ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	services, err := c.serviceList(ctx)
	if err != nil {
		return fmt.Errorf("failed to get service list: %w", err)
	}

	var serviceID string

	sem := make(chan struct{}, c.MaxThreads)
	var wg sync.WaitGroup

	for _, service := range services {
		if c.validService(service) {
			sem <- struct{}{}
			wg.Add(1)

			go func(service swarm.Service) {
				defer wg.Done()
				defer func() { <-sem }()

				// try to identify this service
				if _, ok := service.Spec.Labels[serviceLabel]; ok {
					serviceID = service.ID
					return
				}

				if len(imageName) > 0 {
					hasMatch := false
					for _, imageMatch := range imageName {
						if strings.HasPrefix(service.Spec.TaskTemplate.ContainerSpec.Image, imageMatch) {
							hasMatch = true
							break
						}
					}

					if !hasMatch {
						return
					}
				}

				if err = c.updateServiceWithRetries(ctx, service); err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						slog.Error("Service update canceled", "service", service.Spec.Name)
						return
					}
					slog.Error("Cannot update service", "service", service.Spec.Name, "error", err)
				}
			}(service)
		} else {
			slog.Debug("Service was ignored by blacklist or missing label", "service", service.Spec.Name)
		}
	}

	if serviceID != "" {
		// refresh service
		service, _, err := c.client.ServiceInspectWithRaw(ctx, serviceID, types.ServiceInspectOptions{})
		if err != nil {
			return fmt.Errorf("cannot inspect the service %s: %w", serviceID, err)
		}

		err = c.updateServiceWithRetries(ctx, service)
		if err != nil {
			return fmt.Errorf("failed to update the service %s: %w", serviceID, err)
		}
	}

	return nil
}

func (c *Swarm) getImageDigest(ctx context.Context, image, encodedAuth string) (string, error) {
	namedRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", fmt.Errorf("failed to parse image name: %w", err)
	}

	if _, isCanonical := namedRef.(reference.Canonical); isCanonical {
		return "", errors.New("the image name already have a digest")
	}

	distributionInspect, err := c.client.DistributionInspect(ctx, image, encodedAuth)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image: %w", err)
	}

	// ensure that image gets a default tag if none is provided
	img, err := reference.WithDigest(namedRef, distributionInspect.Descriptor.Digest)
	if err != nil {
		return "", fmt.Errorf("the image name has an invalid format: %w", err)
	}

	return reference.FamiliarString(img), nil
}
