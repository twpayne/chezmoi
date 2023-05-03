package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestGoToolDistList(t *testing.T) {
	_, err := goToolDistList()
	assert.NoError(t, err)
}
