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

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// DockerClient interacts with a Docker Swarm.
type DockerClient interface {
	DistributionInspect(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error)
	RetrieveAuthTokenFromImage(image string) (string, error)
	ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error)
	ServiceInspectWithRaw(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
}

type dockerClient struct {
	apiClient  *client.Client
	configFile *configfile.ConfigFile
}

func (c *dockerClient) DistributionInspect(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error) {
	return c.apiClient.DistributionInspect(ctx, image, encodedAuth)
}

func (c *dockerClient) RetrieveAuthTokenFromImage(image string) (string, error) {
	return command.RetrieveAuthTokenFromImage(c.configFile, image)
}

func (c *dockerClient) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error) {
	return c.apiClient.ServiceUpdate(ctx, serviceID, version, service, options)
}

func (c *dockerClient) ServiceInspectWithRaw(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	return c.apiClient.ServiceInspectWithRaw(ctx, serviceID, opts)
}

func (c *dockerClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.apiClient.ServiceList(ctx, options)
}
