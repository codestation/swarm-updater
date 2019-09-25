package main

import (
	"context"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	test "github.com/stretchr/testify/assert"
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
	ctx := context.TODO()

	serviceMock.On("ServiceList", ctx, types.ServiceListOptions{}).Return(make([]swarm.Service, 0), nil)

	err := s.UpdateServices(ctx)
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

	serviceMock.On("ServiceList", ctx, types.ServiceListOptions{}).Return(services, nil)

	for _, service := range services {
		imageDigest := service.Spec.TaskTemplate.ContainerSpec.Image
		cliMock.On("RetrieveAuthTokenFromImage", ctx, imageDigest).Return(tokenAuth, nil)

		imageName := strings.Split(imageDigest, "@")[0]
		distributionMock.On("DistributionInspect",
			ctx, imageName, tokenAuth,
		).Return(registry.DistributionInspect{Descriptor: v1.Descriptor{Digest: updatedDigest}}, nil)

		serviceMock.On("ServiceInspectWithRaw",
			ctx, service.ID, types.ServiceInspectOptions{},
		).Return(services[0], []byte{}, nil)

		serviceMock.On("ServiceUpdate",
			ctx, service.ID, service.Version, service.Spec, types.ServiceUpdateOptions{EncodedRegistryAuth: tokenAuth},
		).Return(types.ServiceUpdateResponse{}, nil)
	}

	// disable log
	logrus.SetOutput(ioutil.Discard)

	err := s.UpdateServices(ctx)
	assert.NoError(err)

	serviceMock.AssertExpectations(t)
	cliMock.AssertExpectations(t)
}
