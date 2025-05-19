package cmd

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestParseMackupApplication(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lines    []string
		expected mackupApplicationConfig
	}{
		{
			name: "curl.cfg",
			lines: []string{
				"[application]",
				"name = Curl",
				"",
				"[configuration_files]",
				".netrc",
				".curlrc",
			},
			expected: mackupApplicationConfig{
				Application: mackupApplicationApplicationConfig{
					Name: "Curl",
				},
				ConfigurationFiles: []chezmoi.RelPath{
					chezmoi.NewRelPath(".netrc"),
					chezmoi.NewRelPath(".curlrc"),
				},
			},
		},
		{
			name: "vscode.cfg",
			lines: []string{
				"[application]",
				"name = Visual Studio Code",
				"",
				"[configuration_files]",
				"Library/Application Support/Code/User/snippets",
				"Library/Application Support/Code/User/keybindings.json",
				"Library/Application Support/Code/User/settings.json",
				"",
				"[xdg_configuration_files]",
				"Code/User/snippets",
				"Code/User/keybindings.json",
				"Code/User/settings.json",
			},
			expected: mackupApplicationConfig{
				Application: mackupApplicationApplicationConfig{
					Name: "Visual Studio Code",
				},
				ConfigurationFiles: []chezmoi.RelPath{
					chezmoi.NewRelPath("Library/Application Support/Code/User/snippets"),
					chezmoi.NewRelPath("Library/Application Support/Code/User/keybindings.json"),
					chezmoi.NewRelPath("Library/Application Support/Code/User/settings.json"),
				},
				XDGConfigurationFiles: []chezmoi.RelPath{
					chezmoi.NewRelPath("Code/User/snippets"),
					chezmoi.NewRelPath("Code/User/keybindings.json"),
					chezmoi.NewRelPath("Code/User/settings.json"),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := parseMackupApplication([]byte(chezmoitest.JoinLines(tc.lines...)))
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
