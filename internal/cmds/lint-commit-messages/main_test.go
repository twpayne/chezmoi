package main

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCommitRx(t *testing.T) {
	prefix := strings.Repeat("0", 40) + " "
	for s, match := range map[string]bool{
		"chore(deps): text":     true,
		"chore(deps-dev): text": true,
		"chore: text":           true,
		"docs: text":            true,
		"feat: text":            true,
		"fix: text":             true,
		"fixup!":                false,
		"snapshot":              false,
	} {
		assert.Equal(t, match, commitRx.MatchString(prefix+s))
	}
}
