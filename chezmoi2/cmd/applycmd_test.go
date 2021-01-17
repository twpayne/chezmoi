package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

func TestApplyCmd(t *testing.T) {
	for _, tc := range []struct {
		name      string
		extraRoot interface{}
		args      []string
		tests     []interface{}
	}{
		{
			name: "all",
			tests: []interface{}{
				vfst.TestPath("/home/user/.absent",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
				),
				vfst.TestPath("/home/user/.dir/subdir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
				vfst.TestPath("/home/user/.empty",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContents(nil),
				),
				vfst.TestPath("/home/user/.executable",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .executable\n"),
				),
				vfst.TestPath("/home/user/.exists",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .exists\n"),
				),
				vfst.TestPath("/home/user/.file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .file\n"),
				),
				vfst.TestPath("/home/user/.private",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o600&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .private\n"),
				),
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
				vfst.TestPath("/home/user/.template",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("key = value\n"),
				),
			},
		},
		{
			name: "all_with_--dry-run",
			args: []string{"--dry-run"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.absent",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.empty",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.executable",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.exists",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.file",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.private",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.symlink",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.template",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dir",
			args: []string{"~/.dir"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
				),
				vfst.TestPath("/home/user/.dir/subdir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
		{
			name: "dir_with_--recursive=false",
			args: []string{"~/.dir", "--recursive=false"},
			tests: []interface{}{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoi.GetUmask()),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.dir/subdir",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "exists",
			args: []string{"~/.exists"},
			extraRoot: map[string]interface{}{
				"/home/user/.exists": "# existing contents of .exists\n",
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.exists",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoi.GetUmask()),
					vfst.TestContentsString("# existing contents of .exists\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]interface{}{
				"/home/user": map[string]interface{}{
					".config": map[string]interface{}{
						"chezmoi": map[string]interface{}{
							"chezmoi.toml": chezmoitest.JoinLines(
								`[data]`,
								`  variable = "value"`,
							),
						},
					},
					".local": map[string]interface{}{
						"share": map[string]interface{}{
							"chezmoi": map[string]interface{}{
								"dot_absent": "",
								"dot_dir": map[string]interface{}{
									"file": "# contents of .dir/file\n",
									"subdir": map[string]interface{}{
										"file": "# contents of .dir/subdir/file\n",
									},
								},
								"empty_dot_empty":           "",
								"executable_dot_executable": "# contents of .executable\n",
								"exists_dot_exists":         "# contents of .exists\n",
								"dot_file":                  "# contents of .file\n",
								"private_dot_private":       "# contents of .private\n",
								"symlink_dot_symlink":       ".dir/subdir/file\n",
								"dot_template.tmpl": chezmoitest.JoinLines(
									`key = {{ "value" }}`,
								),
							},
						},
					},
				},
			}, func(fs vfs.FS) {
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(fs, tc.extraRoot))
				}
				require.NoError(t, newTestConfig(t, fs).execute(append([]string{"apply"}, tc.args...)))
				vfst.RunTests(t, fs, "", tc.tests)
			})
		})
	}
}
