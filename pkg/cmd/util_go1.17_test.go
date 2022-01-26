//go:build go1.17
// +build go1.17

package cmd

import "testing"

func testSetenv(t *testing.T, key, value string) {
	t.Helper()
	t.Setenv(key, value)
}
