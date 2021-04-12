package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v2"
	"github.com/twpayne/go-vfs/v2/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestAddCmd(t *testing.T) {
	for _, tc := range []struct {
		name  string
		root  interface{}
		args  []string
		tests []interface{}
	}{
		{
			name: "dir",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": &vfst.Dir{Perm: 0o777},
				},
			},
			args: []string{"~/.dir"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
			},
		},
		{
			name: "dir_with_file",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": &vfst.Dir{
						Perm: 0o777,
						Entries: map[string]interface{}{
							"file": "# contents of .dir/file\n",
						},
					},
				},
			},
			args: []string{"~/.dir"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
			},
		},
		{
			name: "dir_with_file_with_--recursive=false",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": &vfst.Dir{
						Perm: 0o777,
						Entries: map[string]interface{}{
							"file": "# contents of .dir/file\n",
						},
					},
				},
			},
			args: []string{"~/.dir", "--recursive=false"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dir_private_unix",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": &vfst.Dir{
						Perm: 0o700,
						Entries: map[string]interface{}{
							"file": "# contents of .dir/file\n",
						},
					},
				},
			},
			args: []string{"~/.dir"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
			},
		},
		{
			name: "dir_file_private_unix",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": &vfst.Dir{
						Perm: 0o700,
						Entries: map[string]interface{}{
							"file": "# contents of .dir/file\n",
						},
					},
				},
			},
			args: []string{"~/.dir/file"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
			},
		},
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".empty": "",
				},
			},
			args: []string{"~/.empty"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/empty_dot_empty",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "empty_with_--empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".empty": "",
				},
			},
			args: []string{"--empty", "~/.empty"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/empty_dot_empty",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContents(nil),
				),
			},
		},
		{
			name: "executable_unix",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".executable": &vfst.File{
						Perm:     0o777,
						Contents: []byte("#!/bin/sh\n"),
					},
				},
			},
			args: []string{"~/.executable"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/executable_dot_executable",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("#!/bin/sh\n"),
				),
			},
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".file": "# contents of .file\n",
				},
			},
			args: []string{"~/.file"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".symlink": &vfst.Symlink{
						Target: ".dir/subdir/file",
					},
				},
			},
			args: []string{"~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/symlink_dot_symlink",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString(".dir/subdir/file\n"),
				),
			},
		},
		{
			name: "symlink_with_--follow",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".file": "# contents of .file\n",
					".symlink": &vfst.Symlink{
						Target: ".file",
					},
				},
			},
			args: []string{"--follow", "~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_symlink",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				require.NoError(t, newTestConfig(t, fs).execute(append([]string{"add"}, tc.args...)))
				vfst.RunTests(t, fs, "", tc.tests...)
			})
		})
	}
}
