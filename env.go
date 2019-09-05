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
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const dockerAPIMinVersion string = "1.30"

func setEnvOptStr(env string, opt string) error {
	if opt != "" && opt != os.Getenv(env) {
		err := os.Setenv(env, opt)
		if err != nil {
			return err
		}
	}
	return nil
}

func setEnvOptBool(env string, opt bool) error {
	if opt == true {
		return setEnvOptStr(env, "1")
	}
	return nil
}

// envConfig translates the command-line options into environment variables
// that will initialize the api serviceClient
func envConfig(c *cli.Context) error {
	var err error

	err = setEnvOptStr("DOCKER_HOST", c.String("host"))
	if err != nil {
		return errors.Wrap(err, "failed to set environment DOCKER_HOST")
	}

	err = setEnvOptStr("DOCKER_CONFIG", c.String("config"))
	if err != nil {
		return errors.Wrap(err, "failed to set environment DOCKER_CONFIG")
	}

	err = setEnvOptBool("DOCKER_TLS_VERIFY", c.Bool("tlsverify"))
	if err != nil {
		return errors.Wrap(err, "failed to set environment DOCKER_TLS_VERIFY")
	}

	err = setEnvOptStr("DOCKER_API_VERSION", dockerAPIMinVersion)
	if err != nil {
		return errors.Wrap(err, "failed to set environment DOCKER_API_VERSION")
	}

	return err
}
