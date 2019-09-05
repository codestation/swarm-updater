package main

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	"github.com/stretchr/testify/mock"
)

type CliMock struct {
	mock.Mock
}

func (c *CliMock) RetrieveAuthTokenFromImage(ctx context.Context, image string) (string, error) {
	args := c.Called(ctx, image)
	return args.String(0), args.Error(1)
}

type DistributionAPIClientMock struct {
	mock.Mock
}

func (c *DistributionAPIClientMock) DistributionInspect(ctx context.Context, image, encodedRegistryAuth string) (registry.DistributionInspect, error) {
	args := c.Called(ctx, image, encodedRegistryAuth)
	return args.Get(0).(registry.DistributionInspect), args.Error(1)
}

type ServiceAPIClientMock struct {
	mock.Mock
}

func (c *ServiceAPIClientMock) ServiceCreate(ctx context.Context, service swarm.ServiceSpec, options types.ServiceCreateOptions) (types.ServiceCreateResponse, error) {
	panic("implement me")
}

func (c *ServiceAPIClientMock) ServiceInspectWithRaw(ctx context.Context, serviceID string, options types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	args := c.Called(ctx, serviceID, options)
	return args.Get(0).(swarm.Service), args.Get(1).([]byte), args.Error(2)
}

func (c *ServiceAPIClientMock) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	args := c.Called(ctx, options)
	return args.Get(0).([]swarm.Service), args.Error(1)
}

func (c *ServiceAPIClientMock) ServiceRemove(ctx context.Context, serviceID string) error {
	panic("implement me")
}

func (c *ServiceAPIClientMock) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (types.ServiceUpdateResponse, error) {
	args := c.Called(ctx, serviceID, version, service, options)
	return args.Get(0).(types.ServiceUpdateResponse), args.Error(1)
}

func (c *ServiceAPIClientMock) ServiceLogs(ctx context.Context, serviceID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	panic("implement me")
}

func (c *ServiceAPIClientMock) TaskLogs(ctx context.Context, taskID string, options types.ContainerLogsOptions) (io.ReadCloser, error) {
	panic("implement me")
}

func (c *ServiceAPIClientMock) TaskInspectWithRaw(ctx context.Context, taskID string) (swarm.Task, []byte, error) {
	panic("implement me")
}

func (c *ServiceAPIClientMock) TaskList(ctx context.Context, options types.TaskListOptions) ([]swarm.Task, error) {
	panic("implement me")
}
