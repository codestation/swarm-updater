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
	"fmt"
	"log/slog"
	"regexp"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/api/types/swarm"
	test "github.com/stretchr/testify/assert"
)

type dockerClientMock struct {
	DistributionInspectFn        func(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error)
	RetrieveAuthTokenFromImageFn func(image string) (string, error)
	ServiceUpdateFn              func(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error)
	ServiceInspectWithRawFn      func(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error)
	ServiceListFn                func(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error)
}

func (s *dockerClientMock) DistributionInspect(ctx context.Context, image, encodedAuth string) (registry.DistributionInspect, error) {
	if s.DistributionInspectFn != nil {
		return s.DistributionInspectFn(ctx, image, encodedAuth)
	}

	return registry.DistributionInspect{}, nil
}

func (s *dockerClientMock) RetrieveAuthTokenFromImage(image string) (string, error) {
	if s.RetrieveAuthTokenFromImageFn != nil {
		return s.RetrieveAuthTokenFromImageFn(image)
	}

	return "", nil
}

func (s *dockerClientMock) ServiceUpdate(ctx context.Context, serviceID string, version swarm.Version, service swarm.ServiceSpec, options types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error) {
	if s.ServiceUpdateFn != nil {
		return s.ServiceUpdateFn(ctx, serviceID, version, service, options)
	}

	return swarm.ServiceUpdateResponse{}, nil
}

func (s *dockerClientMock) ServiceInspectWithRaw(ctx context.Context, serviceID string, opts types.ServiceInspectOptions) (swarm.Service, []byte, error) {
	if s.ServiceInspectWithRawFn != nil {
		return s.ServiceInspectWithRawFn(ctx, serviceID, opts)
	}

	return swarm.Service{}, nil, nil
}

func (s *dockerClientMock) ServiceList(ctx context.Context, options types.ServiceListOptions) ([]swarm.Service, error) {
	if s.ServiceListFn != nil {
		return s.ServiceListFn(ctx, options)
	}

	return []swarm.Service{}, nil
}

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

	mock := dockerClientMock{}
	mock.ServiceListFn = func(_ context.Context, _ types.ServiceListOptions) ([]swarm.Service, error) {
		return []swarm.Service{}, nil
	}

	s := Swarm{client: &mock}
	err := s.UpdateServices(context.TODO())
	assert.NoError(err)
}

func TestUpdateServices(t *testing.T) {
	assert := test.New(t)

	services := []swarm.Service{
		{
			ID: "1",
			Spec: swarm.ServiceSpec{
				Annotations: swarm.Annotations{Name: "service_foo"},
				TaskTemplate: swarm.TaskSpec{
					ContainerSpec: &swarm.ContainerSpec{Image: "foo:latest@sha256:0000000000000000000000000000000000000000000000000000000000000000"},
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
					ContainerSpec: &swarm.ContainerSpec{Image: "bar:latest@sha256:0000000000000000000000000000000000000000000000000000000000000000"},
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
					ContainerSpec: &swarm.ContainerSpec{Image: "baz:latest@sha256:0000000000000000000000000000000000000000000000000000000000000000"},
				},
			},
			PreviousSpec: &swarm.ServiceSpec{
				TaskTemplate: swarm.TaskSpec{ContainerSpec: &swarm.ContainerSpec{}},
			},
		},
	}

	mock := dockerClientMock{}

	mock.ServiceListFn = func(_ context.Context, _ types.ServiceListOptions) ([]swarm.Service, error) {
		return services, nil
	}

	mock.ServiceInspectWithRawFn = func(_ context.Context, serviceID string, _ types.ServiceInspectOptions) (swarm.Service, []byte, error) {
		for _, service := range services {
			if service.ID == serviceID {
				return service, nil, nil
			}
		}

		assert.Fail("Should be on the service list", "%s isn't on service list", serviceID)

		return swarm.Service{}, nil, fmt.Errorf("service not found: %s", serviceID)
	}

	mock.ServiceUpdateFn = func(_ context.Context, serviceID string, _ swarm.Version, service swarm.ServiceSpec, _ types.ServiceUpdateOptions) (swarm.ServiceUpdateResponse, error) {
		for _, serv := range services {
			if serv.ID == serviceID {
				image := service.TaskTemplate.ContainerSpec.Image
				regex := regexp.MustCompile(".*@sha256:.*")
				matched := regex.MatchString(image)
				assert.False(matched, "%s doesn't has the hash stripped", image)

				serv.PreviousSpec.TaskTemplate.ContainerSpec.Image = image
				serv.Spec.TaskTemplate.ContainerSpec.Image = image + "@sha256:1111111111111111111111111111111111111111111111111111111111111111"

				return swarm.ServiceUpdateResponse{}, nil
			}
		}

		assert.Fail("Should be on the service list", "%s isn't on service list", serviceID)

		return swarm.ServiceUpdateResponse{}, fmt.Errorf("service not found: %s", serviceID)
	}

	// disable slog output
	slog.SetDefault(slog.New(slog.DiscardHandler))

	s := Swarm{client: &mock, MaxThreads: 1}
	err := s.UpdateServices(context.TODO())
	assert.NoError(err)
}
