package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-vfs/v4/vfst"

	"github.com/twpayne/chezmoi/v2/internal/archive"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestImportCmd(t *testing.T) {
	data, err := archive.NewTar(map[string]interface{}{
		"archive": map[string]interface{}{
			".dir": map[string]interface{}{
				".file": "# contents of archive/.dir/.file\n",
				".symlink": &archive.Symlink{
					Target: ".file",
				},
			},
		},
	})
	require.NoError(t, err)

	for _, tc := range []struct {
		args      []string
		extraRoot interface{}
		tests     []interface{}
	}{
		{
			args: []string{
				"--strip-components=1",
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular,
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
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dir": &vfst.Dir{Perm: 0o777},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular,
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
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dir/file": "# contents of dir/file\n",
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/file",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular,
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
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dir": &vfst.Dir{Perm: 0o777},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of archive/.dir/.file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dir/exact_dot_dir/symlink_dot_symlink",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".file\n"),
				),
			},
		},
	} {
		t.Run(strings.Join(tc.args, "_"), func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]interface{}{
				"/home/user": &vfst.Dir{Perm: 0o777},
			}, func(fileSystem vfs.FS) {
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				config := newTestConfig(t, fileSystem, withStdin(bytes.NewReader(data)))
				require.NoError(t, config.execute(append([]string{"import"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.tests...)
			})
		})
	}
}
