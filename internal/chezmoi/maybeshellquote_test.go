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
		assert.Equal(t, expected, MaybeShellQuote(s), "quoting %q", s)
	}
}
