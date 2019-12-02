package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaybeShellQuote(t *testing.T) {
	for s, expected := range map[string]string{
		``:    `''`,
		`'`:   `\'`,
		`''`:  `\'\'`,
		`'a'`: `\''a'\'`,
		`\`:   `'\\'`,
		`\a`:  `'\\a'`,
		`$a`:  `'$a'`,
		`a`:   `a`,
		`a/b`: `a/b`,
		`a b`: `'a b'`,
	} {
		assert.Equal(t, expected, maybeShellQuote(s), "quoting %q", s)
	}
}
