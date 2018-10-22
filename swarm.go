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
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"

	"megpoid.xyz/go/swarm-updater/log"
)

const serviceLabel string = "xyz.megpoid.swarm-updater.enable"

// Swarm struct to handle all the service operations
type Swarm struct {
	client      DockerClient
	Blacklist   []*regexp.Regexp
	LabelEnable bool
}

func (c *Swarm) validService(service swarm.Service) bool {
	if c.LabelEnable {
		label := service.Spec.Labels[serviceLabel]
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
	ctx := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize docker client: %s", err.Error())
	}

	dockerCli := command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false)
	dockerCli.Initialize(flags.NewClientOptions())

	return &Swarm{client: &dockerClient{client: cli, ctx: ctx, cli: dockerCli}}, nil
}

func (c *Swarm) serviceList() ([]swarm.Service, error) {
	services, err := c.client.ServiceList(types.ServiceListOptions{})
	if err != nil {
		return nil, fmt.Errorf("ServiceList: %s", err.Error())
	}

	return services, nil
}

func (c *Swarm) updateService(service swarm.Service) error {
	image := service.Spec.TaskTemplate.ContainerSpec.Image

	// get docker auth
	encodedAuth, err := c.client.RetrieveAuthTokenFromImage(image)
	if err != nil {
		return fmt.Errorf("cannot retrieve auth token from service's image %s", service.Spec.Name)
	}

	// remove image hash from name
	imageName := strings.Split(image, "@sha")[0]
	service.Spec.TaskTemplate.ContainerSpec.Image = imageName

	response, err := c.client.ServiceUpdate(service.ID, service.Version,
		service.Spec, types.ServiceUpdateOptions{EncodedRegistryAuth: encodedAuth, QueryRegistry: true})
	if err != nil {
		return fmt.Errorf("cannot update service %s: %s", service.Spec.Name, err.Error())
	}

	if len(response.Warnings) > 0 {
		for _, warning := range response.Warnings {
			log.Debug("response warning:\n%s", warning)
		}
	}

	updatedService, _, err := c.client.ServiceInspectWithRaw(service.ID, types.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("cannot inspect service %s to check update status: %s", service.Spec.Name, err.Error())
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

// UpdateServices updates all the services from a Docker swarm
func (c *Swarm) UpdateServices() error {
	services, err := c.serviceList()
	if err != nil {
		return fmt.Errorf("failed to get service list: %s", err.Error())
	}

	var serviceID string

	for _, service := range services {
		if c.validService(service) {
			// try to identify this service, naive approach
			namedRef, _ := reference.ParseNormalizedNamed(service.Spec.TaskTemplate.ContainerSpec.Image)
			currentImageName := reference.Path(namedRef)

			if currentImageName == ImageName {
				serviceID = service.ID
				continue
			}

			c.updateService(service)
		} else {
			log.Debug("Service %s was ignored by blacklist or missing label", service.Spec.Name)
		}
	}

	if serviceID != "" {
		// refresh service
		service, _, err := c.client.ServiceInspectWithRaw(serviceID, types.ServiceInspectOptions{})
		if err != nil {
			return fmt.Errorf("cannot inspect this service: %s", err.Error())
		}

		err = c.updateService(service)
		if err != nil {
			return fmt.Errorf("failed to update this service: %s", err.Error())
		}
	}

	return nil
}
