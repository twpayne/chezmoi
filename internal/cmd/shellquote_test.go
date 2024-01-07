package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestShellQuote(t *testing.T) {
	for s, expected := range map[string]string{
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
	} {
		assert.Equal(t, expected, shellQuote(s), "quoting %q", s)
	}
}

func TestShellQuoteCommand(t *testing.T) {
	for _, tc := range []struct {
		command  string
		expected string
		args     []string
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
