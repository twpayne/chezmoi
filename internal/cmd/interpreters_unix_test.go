//go:build !windows

package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func TestNewDefaultInterpreters_PS1(t *testing.T) {
	for _, tc := range []struct {
		name           string
		findExecutable func([]string, []string) (string, error)
		expected       chezmoi.Interpreter
	}{
		{
			name: "pwsh_available",
			findExecutable: func(names, paths []string) (string, error) {
				if names[0] == "pwsh" {
					return "/usr/bin/pwsh", nil
				}
				return "", nil
			},
			expected: chezmoi.Interpreter{
				Command: "pwsh",
				Args:    []string{"-NoLogo", "-File"},
			},
		},
		{
			name: "pwsh_not_available",
			findExecutable: func(names, paths []string) (string, error) {
				return "", nil
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interpreters := NewDefaultInterpreters(tc.findExecutable)
			assert.Equal(t, tc.expected, interpreters["ps1"])
		})
	}
}
