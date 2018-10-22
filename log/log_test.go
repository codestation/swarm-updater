package log

import (
	"reflect"
	"testing"

	test "github.com/stretchr/testify/assert"
)

func TestDebugLog(t *testing.T) {
	assert := test.New(t)

	nullDebug := reflect.ValueOf(Debug)
	assert.Equal(nullDebug, reflect.ValueOf(nullLogger))
	EnableDebug(true)
	assert.NotEqual(nullDebug, reflect.ValueOf(Printf))
}
