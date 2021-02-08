package chezmoi

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

func TestSourceStateAdd(t *testing.T) {
	for _, tc := range []struct {
		name         string
		destAbsPaths AbsPaths
		addOptions   AddOptions
		extraRoot    interface{}
		tests        []interface{}
	}{
		{
			name: "dir",
			destAbsPaths: AbsPaths{
				"/home/user/.dir",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dir_change_attributes",
			destAbsPaths: AbsPaths{
				"/home/user/.dir",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi/exact_dot_dir/file": "# contents of file\n",
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/exact_dot_dir",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of file\n"),
				),
			},
		},
		{
			name: "dir_file",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
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
			name: "dir_file_existing_dir",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_dir": &vfst.Dir{Perm: 0o777},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
			},
		},
		{
			name: "dir_subdir",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/subdir",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dir_subdir_file",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/subdir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
		{
			name: "dir_subdir_file_existing_dir_subdir",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/subdir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_dir/subdir": &vfst.Dir{Perm: 0o777},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
		{
			name: "empty",
			destAbsPaths: AbsPaths{
				"/home/user/.empty",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_empty",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "empty_with_empty",
			destAbsPaths: AbsPaths{
				"/home/user/.empty",
			},
			addOptions: AddOptions{
				Empty:   true,
				Include: NewIncludeSet(IncludeAll),
			},
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
			destAbsPaths: AbsPaths{
				"/home/user/.executable",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/executable_dot_executable",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .executable\n"),
				),
			},
		},
		{
			name: "executable_windows",
			destAbsPaths: AbsPaths{
				"/home/user/.executable",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_executable",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .executable\n"),
				),
			},
		},
		{
			name: "create",
			destAbsPaths: AbsPaths{
				"/home/user/.create",
			},
			addOptions: AddOptions{
				Create:  true,
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/create_dot_create",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .create\n"),
				),
			},
		},
		{
			name: "file",
			destAbsPaths: AbsPaths{
				"/home/user/.file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
		{
			name: "file_change_attributes",
			destAbsPaths: AbsPaths{
				"/home/user/.file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/executable_dot_file": "# contents of .file\n",
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/executable_dot_file",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "file_replace_contents",
			destAbsPaths: AbsPaths{
				"/home/user/.file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_file": "# old contents of .file\n",
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
		{
			name: "private_unix",
			destAbsPaths: AbsPaths{
				"/home/user/.private",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_private",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .private\n"),
				),
			},
		},
		{
			name: "private_windows",
			destAbsPaths: AbsPaths{
				"/home/user/.private",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_private",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .private\n"),
				),
			},
		},
		{
			name: "symlink",
			destAbsPaths: AbsPaths{
				"/home/user/.symlink",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/symlink_dot_symlink",
					vfst.TestModeIsRegular,
					vfst.TestContentsString(".dir/subdir/file\n"),
				),
			},
		},
		{
			name: "symlink_backslash_windows",
			destAbsPaths: AbsPaths{
				"/home/user/.symlink_windows",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".symlink_windows": &vfst.Symlink{Target: ".dir\\subdir\\file"},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/symlink_dot_symlink_windows",
					vfst.TestModeIsRegular,
					vfst.TestContentsString(".dir/subdir/file\n"),
				),
			},
		},
		{
			name: "template",
			destAbsPaths: AbsPaths{
				"/home/user/.template",
			},
			addOptions: AddOptions{
				AutoTemplate: true,
				Include:      NewIncludeSet(IncludeAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_template.tmpl",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("key = {{ .variable }}\n"),
				),
			},
		},
		{
			name: "dir_and_dir_file",
			destAbsPaths: AbsPaths{
				"/home/user/.dir",
				"/home/user/.dir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
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
			name: "file_in_dir_exact_subdir",
			destAbsPaths: AbsPaths{
				"/home/user/.dir/subdir/file",
			},
			addOptions: AddOptions{
				Include: NewIncludeSet(IncludeAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_dir/exact_subdir": &vfst.Dir{Perm: 0o777},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/exact_subdir/file",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)

			chezmoitest.WithTestFS(t, map[string]interface{}{
				"/home/user": map[string]interface{}{
					".create": "# contents of .create\n",
					".dir": map[string]interface{}{
						"file": "# contents of .dir/file\n",
						"subdir": map[string]interface{}{
							"file": "# contents of .dir/subdir/file\n",
						},
					},
					".empty": "",
					".executable": &vfst.File{
						Perm:     0o777,
						Contents: []byte("# contents of .executable\n"),
					},
					".file": "# contents of .file\n",
					".local": map[string]interface{}{
						"share": map[string]interface{}{
							"chezmoi": &vfst.Dir{Perm: 0o777},
						},
					},
					".private": &vfst.File{
						Perm:     0o600,
						Contents: []byte("# contents of .private\n"),
					},
					".symlink":  &vfst.Symlink{Target: ".dir/subdir/file"},
					".template": "key = value\n",
				},
			}, func(fs vfs.FS) {
				system := NewRealSystem(fs)
				persistentState := NewMockPersistentState()
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(system.UnderlyingFS(), tc.extraRoot))
				}

				s := NewSourceState(
					WithDestDir("/home/user"),
					WithSourceDir("/home/user/.local/share/chezmoi"),
					WithSystem(system),
					withUserTemplateData(map[string]interface{}{
						"variable": "value",
					}),
				)
				require.NoError(t, s.Read())
				requireEvaluateAll(t, s, system)

				destAbsPathInfos := make(map[AbsPath]os.FileInfo)
				for _, destAbsPath := range tc.destAbsPaths {
					require.NoError(t, s.AddDestAbsPathInfos(destAbsPathInfos, system, destAbsPath, nil))
				}
				require.NoError(t, s.Add(system, persistentState, system, destAbsPathInfos, &tc.addOptions))

				vfst.RunTests(t, fs, "", tc.tests...)
			})
		})
	}
}

