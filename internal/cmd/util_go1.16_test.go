//go:build !go1.17
// +build !go1.17

package cmd

import (
	"os"
	"testing"
)

func testSetenv(t *testing.T, key, value string) {
	t.Helper()
	prevValue := os.Getenv(key)
	t.Cleanup(func() {
		os.Setenv(key, prevValue)
	})
	os.Setenv(key, value)
}
