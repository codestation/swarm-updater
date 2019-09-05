package main

import (
	"context"
	"regexp"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	test "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"megpoid.xyz/go/swarm-updater/log"
)

func TestValidServiceLabel(t *testing.T) {
	assert := test.New(t)

	s := Swarm{LabelEnable: true}
	service := swarm.Service{}

	ok := s.validService(service)
	assert.False(ok)

	service.Spec.Labels = map[string]string{enabledServiceLabel: "false"}
	ok = s.validService(service)
	assert.False(ok)

	service.Spec.Labels = map[string]string{enabledServiceLabel: "true"}
	ok = s.validService(service)
	assert.True(ok)
}

func TestValidServiceBlacklist(t *testing.T) {
	assert := test.New(t)

	s := Swarm{LabelEnable: false}
	service := swarm.Service{}
	service.Spec.Name = "service_foobar"

	ok := s.validService(service)
	assert.True(ok)

	s.Blacklist = []*regexp.Regexp{regexp.MustCompile("service_foobar")}
	ok = s.validService(service)
	assert.False(ok)

	s.Blacklist = []*regexp.Regexp{regexp.MustCompile("service_barfoo")}
	ok = s.validService(service)
	assert.True(ok)

	s.Blacklist = []*regexp.Regexp{
		regexp.MustCompile("service_barfoo1"),
		regexp.MustCompile("service_foobar"),
		regexp.MustCompile("service_barfoo2"),
	}
	ok = s.validService(service)
	assert.False(ok)

	s.Blacklist = []*regexp.Regexp{regexp.MustCompile("")}
	ok = s.validService(service)
	assert.False(ok)
}

func TestUpdateServiceEmpty(t *testing.T) {
	assert := test.New(t)

	serviceMock := &ServiceAPIClientMock{}
	s := Swarm{serviceClient: serviceMock}

	var serviceList []swarm.Service
	serviceMock.On("ServiceList",
		mock.Anything,
		mock.Anything,
	).Return(serviceList, nil)

	err := s.UpdateServices(context.TODO())
	assert.NoError(err)

	serviceMock.AssertExpectations(t)
}

func TestUpdateServices(t *testing.T) {
	assert := test.New(t)

	const currentDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	const updatedDigest = "sha256:0000000000000000000000000000000000000000000000000000000000000001"

	services := []swarm.Service{
		{
			ID: "1",
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{Name: "service_foo"},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Image: "foo:latest@" + currentDigest},
				},
			},
			PreviousSpec: &swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
			},
		},
		{
			ID: "2",
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{Name: "service_bar"},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Image: "bar:latest@" + currentDigest},
				},
			},
			PreviousSpec: &swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
			},
		},
		{
			ID: "3",
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{Name: "service_baz"},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Image: "baz:latest@" + currentDigest},
				},
			},
			PreviousSpec: &swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
			},
		},
	}

	cliMock := &CliMock{}
	serviceMock := &ServiceAPIClientMock{}
	distributionMock := &DistributionAPIClientMock{}

	s := Swarm{serviceClient: serviceMock, distributionClient: distributionMock, cli: cliMock}
	s.cli = cliMock

	ctx := context.TODO()
	tokenAuth := "token_auth"

	serviceMock.On("ServiceList",
		ctx, mock.AnythingOfType("ServiceListOptions"),
	).Return(services, nil)

	cliMock.On("RetrieveAuthTokenFromImage",
		ctx,
		mock.AnythingOfType("string"),
	).Return(tokenAuth, nil)

	distributionMock.On("DistributionInspect",
		ctx,
		mock.AnythingOfType("string"),
		tokenAuth,
	).Return(registry.DistributionInspect{
		Descriptor: v1.Descriptor{
			Digest: updatedDigest,
		},
	}, nil)
	serviceMock.On("ServiceInspectWithRaw",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("ServiceInspectOptions"),
	).Return(services[0], []byte{}, nil)
	serviceMock.On("ServiceUpdate",
		ctx,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("Version"),
		mock.AnythingOfType("ServiceSpec"),
		mock.AnythingOfType("ServiceUpdateOptions"),
	).Return(types.ServiceUpdateResponse{}, nil)

	// disable log
	log.Printf = log.Debug

	err := s.UpdateServices(ctx)
	assert.NoError(err)

	serviceMock.AssertExpectations(t)
	cliMock.AssertExpectations(t)
}
