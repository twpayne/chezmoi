package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSecretFunc(t *testing.T) {
	t.Parallel()

	c, args := getSecretTestConfig()

	var value interface{}
	assert.NotPanics(t, func() {
		value = c.secretFunc(args...)
	})
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, value, c.secretFunc(args...))
}

func TestSecretJSONFunc(t *testing.T) {
	t.Parallel()

	c, args := getSecretJSONTestConfig()

	var value interface{}
	assert.NotPanics(t, func() {
		value = c.secretJSONFunc(args...)
	})
	time.Sleep(1100 * time.Millisecond)
	assert.Equal(t, value, c.secretJSONFunc(args...))
}
