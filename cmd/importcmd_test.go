package cmd

import (
	"archive/tar"
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v2"
	"github.com/twpayne/go-vfs/v2/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestImportCmd(t *testing.T) {
	b := &bytes.Buffer{}
	w := tar.NewWriter(b)
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "archive/",
		Mode:     0o777,
	}))
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeDir,
		Name:     "archive/.dir/",
		Mode:     0o777,
	}))
	data := []byte("# contents of archive/.dir/.file\n")
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeReg,
		Name:     "archive/.dir/.file",
		Size:     int64(len(data)),
		Mode:     0o666,
	}))
	_, err := w.Write(data)
	assert.NoError(t, err)
	linkname := ".file"
	assert.NoError(t, w.WriteHeader(&tar.Header{
		Typeflag: tar.TypeSymlink,
		Name:     "archive/.dir/.symlink",
		Linkname: linkname,
	}))
	require.NoError(t, w.Close())

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
			}, func(fs vfs.FS) {
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(fs, tc.extraRoot))
				}
				c := newTestConfig(t, fs, withStdin(bytes.NewReader(b.Bytes())))
				require.NoError(t, c.execute(append([]string{"import"}, tc.args...)))
				vfst.RunTests(t, fs, "", tc.tests...)
			})
		})
	}
}
