package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestInstallInitShellShArgs(t *testing.T) {
	script := "#!/bin/sh\nusername=\"${LOGNAME}\"\necho don't\n"

	t.Run("unquoted", func(t *testing.T) {
		args := installInitShellShArgs(script, false)
		assert.Equal(t, []string{"sh", "-c", script}, args)
	})

	t.Run("quoted survives reparse", func(t *testing.T) {
		args := installInitShellShArgs(script, true)
		assert.Equal(t, "sh", args[0])
		assert.Equal(t, "-c", args[1])
		quoted := args[2]
		assert.NotEqual(t, script, quoted)
		// Simulate the remote shell re-parsing the concatenated command: the
		// quoted script must be a single word that, after quote removal,
		// reproduces the original script exactly.
		unquoted := unquoteTestHelper(t, quoted)
		assert.Equal(t, script, unquoted)
	})
}

// unquoteTestHelper reverses shellQuote's single-quote encoding the way a
// POSIX shell or fish does when re-parsing a command line.
func unquoteTestHelper(t *testing.T, quoted string) string {
	t.Helper()
	var result strings.Builder
	i := 0
	for i < len(quoted) {
		switch quoted[i] {
		case '\'':
			i++
			for i < len(quoted) && quoted[i] != '\'' {
				result.WriteByte(quoted[i])
				i++
			}
			i++
		case '\\':
			i++
			result.WriteByte(quoted[i])
			i++
		default:
			result.WriteByte(quoted[i])
			i++
		}
	}
	return result.String()
}
