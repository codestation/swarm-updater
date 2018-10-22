package main

import (
	"os"
	"testing"

	test "github.com/stretchr/testify/assert"
)

func TestSetEnv(t *testing.T) {
	assert := test.New(t)

	os.Unsetenv("FOOBAR")
	err := setEnvOptStr("FOOBAR", "test")
	assert.NoError(err)
	assert.Equal("test", os.Getenv("FOOBAR"))

	os.Unsetenv("FOOBAR")
	err = setEnvOptStr("FOOBAR", "")
	assert.NoError(err)
	assert.Equal("", os.Getenv("FOOBAR"))

	os.Setenv("FOOBAR", "foobar")
	err = setEnvOptStr("FOOBAR", "test2")
	assert.NoError(err)
	assert.Equal("test2", os.Getenv("FOOBAR"))

	os.Setenv("FOOBAR", "foobar")
	err = setEnvOptStr("FOOBAR", "")
	assert.NoError(err)
	assert.Equal("foobar", os.Getenv("FOOBAR"))
}

func TestSetBool(t *testing.T) {
	assert := test.New(t)

	os.Unsetenv("FOOBAR")
	err := setEnvOptBool("FOOBAR", true)
	assert.NoError(err)
	assert.Equal("1", os.Getenv("FOOBAR"))

	os.Unsetenv("FOOBAR")
	err = setEnvOptBool("FOOBAR", false)
	assert.NoError(err)
	assert.Equal("", os.Getenv("FOOBAR"))

	os.Setenv("FOOBAR", "foobar")
	err = setEnvOptBool("FOOBAR", true)
	assert.NoError(err)
	assert.Equal("1", os.Getenv("FOOBAR"))

	os.Setenv("FOOBAR", "foobar")
	err = setEnvOptBool("FOOBAR", false)
	assert.NoError(err)
	assert.Equal("foobar", os.Getenv("FOOBAR"))
}
