package cmd

import (
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestSymlinks(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extraRoot any
		args      []string
		tests     []any
	}{
		{
			name: "symlink_forward_slash_unix",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/file",
			},
			args: []string{"~/.symlink"},
			tests: []any{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir/file"),
				),
			},
		},
		{
			name: "symlink_forward_slash_windows",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/file",
			},
			args: []string{"~/.symlink"},
			tests: []any{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir\\file"),
				),
			},
		},
		{
			name: "symlink_backward_slash_windows",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir\\file",
			},
			args: []string{"~/.symlink"},
			tests: []any{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(".dir\\file"),
				),
			},
		},
		{
			name: "symlink_mixed_slash_windows",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/symlink_dot_symlink": ".dir/subdir\\file",
			},
			args: []string{"~/.symlink"},
			tests: []any{
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
					assert.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				assert.NoError(t, newTestConfig(t, fileSystem).execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.tests)
				assert.NoError(t, newTestConfig(t, fileSystem).execute([]string{"verify"}))
			})
		})
	}
}
