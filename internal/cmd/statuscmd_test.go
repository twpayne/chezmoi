package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestStatusCmd(t *testing.T) {
	for _, tc := range []struct {
		name           string
		root           any
		args           []string
		postApplyTests []any
		stdoutStr      string
	}{
		{
			name: "add_file",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_bashrc": "# contents of .bashrc\n",
			},
			args: []string{"~/.bashrc"},
			stdoutStr: chezmoitest.JoinLines(
				` A .bashrc`,
			),
			postApplyTests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_bashrc",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .bashrc\n"),
				),
			},
		},
		{
			name: "update_symlink",
			root: map[string]any{
				"/home/user/.symlink": &vfst.Symlink{Target: "old-target"},
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": "new-target\n",
			},
			args: []string{"~/.symlink"},
			postApplyTests: []any{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestSymlinkTarget("new-target"),
				),
			},
			stdoutStr: chezmoitest.JoinLines(
				` M .symlink`,
			),
		},
		{
			name: "path_style",
			root: map[string]any{
				"/home/user/.config/chezmoi/chezmoi.toml": chezmoitest.JoinLines(
					`[status]`,
					`    pathStyle = "relative"`,
				),
				"/home/user/.local/share/chezmoi": &vfst.Dir{Perm: 0o755},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				stdout := strings.Builder{}
				config1 := newTestConfig(t, fileSystem, withStdout(&stdout))
				assert.NoError(t, config1.execute(append([]string{"status"}, tc.args...)))
				assert.Equal(t, tc.stdoutStr, stdout.String())

				config2 := newTestConfig(t, fileSystem)
				assert.NoError(t, config2.execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.postApplyTests...)

				stdout.Reset()
				config3 := newTestConfig(t, fileSystem, withStdout(&stdout))
				assert.NoError(t, config3.execute(append([]string{"status"}, tc.args...)))
				assert.Zero(t, stdout.String())
			})
		})
	}
}
