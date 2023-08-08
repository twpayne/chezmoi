//go:build !windows && !darwin

package chezmoi

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFindExecutable(t *testing.T) {
	tests := []struct {
		files    []string
		paths    []string
		expected string
	}{
		{
			files:    []string{"yes"},
			paths:    []string{"/usr/bin", "/bin"},
			expected: "/usr/bin/yes",
		},
		{
			files:    []string{"sh"},
			paths:    []string{"/bin", "/usr/bin"},
			expected: "/bin/sh",
		},
		{
			files:    []string{"chezmoish"},
			paths:    []string{"/bin", "/usr/bin"},
			expected: "",
		},
		{
			files:    []string{"chezmoish", "yes"},
			paths:    []string{"/usr/bin", "/bin"},
			expected: "/usr/bin/yes",
		},
		{
			files:    []string{"chezmoish", "sh"},
			paths:    []string{"/bin", "/usr/bin"},
			expected: "/bin/sh",
		},
		{
			files:    []string{"chezmoish", "chezvoush"},
			paths:    []string{"/bin", "/usr/bin"},
			expected: "",
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf(
			"FindExecutable %#v in %#v as %#v",
			test.files,
			test.paths,
			test.expected,
		)
		t.Run(name, func(t *testing.T) {
			actual, err := FindExecutable(test.files, test.paths)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
