package cmd

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestManagedCmd(t *testing.T) {
	templateContents := `{{ fail "Template should not be executed" }}`
	for _, tc := range []struct {
		name           string
		root           any
		args           []string
		expectedOutput string
	}{
		{
			name: "simple",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_file": "# contents of .file\n",
			},
			expectedOutput: chezmoitest.JoinLines(
				".file",
			),
		},
		{
			name: "template",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_template.tmpl": templateContents,
			},
			expectedOutput: chezmoitest.JoinLines(
				".template",
			),
		},
		{
			name: "create_template",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/create_dot_file.tmpl": templateContents,
			},
			expectedOutput: chezmoitest.JoinLines(
				".file",
			),
		},
		{
			name: "modify_template",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/modify_dot_file.tmpl": templateContents,
			},
			expectedOutput: chezmoitest.JoinLines(
				".file",
			),
		},
		{
			name: "remove",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi/.chezmoiremove": chezmoitest.JoinLines(
						".remove",
					),
					".remove": "",
				},
			},
			expectedOutput: chezmoitest.JoinLines(
				".remove",
			),
		},
		{
			name: "script_template",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/run_script.tmpl": templateContents,
			},
			args: []string{
				"--include", "always,scripts",
			},
			expectedOutput: chezmoitest.JoinLines(
				"script",
			),
		},
		{
			name: "symlink_template",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink.tmpl": templateContents,
			},
			expectedOutput: chezmoitest.JoinLines(
				".symlink",
			),
		},
		{
			name: "external_git_repo",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/.chezmoiexternal.toml": chezmoitest.JoinLines(
					`[".dir"]`,
					`    type = "git-repo"`,
					`    url = "https://github.com/example/example.git"`,
				),
			},
			expectedOutput: chezmoitest.JoinLines(
				".dir",
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				stdout := &bytes.Buffer{}
				config := newTestConfig(t, fileSystem, withStdout(stdout))
				assert.NoError(t, config.execute(append([]string{"managed"}, tc.args...)))
				assert.Equal(t, tc.expectedOutput, stdout.String())
			})
		})
	}
}