func TestSourceStateApplyAll(t *testing.T) {
	// FIXME script tests
	// FIXME script template tests

	for _, tc := range []struct {
		name               string
		root               interface{}
		sourceStateOptions []SourceStateOption
		tests              []interface{}
	}{
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": &vfst.Dir{Perm: 0o777},
				},
			},
		},
		{
			name: "dir",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"dot_dir": &vfst.Dir{Perm: 0o777},
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
			},
		},
		{
			name: "dir_exact",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".dir": map[string]interface{}{
						"file": "# contents of .dir/file\n",
					},
					".local/share/chezmoi": map[string]interface{}{
						"exact_dot_dir": &vfst.Dir{Perm: 0o777},
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"dot_file": "# contents of .file\n",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
		{
			name: "file_remove_empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".empty": "# contents of .empty\n",
					".local/share/chezmoi": map[string]interface{}{
						"dot_empty": "",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.empty",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "file_create_empty",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"empty_dot_empty": "",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.empty",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContents(nil),
				),
			},
		},
		{
			name: "file_template",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"dot_template.tmpl": "key = {{ .variable }}\n",
					},
				},
			},
			sourceStateOptions: []SourceStateOption{
				withUserTemplateData(map[string]interface{}{
					"variable": "value",
				}),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.template",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("key = value\n"),
				),
			},
		},
		{
			name: "create",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"create_dot_create": "# contents of .create\n",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .create\n"),
				),
			},
		},
		{
			name: "create_no_replace",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"create_dot_create": "# contents of .create\n",
					},
					".create": "# existing contents of .create\n",
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# existing contents of .create\n"),
				),
			},
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"symlink_dot_symlink": ".dir/subdir/file\n",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
			},
		},
		{
			name: "symlink_template",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".local/share/chezmoi": map[string]interface{}{
						"symlink_dot_symlink.tmpl": `{{ ".dir/subdir/file" }}` + "\n",
					},
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				system := NewRealSystem(fs)
				persistentState := NewMockPersistentState()
				sourceStateOptions := []SourceStateOption{
					WithDestDir("/home/user"),
					WithSourceDir("/home/user/.local/share/chezmoi"),
					WithSystem(system),
				}
				sourceStateOptions = append(sourceStateOptions, tc.sourceStateOptions...)
				s := NewSourceState(sourceStateOptions...)
				require.NoError(t, s.Read())
				requireEvaluateAll(t, s, system)
				require.NoError(t, s.applyAll(system, system, persistentState, "/home/user", ApplyOptions{
					Umask: chezmoitest.Umask,
				}))

				vfst.RunTests(t, fs, "", tc.tests...)
			})
		})
	}
}

