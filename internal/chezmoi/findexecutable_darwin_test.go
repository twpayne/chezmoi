package chezmoi

import (
	"fmt"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFindExecutable(t *testing.T) {
	tests := []struct {
		expected string
		files    []string
		paths    []string
	}{
		{
			files:    []string{"sh"},
			paths:    []string{"/usr/bin", "/bin"},
			expected: "/bin/sh",
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
			files:    []string{"chezmoish", "sh"},
			paths:    []string{"/usr/bin", "/bin"},
			expected: "/bin/sh",
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
		format := "FindExecutable %#v in %#v as %#v"
		name := fmt.Sprintf(format, test.files, test.paths, test.expected)
		t.Run(name, func(t *testing.T) {
			actual, err := FindExecutable(test.files, test.paths)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
