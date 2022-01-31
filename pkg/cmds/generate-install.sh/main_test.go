package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoToolDistList(t *testing.T) {
	_, err := goToolDistList()
	require.NoError(t, err)
}
