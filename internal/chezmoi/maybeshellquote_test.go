package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaybeShellQuote(t *testing.T) {
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
		assert.Equal(t, expected, maybeShellQuote(s), "quoting %q", s)
	}
}

func TestShellQuoteArgs(t *testing.T) {
	for _, tc := range []struct {
		args     []string
		expected string
	}{
		{
			args:     []string{},
			expected: "",
		},
		{
			args:     []string{"foo"},
			expected: "foo",
		},
		{
			args:     []string{"foo", "bar baz"},
			expected: "foo 'bar baz'",
		},
	} {
		assert.Equal(t, tc.expected, ShellQuoteArgs(tc.args))
	}
}