func TestSourceStateRead(t *testing.T) {
	for _, tc := range []struct {
		name                string
		root                interface{}
		expectedError       string
		expectedSourceState *SourceState
	}{
		{
			name: "empty",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": &vfst.Dir{Perm: 0o777},
			},
			expectedSourceState: NewSourceState(),
		},
		{
			name: "dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir": &vfst.Dir{Perm: 0o777},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"dir": &SourceStateDir{
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777,
						},
					},
				}),
			),
		},
		{
			name: "file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dot_file": "# contents of .file\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					".file": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
						},
						lazyContents: newLazyContents([]byte("# contents of .file\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o666,
							lazyContents: newLazyContents([]byte("# contents of .file\n")),
						},
					},
				}),
			),
		},
		{
			name: "duplicate_target_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dot_file":      "# contents of .file\n",
					"dot_file.tmpl": "# contents of .file\n",
				},
			},
			expectedError: ".file: duplicate source state entries (dot_file, dot_file.tmpl)",
		},
		{
			name: "duplicate_target_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir":       &vfst.Dir{Perm: 0o777},
					"exact_dir": &vfst.Dir{Perm: 0o777},
				},
			},
			expectedError: "dir: duplicate source state entries (dir, exact_dir)",
		},
		{
			name: "duplicate_target_script",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_script":      "#!/bin/sh\n",
					"run_once_script": "#!/bin/sh\n",
				},
			},
			expectedError: "script: duplicate source state entries (run_once_script, run_script)",
		},
		{
			name: "symlink_with_attr",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".file":               "# contents of .file\n",
					"executable_dot_file": &vfst.Symlink{Target: ".file"},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					".file": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("executable_dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
							Executable: true,
						},
						lazyContents: newLazyContents([]byte("# contents of .file\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o777,
							lazyContents: newLazyContents([]byte("# contents of .file\n")),
						},
					},
				}),
			),
		},
		{
			name: "symlink_script",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".script":    "# contents of .script\n",
					"run_script": &vfst.Symlink{Target: ".script"},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"script": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
						},
						lazyContents: newLazyContents([]byte("# contents of .script\n")),
						targetStateEntry: &TargetStateScript{
							name:         "script",
							lazyContents: newLazyContents([]byte("# contents of .script\n")),
						},
					},
				}),
			),
		},
		{
			name: "script",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_script": "# contents of script\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"script": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
						},
						lazyContents: newLazyContents([]byte("# contents of script\n")),
						targetStateEntry: &TargetStateScript{
							name:         "script",
							lazyContents: newLazyContents([]byte("# contents of script\n")),
						},
					},
				}),
			),
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"symlink_dot_symlink": ".dir/subdir/file",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					".symlink": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("symlink_dot_symlink"),
						Attr: FileAttr{
							TargetName: ".symlink",
							Type:       SourceFileTypeSymlink,
						},
						lazyContents: newLazyContents([]byte(".dir/subdir/file")),
						targetStateEntry: &TargetStateSymlink{
							lazyLinkname: newLazyLinkname(".dir/subdir/file"),
						},
					},
				}),
			),
		},
		{
			name: "file_in_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir": map[string]interface{}{
						"file": "# contents of .dir/file\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"dir": &SourceStateDir{
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777,
						},
					},
					"dir/file": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("dir/file"),
						Attr: FileAttr{
							TargetName: "file",
							Type:       SourceFileTypeFile,
						},
						lazyContents: &lazyContents{
							contents: []byte("# contents of .dir/file\n"),
						},
						targetStateEntry: &TargetStateFile{
							perm: 0o666,
							lazyContents: &lazyContents{
								contents: []byte("# contents of .dir/file\n"),
							},
						},
					},
				}),
			),
		},
		{
			name: "chezmoiignore",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "README.md\n",
				},
			},
			expectedSourceState: NewSourceState(
				withIgnore(
					mustNewPatternSet(t, map[string]bool{
						"README.md": true,
					}),
				),
			),
		},
		{
			name: "chezmoiignore_ignore_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "README.md\n",
					"README.md":      "",
				},
			},
			expectedSourceState: NewSourceState(
				withIgnore(
					mustNewPatternSet(t, map[string]bool{
						"README.md": true,
					}),
				),
			),
		},
		{
			name: "chezmoiignore_exact_dir",
			root: map[string]interface{}{
				"/home/user/dir": map[string]interface{}{
					"file1": "# contents of dir/file1\n",
					"file2": "# contents of dir/file2\n",
					"file3": "# contents of dir/file3\n",
				},
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "dir/file3\n",
					"exact_dir": map[string]interface{}{
						"file1": "# contents of dir/file1\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"dir": &SourceStateDir{
						sourceRelPath: NewSourceRelDirPath("exact_dir"),
						Attr: DirAttr{
							TargetName: "dir",
							Exact:      true,
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777,
						},
					},
					"dir/file1": &SourceStateFile{
						sourceRelPath: NewSourceRelPath("exact_dir/file1"),
						Attr: FileAttr{
							TargetName: "file1",
							Type:       SourceFileTypeFile,
						},
						lazyContents: &lazyContents{
							contents: []byte("# contents of dir/file1\n"),
						},
						targetStateEntry: &TargetStateFile{
							perm: 0o666,
							lazyContents: &lazyContents{
								contents: []byte("# contents of dir/file1\n"),
							},
						},
					},
					"dir/file2": &SourceStateRemove{
						targetRelPath: "dir/file2",
					},
				}),
				withIgnore(
					mustNewPatternSet(t, map[string]bool{
						"dir/file3": true,
					}),
				),
			),
		},
		{
			name: "chezmoiremove",
			root: map[string]interface{}{
				"/home/user/file": "",
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiremove": "file\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"file": &SourceStateRemove{
						targetRelPath: "file",
					},
				}),
			),
		},
		{
			name: "chezmoiremove_and_ignore",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					"file1": "",
					"file2": "",
				},
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiignore": "file2\n",
					".chezmoiremove": "file*\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"file1": &SourceStateRemove{
						targetRelPath: "file1",
					},
				}),
				withIgnore(
					mustNewPatternSet(t, map[string]bool{
						"file2": true,
					}),
				),
			),
		},
		{
			name: "chezmoitemplates",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoitemplates": map[string]interface{}{
						"template": "# contents of .chezmoitemplates/template\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withTemplates(
					map[string]*template.Template{
						"template": template.Must(template.New("template").Option("missingkey=error").Parse("# contents of .chezmoitemplates/template\n")),
					},
				),
			),
		},
		{
			name: "chezmoiversion",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiversion": "1.2.3\n",
				},
			},
			expectedSourceState: NewSourceState(
				withMinVersion(
					semver.Version{
						Major: 1,
						Minor: 2,
						Patch: 3,
					},
				),
			),
		},
		{
			name: "chezmoiversion_multiple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiversion": "1.2.3\n",
					"dir": map[string]interface{}{
						".chezmoiversion": "2.3.4\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					"dir": &SourceStateDir{
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777,
						},
					},
				}),
				withMinVersion(
					semver.Version{
						Major: 2,
						Minor: 3,
						Patch: 4,
					},
				),
			),
		},
		{
			name: "ignore_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".dir": map[string]interface{}{
						"file": "# contents of .dir/file\n",
					},
				},
			},
			expectedSourceState: NewSourceState(),
		},
		{
			name: "ignore_file",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".file": "# contents of .file\n",
				},
			},
			expectedSourceState: NewSourceState(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				system := NewRealSystem(fs)
				s := NewSourceState(
					WithDestDir("/home/user"),
					WithSourceDir("/home/user/.local/share/chezmoi"),
					WithSystem(system),
				)
				err := s.Read()
				if tc.expectedError != "" {
					assert.Error(t, err)
					assert.Equal(t, tc.expectedError, err.Error())
					return
				}
				require.NoError(t, err)
				requireEvaluateAll(t, s, system)
				tc.expectedSourceState.destDirAbsPath = "/home/user"
				tc.expectedSourceState.sourceDirAbsPath = "/home/user/.local/share/chezmoi"
				requireEvaluateAll(t, tc.expectedSourceState, system)
				s.system = nil
				s.templateData = nil
				assert.Equal(t, tc.expectedSourceState, s)
			})
		})
	}
}

