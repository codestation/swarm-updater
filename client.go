package main

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// DockerClient interacts with a Docker Swarm
type DockerClient interface {
	RetrieveAuthTokenFromImage(image string) (string, error)
	ServiceUpdate(serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error)
	ServiceInspectWithRaw(serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error)
	ServiceList(options types.ServiceListOptions) ([]swarm.Service, error)
}

type dockerClient struct {
	apiClient *client.Client
	dockerCli *command.DockerCli
}

func (c *dockerClient) RetrieveAuthTokenFromImage(image string) (string, error) {
	return command.RetrieveAuthTokenFromImage(context.Background(), c.dockerCli, image)
}

func (c *dockerClient) ServiceUpdate(serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error) {
	return c.apiClient.ServiceUpdate(context.Background(), serviceID, version, service, options)
}

func (c *dockerClient) ServiceInspectWithRaw(serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	return c.apiClient.ServiceInspectWithRaw(context.Background(), serviceID, opts)
}

func (c *dockerClient) ServiceList(options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.apiClient.ServiceList(context.Background(), options)
}
