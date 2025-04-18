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
	"os"
	"testing"

	test "github.com/stretchr/testify/assert"
)

func TestSetEnv(t *testing.T) {
	assert := test.New(t)

	err := os.Unsetenv("FOOBAR")
	assert.NoError(err)
	err = setEnvOptStr("FOOBAR", "test")
	assert.NoError(err)
	assert.Equal("test", os.Getenv("FOOBAR"))

	err = os.Unsetenv("FOOBAR")
	assert.NoError(err)
	err = setEnvOptStr("FOOBAR", "")
	assert.NoError(err)
	assert.Equal("", os.Getenv("FOOBAR"))

	err = os.Setenv("FOOBAR", "foobar")
	assert.NoError(err)
	err = setEnvOptStr("FOOBAR", "test2")
	assert.NoError(err)
	assert.Equal("test2", os.Getenv("FOOBAR"))

	err = os.Setenv("FOOBAR", "foobar")
	assert.NoError(err)
	err = setEnvOptStr("FOOBAR", "")
	assert.NoError(err)
	assert.Equal("foobar", os.Getenv("FOOBAR"))
}

func TestSetBool(t *testing.T) {
	assert := test.New(t)

	err := os.Unsetenv("FOOBAR")
	assert.NoError(err)
	err = setEnvOptBool("FOOBAR", true)
	assert.NoError(err)
	assert.Equal("1", os.Getenv("FOOBAR"))

	err = os.Unsetenv("FOOBAR")
	assert.NoError(err)
	err = setEnvOptBool("FOOBAR", false)
	assert.NoError(err)
	assert.Equal("", os.Getenv("FOOBAR"))

	err = os.Setenv("FOOBAR", "foobar")
	assert.NoError(err)
	err = setEnvOptBool("FOOBAR", true)
	assert.NoError(err)
	assert.Equal("1", os.Getenv("FOOBAR"))

	err = os.Setenv("FOOBAR", "foobar")
	assert.NoError(err)
	err = setEnvOptBool("FOOBAR", false)
	assert.NoError(err)
	assert.Equal("foobar", os.Getenv("FOOBAR"))
}
