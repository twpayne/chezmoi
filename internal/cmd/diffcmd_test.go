package cmd

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestDiffCmd(t *testing.T) {
	if chezmoitest.Umask != 0o22 {
		t.Skip("umask not 0o22")
	}
	for _, tc := range []struct {
		name      string
		extraRoot any
		args      []string
		stdoutStr string
	}{
		{
			name: "empty",
		},
		{
			name: "file",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dot_file": "# contents of .file\n",
				},
			},
			stdoutStr: chezmoitest.JoinLines(
				`diff --git a/.file b/.file`,
				`new file mode 100644`,
				`index 0000000000000000000000000000000000000000..8a52cb9ce9551221716a53786ad74104c5902362`,
				`--- /dev/null`,
				`+++ b/.file`,
				`@@ -0,0 +1 @@`,
				`+# contents of .file`,
			),
		},
		{
			name: "simple_exclude_files",
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dot_file":            "# contents of .file\n",
					"symlink_dot_symlink": ".file\n",
				},
			},
			args: []string{
				"--exclude", "files",
			},
			stdoutStr: chezmoitest.JoinLines(
				`diff --git a/.symlink b/.symlink`,
				`new file mode 120000`,
				`index 0000000000000000000000000000000000000000..3e6844d17780d623d817c3e22bcd1128d64422ae`,
				`--- /dev/null`,
				`+++ b/.symlink`,
				`@@ -0,0 +1 @@`,
				`+.file`,
			),
		},
		{
			name: "simple_exclude_files_with_config",
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".config/chezmoi/chezmoi.toml": chezmoitest.JoinLines(
						`[diff]`,
						`    exclude = ["files"]`,
					),
					".local/share/chezmoi": map[string]any{
						"dot_file":            "# contents of .file\n",
						"symlink_dot_symlink": ".file\n",
					},
				},
			},
			stdoutStr: chezmoitest.JoinLines(
				`diff --git a/.symlink b/.symlink`,
				`new file mode 120000`,
				`index 0000000000000000000000000000000000000000..3e6844d17780d623d817c3e22bcd1128d64422ae`,
				`--- /dev/null`,
				`+++ b/.symlink`,
				`@@ -0,0 +1 @@`,
				`+.file`,
			),
		},
		{
			name: "simple_exclude_externals_with_config",
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".config/chezmoi/chezmoi.toml": chezmoitest.JoinLines(
						`[diff]`,
						`    exclude = ["externals"]`,
					),
					".local/share/chezmoi": map[string]any{
						"dot_file":            "# contents of .file\n",
						"symlink_dot_symlink": ".file\n",
					},
				},
			},
			stdoutStr: chezmoitest.JoinLines(
				`diff --git a/.file b/.file`,
				`new file mode 100644`,
				`index 0000000000000000000000000000000000000000..8a52cb9ce9551221716a53786ad74104c5902362`,
				`--- /dev/null`,
				`+++ b/.file`,
				`@@ -0,0 +1 @@`,
				`+# contents of .file`,
				`diff --git a/.symlink b/.symlink`,
				`new file mode 120000`,
				`index 0000000000000000000000000000000000000000..3e6844d17780d623d817c3e22bcd1128d64422ae`,
				`--- /dev/null`,
				`+++ b/.symlink`,
				`@@ -0,0 +1 @@`,
				`+.file`,
			),
		},
		{
			name: "issue_4425",
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".config/git/config": "# contents of .config/git/config\n",
					".local/share/chezmoi": map[string]any{
						"dot_config/remove_git/config": "# contents of .config/git/config\n",
					},
				},
			},
		},
		{
			name: "issue_4425_nested_dir",
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".config/git": map[string]any{
						"subdir": &vfst.Dir{
							Perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					".local/share/chezmoi": map[string]any{
						"dot_config/remove_git/remove_subdir/.keep": "",
					},
				},
			},
			stdoutStr: chezmoitest.JoinLines(
				"diff --git a/.config/git/subdir b/.config/git/subdir",
				"deleted file mode 40755",
				"index e69de29bb2d1d6434b8b29ae775ad8c2e48c5391..0000000000000000000000000000000000000000",
				"--- a/.config/git/subdir",
				"+++ /dev/null",
				"diff --git a/.config/git b/.config/git",
				"deleted file mode 40755",
				"index e69de29bb2d1d6434b8b29ae775ad8c2e48c5391..0000000000000000000000000000000000000000",
				"--- a/.config/git",
				"+++ /dev/null",
			),
		},
		{
			name: "issue_4425_nested_file",
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".config/git/config": "# contents of .config/git/config\n",
					".local/share/chezmoi": map[string]any{
						"dot_config/remove_git/config": "",
					},
				},
			},
			stdoutStr: chezmoitest.JoinLines(
				"diff --git a/.config/git/config b/.config/git/config",
				"deleted file mode 100644",
				"index c3d477a648b5cb9739a359d54234ec5627c3a64b..0000000000000000000000000000000000000000",
				"--- a/.config/git/config",
				"+++ /dev/null",
				"@@ -1 +0,0 @@",
				"-# contents of .config/git/config",
				"diff --git a/.config/git b/.config/git",
				"deleted file mode 40755",
				"index e69de29bb2d1d6434b8b29ae775ad8c2e48c5391..0000000000000000000000000000000000000000",
				"--- a/.config/git",
				"+++ /dev/null",
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user/.local/share/chezmoi": &vfst.Dir{
					Perm: fs.ModePerm &^ chezmoitest.Umask,
				},
			}, func(fileSystem vfs.FS) {
				if tc.extraRoot != nil {
					assert.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				stdout := strings.Builder{}
				config := newTestConfig(t, fileSystem, withStdout(&stdout))
				assert.NoError(t, config.execute(append([]string{"diff"}, tc.args...)))
				assert.Equal(t, tc.stdoutStr, stdout.String())
			})
		})
	}
}
