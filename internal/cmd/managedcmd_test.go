package cmd

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/internal/chezmoitest"
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
			name: "nul_path_separator",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dot_file1": "# contents of .file1\n",
					"dot_file2": "# contents of .file2\n",
				},
			},
			args: []string{
				"-0",
			},
			expectedOutput: ".file1\x00.file2\x00",
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
