package cmd

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/v2/internal/archivetest"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestImportCmd(t *testing.T) {
	data, err := archivetest.NewTar(map[string]any{
		"archive": map[string]any{
			".dir": map[string]any{
				".file": "# contents of archive/.dir/.file\n",
				".symlink": &archivetest.Symlink{
					Target: ".file",
				},
			},
		},
	})
	assert.NoError(t, err)

	for _, tc := range []struct {
		args      []string
		extraRoot any
		tests     []any
	}{
		{
			args: []string{
				"--strip-components=1",
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/dot_file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".file\n"),
				),
			},
		},
		{
			args: []string{
				"--destination=~/dir",
				"--strip-components=1",
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dir": &vfst.Dir{Perm: fs.ModePerm},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/dot_file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".file\n"),
				),
			},
		},
		{
			args: []string{
				"--destination=~/dir",
				"--remove-destination",
				"--strip-components=1",
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dir/file": "# contents of dir/file\n",
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/file",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/dot_file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".file\n"),
				),
			},
		},
		{
			args: []string{
				"--destination=~/dir",
				"--exact",
				"--strip-components=1",
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dir": &vfst.Dir{Perm: fs.ModePerm},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir/dot_file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".file\n"),
				),
			},
		},
	} {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user": &vfst.Dir{Perm: fs.ModePerm},
			}, func(fileSystem vfs.FS) {
				if tc.extraRoot != nil {
					assert.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				config := newTestConfig(t, fileSystem, withStdin(bytes.NewReader(data)))
				assert.NoError(t, config.execute(append([]string{"import"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.tests...)
			})
		})
	}
}
