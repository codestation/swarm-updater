/*
Copyright 2018 codestation

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
	"regexp"
	"strings"
	"sync"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"megpoid.dev/go/swarm-updater/log"
)

const serviceLabel string = "xyz.megpoid.swarm-updater"
const updateOnlyLabel string = "xyz.megpoid.swarm-updater.update-only"
const enabledServiceLabel string = "xyz.megpoid.swarm-updater.enable"

// Swarm struct to handle all the service operations
type Swarm struct {
	client      DockerClient
	Blacklist   []*regexp.Regexp
	LabelEnable bool
	MaxThreads  int
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
func NewSwarm() (*Swarm, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %w", err)
	}

	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker cli: %w", err)
	}

	if err = dockerCli.Initialize(flags.NewClientOptions()); err != nil {
		return nil, fmt.Errorf("failed to initialize docker cli: %w", err)
	}

	return &Swarm{client: &dockerClient{apiClient: cli, dockerCli: dockerCli}}, nil
}

func (c *Swarm) serviceList(ctx context.Context) ([]swarm.Service, error) {
	services, err := c.client.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("ServiceList failed: %w", err)
	}

	return services, nil
}

func (c *Swarm) updateService(ctx context.Context, service swarm.Service) error {
	log.Printf("updating service %s", service.Spec.Name)

	image := service.Spec.TaskTemplate.ContainerSpec.Image
	updateOpts := types.ServiceUpdateOptions{}

	// get docker auth
	encodedAuth, err := c.client.RetrieveAuthTokenFromImage(ctx, image)
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
		log.Debug("Service %s is already up to date", service.Spec.Name)

		return nil
	}

	if strings.ToLower(service.Spec.Labels[updateOnlyLabel]) == "true" {
		if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
			*service.Spec.Mode.Replicated.Replicas = 0
		}
	}

	log.Debug("Updating service %s...", service.Spec.Name)
	response, err := c.client.ServiceUpdate(ctx, service.ID, service.Version, service.Spec, updateOpts)
	if err != nil {
		return fmt.Errorf("failed to update service %s: %w", service.Spec.Name, err)
	}

	for _, warning := range response.Warnings {
		log.Debug("response warning:\n%s", warning)
	}

	updatedService, _, err := c.client.ServiceInspectWithRaw(ctx, service.ID, types.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("cannot inspect service %s to check update status: %w", service.Spec.Name, err)
	}

	previous := updatedService.PreviousSpec.TaskTemplate.ContainerSpec.Image
	current := updatedService.Spec.TaskTemplate.ContainerSpec.Image

	if previous != current {
		log.Printf("Service %s updated to %s", service.Spec.Name, current)
	} else {
		log.Debug("Service %s is up to date", service.Spec.Name)
	}

	return nil
}

// UpdateServices updates all the services from a Docker swarm that matches the specified image name.
// If no images are passed then it updates all the services.
func (c *Swarm) UpdateServices(ctx context.Context, imageName ...string) error {
	services, err := c.serviceList(ctx)
	if err != nil {
		return fmt.Errorf("failed to get service list: %w", err)
	}

	wg := sync.WaitGroup{}
	serviceLock := sync.Mutex{}

	// create a temporary copy for the queue
	serviceQueue := append([]swarm.Service{}, services...)
	log.Printf("found %d services", len(serviceQueue))

	// only create as many as we need
	threads := c.MaxThreads
	if len(serviceQueue) < threads {
		threads = len(serviceQueue)
	}

	log.Printf("starting %d threads", threads)

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(key int) {
			defer wg.Done()
			log.Printf("starting thread %d", key)

			for {
				if len(serviceQueue) == 0 {
					log.Printf("exiting thread %d", key)
					return
				}

				serviceLock.Lock()
				service, newQueue := serviceQueue[0], serviceQueue[1:]
				serviceQueue = newQueue
				serviceLock.Unlock()

				if !c.validService(service) {
					log.Debug("Service %s was ignored by blacklist or missing label", service.Spec.Name)
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

				if err = c.updateService(ctx, service); err != nil {
					if ctx.Err() == context.Canceled {
						log.Printf("Service update canceled")
						return
					}

					log.Printf("Cannot update service %s: %s", service.Spec.Name, err.Error())
				}
			}
		}(i)
	}

	wg.Wait()

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
