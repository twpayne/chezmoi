package cmd

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v3"
	"github.com/twpayne/go-vfs/v3/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestSymlinks(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extraRoot interface{}
		args      []string
		tests     []interface{}
	}{
		{
			name: "symlink_forward_slash_unix",
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/file",
			},
			args: []string{"~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir/file"),
				),
			},
		},
		{
			name: "symlink_forward_slash_windows",
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/file",
			},
			args: []string{"~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir\\file"),
				),
			},
		},
		{
			name: "symlink_backward_slash_windows",
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir\\file",
			},
			args: []string{"~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir\\file"),
				),
			},
		},
		{
			name: "symlink_mixed_slash_windows",
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/subdir\\file",
			},
			args: []string{"~/.symlink"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir\\subdir\\file"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)
			chezmoitest.WithTestFS(t, nil, func(fileSystem vfs.FS) {
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				require.NoError(t, newTestConfig(t, fileSystem).execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.tests)
				require.NoError(t, newTestConfig(t, fileSystem).execute([]string{"verify"}))
			})
		})
	}
}
