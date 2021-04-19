package main

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// DockerClient interacts with a Docker Swarm.
type DockerClient interface {
	DistributionInspect(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error)
	RetrieveAuthTokenFromImage(ctx context.Context, image string) (string, error)
	ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error)
	ServiceInspectWithRaw(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error)
	ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
}

type dockerClient struct {
	apiClient *client.Client
	dockerCli command.Cli
}

func (c *dockerClient) DistributionInspect(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error) {
	return c.apiClient.DistributionInspect(ctx, image, encodedAuth)
}

func (c *dockerClient) RetrieveAuthTokenFromImage(ctx context.Context, image string) (string, error) {
	return command.RetrieveAuthTokenFromImage(ctx, c.dockerCli, image)
}

func (c *dockerClient) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error) {
	return c.apiClient.ServiceUpdate(ctx, serviceID, version, service, options)
}

func (c *dockerClient) ServiceInspectWithRaw(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	return c.apiClient.ServiceInspectWithRaw(ctx, serviceID, opts)
}

func (c *dockerClient) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	return c.apiClient.ServiceList(ctx, options)
}
