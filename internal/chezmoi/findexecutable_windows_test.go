//go:build windows

package chezmoi

import (
	"fmt"
	"strings"
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
			files: []string{"powershell.exe"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		},
		{
			files: []string{"powershell"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		},
		{
			files: []string{"weakshell.exe"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "",
		},
		{
			files: []string{"weakshell"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "",
		},
		{
			files: []string{"weakshell.exe", "powershell.exe"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		},
		{
			files: []string{"weakshell", "powershell"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0\\powershell.exe",
		},
		{
			files: []string{"weakshell.exe", "chezmoishell.exe"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "",
		},
		{
			files: []string{"weakshell", "chezmoishell"},
			paths: []string{
				"c:\\windows\\system32",
				"c:\\windows\\system64",
				"C:\\WINDOWS\\System32\\WindowsPowerShell\\v1.0",
			},
			expected: "",
		},
	}

	for _, test := range tests {
		name := fmt.Sprintf("FindExecutable %v in %#v as %v", test.files, test.paths, test.expected)
		t.Run(name, func(t *testing.T) {
			actual, err := FindExecutable(test.files, test.paths)
			assert.NoError(t, err)
			assert.Equal(t, strings.ToLower(test.expected), strings.ToLower(actual))
		})
	}
}
