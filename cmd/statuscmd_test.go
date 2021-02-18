package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v2"
	"github.com/twpayne/go-vfs/v2/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestStatusCmd(t *testing.T) {
	for _, tc := range []struct {
		name           string
		root           interface{}
		args           []string
		postApplyTests []interface{}
		stdoutStr      string
	}{
		{
			name: "add_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_bashrc": "# contents of .bashrc\n",
			},
			args: []string{"~/.bashrc"},
			stdoutStr: chezmoitest.JoinLines(
				` A .bashrc`,
			),
			postApplyTests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_bashrc",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .bashrc\n"),
				),
			},
		},
		{
			name: "update_symlink",
			root: map[string]interface{}{
				"/home/user/.symlink": &vfst.Symlink{Target: "old-target"},
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": "new-target\n",
			},
			args: []string{"~/.symlink"},
			postApplyTests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestSymlinkTarget("new-target"),
				),
			},
			stdoutStr: chezmoitest.JoinLines(
				` M .symlink`,
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				var stdout strings.Builder
				require.NoError(t, newTestConfig(t, fs, withStdout(&stdout)).execute(append([]string{"status"}, tc.args...)))
				assert.Equal(t, tc.stdoutStr, stdout.String())

				require.NoError(t, newTestConfig(t, fs).execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fs, "", tc.postApplyTests...)

				stdout.Reset()
				require.NoError(t, newTestConfig(t, fs, withStdout(&stdout)).execute(append([]string{"status"}, tc.args...)))
				assert.Empty(t, stdout.String())
			})
		})
	}
}
