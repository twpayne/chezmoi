package cmd

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestApplyCmd(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extraRoot any
		args      []string
		tests     []any
	}{
		{
			name: "all",
			tests: []any{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .create\n"),
				),
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/subdir/file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
				vfst.TestPath("/home/user/.empty",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContents(nil),
				),
				vfst.TestPath("/home/user/.executable",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .executable\n"),
				),
				vfst.TestPath("/home/user/.file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
				vfst.TestPath("/home/user/.private",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o600&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .private\n"),
				),
				vfst.TestPath("/home/user/.remove",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
				vfst.TestPath("/home/user/.template",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("key = value\n"),
				),
			},
		},
		{
			name: "all_with_--dry-run",
			args: []string{"--dry-run"},
			tests: []any{
				vfst.TestPath("/home/user/.create",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.dir",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.empty",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.executable",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.file",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.private",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.remove",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.symlink",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.template",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "dir",
			args: []string{"~/.dir"},
			tests: []any{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/subdir/file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
		{
			name: "dir_with_--recursive=false",
			args: []string{"~/.dir", "--recursive=false"},
			tests: []any{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "create",
			args: []string{"~/.create"},
			extraRoot: map[string]any{
				"/home/user/.create": "# existing contents of .create\n",
			},
			tests: []any{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# existing contents of .create\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user": map[string]any{
					".config": map[string]any{
						"chezmoi": map[string]any{
							"chezmoi.toml": chezmoitest.JoinLines(
								`[data]`,
								`  variable = "value"`,
							),
						},
					},
					".local": map[string]any{
						"share": map[string]any{
							"chezmoi": map[string]any{
								"create_dot_create": "# contents of .create\n",
								"dot_dir": map[string]any{
									"file": "# contents of .dir/file\n",
									"subdir": map[string]any{
										"file": "# contents of .dir/subdir/file\n",
									},
								},
								"dot_file":   "# contents of .file\n",
								"dot_remove": "",
								"dot_template.tmpl": chezmoitest.JoinLines(
									`key = {{ "value" }}`,
								),
								"empty_dot_empty":           "",
								"executable_dot_executable": "# contents of .executable\n",
								"private_dot_private":       "# contents of .private\n",
								"symlink_dot_symlink":       ".dir/subdir/file\n",
							},
						},
					},
				},
			}, func(fileSystem vfs.FS) {
				if tc.extraRoot != nil {
					assert.NoError(t, vfst.NewBuilder().Build(fileSystem, tc.extraRoot))
				}
				assert.NoError(t, newTestConfig(t, fileSystem).execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fileSystem, "", tc.tests)
			})
		})
	}
}

func TestIssue2132(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]interface{}{
		"/home/user/.local/share/chezmoi/remove_dot_dir/non_existent_file": "",
	}, func(fileSystem vfs.FS) {
		config1 := newTestConfig(t, fileSystem)
		assert.NoError(t, config1.execute([]string{"apply"}))
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath("/home/user/.dir",
				vfst.TestDoesNotExist(),
			),
		)
		config2 := newTestConfig(t, fileSystem)
		assert.NoError(t, config2.execute([]string{"apply", "--no-tty"}))
		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath("/home/user/.dir",
				vfst.TestDoesNotExist(),
			),
		)
	})
}

func TestIssue3206(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user": map[string]any{
			".local/share/chezmoi": map[string]any{
				".chezmoiignore": "",
				"dot_config/private_expanso/match": map[string]any{
					".chezmoidata.yaml": "key: value\n",
					"greek.yml.tmpl":    "{{ .key }}",
				},
			},
		},
	}, func(fileSystem vfs.FS) {
		assert.NoError(t, newTestConfig(t, fileSystem).execute([]string{"apply"}))
	})
}

func TestIssue3216(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user": map[string]any{
			".local/share/chezmoi": map[string]any{
				".chezmoiignore": "",
				"dot_config/private_expanso/match": map[string]any{
					".chezmoidata.yaml": "",
					"greek.yml.tmpl":    "{{ .chezmoi.os }}",
				},
			},
		},
	}, func(fileSystem vfs.FS) {
		assert.NoError(t, newTestConfig(t, fileSystem).execute([]string{"apply"}))
	})
}
