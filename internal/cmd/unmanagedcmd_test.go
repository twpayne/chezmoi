package cmd

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestUnmanagedCmd(t *testing.T) {
	for _, tc := range []struct {
		name           string
		root           any
		postFunc       func(vfs.FS) error
		args           []string
		expectedOutput string
	}{
		{
			name: "simple",
			root: map[string]any{
				"/home/user": map[string]any{
					".file":                         "",
					".local/share/chezmoi/dot_file": "# contents of .file\n",
					".unmanaged":                    "",
				},
			},
			expectedOutput: chezmoitest.JoinLines(
				".local",
				".unmanaged",
			),
		},
		{
			name: "private_subdir",
			root: map[string]any{
				"/home/user": map[string]any{
					".dir": map[string]any{
						"subdir/file": "",
					},
					".local/share/chezmoi/dot_dir/subdir/file": "",
				},
			},
			postFunc: func(fileSystem vfs.FS) error {
				return fileSystem.Chmod("/home/user/.dir", 0)
			},
			args: []string{"--keep-going"},
			expectedOutput: chezmoitest.JoinLines(
				".local",
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				if tc.postFunc != nil {
					assert.NoError(t, tc.postFunc(fileSystem))
				}
				stdout := &bytes.Buffer{}
				config := newTestConfig(t, fileSystem, withStdout(stdout))
				assert.NoError(t, config.execute(append([]string{"unmanaged"}, tc.args...)))
				assert.Equal(t, tc.expectedOutput, stdout.String())
			})
		})
	}
}
