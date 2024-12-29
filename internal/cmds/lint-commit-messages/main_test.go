package main

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestCommitRx(t *testing.T) {
	prefix := strings.Repeat("0", 40) + " "
	for s, match := range map[string]bool{
		"chore(deps): Text":     true,
		"chore(deps-dev): Text": true,
		"chore: Text":           true,
		"docs: Text":            true,
		"feat: Text":            true,
		"fix: Text":             true,
		"fixup!":                false,
		"snapshot":              false,
	} {
		assert.Equal(t, match, commitRx.MatchString(prefix+s))
	}
}
