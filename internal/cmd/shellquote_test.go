package cmd

import (
	"maps"
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestShellQuote(t *testing.T) {
	tcs := map[string]string{
		``:            `''`,
		`'`:           `\'`,
		`''`:          `\'\'`,
		`'a'`:         `\''a'\'`,
		`\`:           `'\\'`,
		`\a`:          `'\\a'`,
		`$a`:          `'$a'`,
		`a`:           `a`,
		`a/b`:         `a/b`,
		`a b`:         `'a b'`,
		`--arg`:       `--arg`,
		`--arg=value`: `--arg=value`,
		`it's`:        `'it'\''s'`,
		`$(echo foo)`: `'$(echo foo)'`,
	}
	for _, s := range slices.Sorted(maps.Keys(tcs)) {
		assert.Equal(t, tcs[s], shellQuote(s), "quoting %q", s)
	}
}

func TestShellQuoteCommand(t *testing.T) {
	for _, tc := range []struct {
		command  string
		args     []string
		expected string
	}{
		{
			command:  "command",
			expected: "command",
		},
		{
			command:  "command with spaces",
			expected: "'command with spaces'",
		},
		{
			command:  "command",
			args:     []string{"arg1"},
			expected: "command arg1",
		},
		{
			command:  "command",
			args:     []string{"arg1", "arg 2 with spaces"},
			expected: "command arg1 'arg 2 with spaces'",
		},
	} {
		assert.Equal(t, tc.expected, shellQuoteCommand(tc.command, tc.args))
	}
}