func TestSourceStateTargetRelPaths(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		root                   interface{}
		expectedTargetRelPaths RelPaths
	}{
		{
			name:                   "empty",
			root:                   nil,
			expectedTargetRelPaths: RelPaths{},
		},
		{
			name: "scripts",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_before_1before": "",
					"run_before_2before": "",
					"run_before_3before": "",
					"run_1":              "",
					"run_2":              "",
					"run_3":              "",
					"run_after_1after":   "",
					"run_after_2after":   "",
					"run_after_3after":   "",
				},
			},
			expectedTargetRelPaths: RelPaths{
				"1before",
				"2before",
				"3before",
				"1",
				"2",
				"3",
				"1after",
				"2after",
				"3after",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				s := NewSourceState(
					WithSourceDir("/home/user/.local/share/chezmoi"),
					WithSystem(NewRealSystem(fs)),
				)
				require.NoError(t, s.Read())
				assert.Equal(t, tc.expectedTargetRelPaths, s.TargetRelPaths())
			})
		})
	}
}

// requireEvaluateAll requires that every target state entry in s evaluates
// without error.
func requireEvaluateAll(t *testing.T, s *SourceState, destSystem System) {
	t.Helper()
	for _, targetRelPath := range s.TargetRelPaths() {
		sourceStateEntry := s.entries[targetRelPath]
		require.NoError(t, sourceStateEntry.Evaluate())
		destAbsPath := s.destDirAbsPath.Join(targetRelPath)
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(destSystem, destAbsPath)
		require.NoError(t, err)
		require.NoError(t, targetStateEntry.Evaluate())
	}
}

func withEntries(sourceEntries map[RelPath]SourceStateEntry) SourceStateOption {
	return func(s *SourceState) {
		s.entries = sourceEntries
	}
}

func withIgnore(ignore *patternSet) SourceStateOption {
	return func(s *SourceState) {
		s.ignore = ignore
	}
}

func withMinVersion(minVersion semver.Version) SourceStateOption {
	return func(s *SourceState) {
		s.minVersion = minVersion
	}
}

// withUserTemplateData adds template data.
func withUserTemplateData(templateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		recursiveMerge(s.userTemplateData, templateData)
	}
}

func withTemplates(templates map[string]*template.Template) SourceStateOption {
	return func(s *SourceState) {
		s.templates = templates
	}
}
