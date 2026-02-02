package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func TestNewDefaultInterpreters_OtherInterpreters(t *testing.T) {
	// Verify that other interpreters are not affected by ps1 changes
	interpreters := NewDefaultInterpreters(func([]string, []string) (string, error) {
		return "", nil
	})

	for _, tc := range []struct {
		ext      string
		expected chezmoi.Interpreter
	}{
		{ext: "bat"},
		{ext: "cmd"},
		{ext: "com"},
		{ext: "exe"},
		{ext: "nu", expected: chezmoi.Interpreter{Command: "nu"}},
		{ext: "pl", expected: chezmoi.Interpreter{Command: "perl"}},
		{ext: "py", expected: chezmoi.Interpreter{Command: "python3"}},
		{ext: "rb", expected: chezmoi.Interpreter{Command: "ruby"}},
	} {
		t.Run(tc.ext, func(t *testing.T) {
			assert.Equal(t, tc.expected, interpreters[tc.ext])
		})
	}
}
