package chezmoi

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-vfs/v4/vfst"

	"github.com/twpayne/chezmoi/v2/internal/archive"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestArchiveGitIgnorePatterns(t *testing.T) {
	data, err := archive.NewTar(map[string]interface{}{
		".gitignore": chezmoitest.JoinLines(
			"# comment",
			"*.ignore",
			"/*.ignorehere",
		),
		"dir": map[string]interface{}{
			".gitignore": chezmoitest.JoinLines(
				"*.ignore",
				"/*.ignorehere",
			),
		},
	})
	require.NoError(t, err)

	expectedPatternSet := newPatternSet()
	assert.NoError(t, expectedPatternSet.add("**/*.ignore", true))
	assert.NoError(t, expectedPatternSet.add("/*.ignorehere", true))
	assert.NoError(t, expectedPatternSet.add("dir/**/*.ignore", true))
	assert.NoError(t, expectedPatternSet.add("dir/*.ignorehere", true))

	actualPatternSet, err := archiveGitIgnorePatterns(data, ArchiveFormatTar)
	require.NoError(t, err)
	assert.Equal(t, expectedPatternSet, actualPatternSet)
}

func TestIgnorePatternFromGitIgnorePattern(t *testing.T) {
	for _, tc := range []struct {
		dir             string
		pattern         string
		expectedPattern string
		expectedInclude bool
	}{
		{
			dir:             ".",
			pattern:         "*.ignore",
			expectedPattern: "**/*.ignore",
			expectedInclude: true,
		},
		{
			dir:             ".",
			pattern:         "/*.ignore",
			expectedPattern: "/*.ignore",
			expectedInclude: true,
		},
		{
			dir:             ".",
			pattern:         "!*.ignore",
			expectedPattern: "**/*.ignore",
			expectedInclude: false,
		},
		{
			dir:             ".",
			pattern:         "!/*.ignore",
			expectedPattern: "/*.ignore",
			expectedInclude: false,
		},
		{
			dir:             "dir",
			pattern:         "*.ignore",
			expectedPattern: "dir/**/*.ignore",
			expectedInclude: true,
		},
		{
			dir:             "dir",
			pattern:         "/*.ignore",
			expectedPattern: "dir/*.ignore",
			expectedInclude: true,
		},
		{
			dir:             "dir",
			pattern:         "!*.ignore",
			expectedPattern: "dir/**/*.ignore",
			expectedInclude: false,
		},
		{
			dir:             "dir",
			pattern:         "!/*.ignore",
			expectedPattern: "dir/*.ignore",
			expectedInclude: false,
		},
	} {
		name := fmt.Sprintf("ignorePatternFromGitIgnorePattern(%q, %q)", tc.dir, tc.pattern)
		t.Run(name, func(t *testing.T) {
			actualPattern, actualInclude := ignorePatternFromGitIgnorePattern(tc.dir, tc.pattern)
			assert.Equal(t, tc.expectedPattern, actualPattern)
			assert.Equal(t, tc.expectedInclude, actualInclude)
		})
	}
}

func TestSourceStateAdd(t *testing.T) {
	for _, tc := range []struct {
		name         string
		destAbsPaths []AbsPath
		addOptions   AddOptions
		extraRoot    interface{}
		tests        []interface{}
	}{
		{
			name: "dir",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/subdir"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/subdir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/subdir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			name: "dir_readonly_unix",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.readonly_dir"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.readonly_dir": &vfst.Dir{Perm: 0o555},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/readonly_dot_readonly_dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0o777&^chezmoitest.Umask),
				),
			},
		},
		{
			name: "empty",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.empty"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_empty",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "empty_with_empty",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.empty"),
			},
			addOptions: AddOptions{
				Empty:   true,
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.executable"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.executable"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.create"),
			},
			addOptions: AddOptions{
				Create:  true,
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.private"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.private"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			name: "file_readonly_unix",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.readonly"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
			},
			extraRoot: map[string]interface{}{
				"/home/user/.readonly": &vfst.File{
					Perm:     0o444,
					Contents: []byte("# contents of .readonly\n"),
				},
			},
			tests: []interface{}{
				vfst.TestPath("/home/user/.local/share/chezmoi/readonly_dot_readonly",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .readonly\n"),
				),
			},
		},
		{
			name: "symlink",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.symlink"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.symlink_windows"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.template"),
			},
			addOptions: AddOptions{
				AutoTemplate: true,
				Include:      NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir"),
				NewAbsPath("/home/user/.dir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/subdir/file"),
			},
			addOptions: AddOptions{
				Include: NewEntryTypeSet(EntryTypesAll),
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
			}, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				persistentState := NewMockPersistentState()
				if tc.extraRoot != nil {
					require.NoError(t, vfst.NewBuilder().Build(system.UnderlyingFS(), tc.extraRoot))
				}

				s := NewSourceState(
					WithBaseSystem(system),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
					withUserTemplateData(map[string]interface{}{
						"variable": "value",
					}),
				)
				require.NoError(t, s.Read(ctx, nil))
				requireEvaluateAll(t, s, system)

				destAbsPathInfos := make(map[AbsPath]fs.FileInfo)
				for _, destAbsPath := range tc.destAbsPaths {
					require.NoError(t, s.AddDestAbsPathInfos(destAbsPathInfos, system, destAbsPath, nil))
				}
				require.NoError(t, s.Add(system, persistentState, system, destAbsPathInfos, &tc.addOptions))

				vfst.RunTests(t, fileSystem, "", tc.tests...)
			})
		})
	}
}

func TestSourceStateAddInExternal(t *testing.T) {
	buffer := &bytes.Buffer{}
	tarWriterSystem := NewTarWriterSystem(buffer, tar.Header{})
	require.NoError(t, tarWriterSystem.Mkdir(NewAbsPath("dir"), 0o777))
	require.NoError(t, tarWriterSystem.WriteFile(NewAbsPath("dir/file"), []byte("# contents of dir/file\n"), 0o666))
	require.NoError(t, tarWriterSystem.Close())
	archiveData := buffer.Bytes()

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(archiveData)
		require.NoError(t, err)
	}))
	defer httpServer.Close()

	root := map[string]interface{}{
		"/home/user": map[string]interface{}{
			".dir/file2": "# contents of .dir/file2\n",
			".local/share/chezmoi": map[string]interface{}{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`[".dir"]`,
					`    type = "archive"`,
					`    url = "`+httpServer.URL+`/archive.tar"`,
					`    stripComponents = 1`,
				),
				"dot_dir": &vfst.Dir{Perm: 0o777},
			},
		},
	}

	chezmoitest.WithTestFS(t, root, func(fileSystem vfs.FS) {
		ctx := context.Background()
		system := NewRealSystem(fileSystem)
		persistentState := NewMockPersistentState()
		s := NewSourceState(
			WithBaseSystem(system),
			WithCacheDir(NewAbsPath("/home/user/.cache/chezmoi")),
			WithDestDir(NewAbsPath("/home/user")),
			WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
			WithSystem(system),
		)
		require.NoError(t, s.Read(ctx, nil))

		destAbsPath := NewAbsPath("/home/user/.dir/file2")
		fileInfo, err := system.Stat(destAbsPath)
		require.NoError(t, err)
		destAbsPathInfos := map[AbsPath]fs.FileInfo{
			destAbsPath: fileInfo,
		}
		require.NoError(t, s.Add(system, persistentState, system, destAbsPathInfos, &AddOptions{
			Include: NewEntryTypeSet(EntryTypesAll),
		}))

		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
				vfst.TestIsDir,
				vfst.TestModePerm(0o777&^chezmoitest.Umask),
			),
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file2",
				vfst.TestModeIsRegular,
				vfst.TestModePerm(0o666&^chezmoitest.Umask),
				vfst.TestContentsString("# contents of .dir/file2\n"),
			),
		)
	})
}

func TestSourceStateApplyAll(t *testing.T) {
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
					vfst.TestModeType(fs.ModeSymlink),
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
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				persistentState := NewMockPersistentState()
				sourceStateOptions := []SourceStateOption{
					WithBaseSystem(system),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
				}
				sourceStateOptions = append(sourceStateOptions, tc.sourceStateOptions...)
				s := NewSourceState(sourceStateOptions...)
				require.NoError(t, s.Read(ctx, nil))
				requireEvaluateAll(t, s, system)
				require.NoError(t, s.applyAll(system, system, persistentState, NewAbsPath("/home/user"), ApplyOptions{
					Include: NewEntryTypeSet(EntryTypesAll),
					Umask:   chezmoitest.Umask,
				}))

				vfst.RunTests(t, fileSystem, "", tc.tests...)
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
					"dir": &vfst.Dir{
						Perm: 0o777 &^ chezmoitest.Umask,
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        "dir",
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777 &^ chezmoitest.Umask,
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
					NewRelPath(".file"): &SourceStateFile{
						origin:        "dot_file",
						sourceRelPath: NewSourceRelPath("dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
						},
						lazyContents: newLazyContents([]byte("# contents of .file\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o666 &^ chezmoitest.Umask,
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
			expectedError: ".file: inconsistent state (dot_file, dot_file.tmpl)",
		},
		{
			name: "duplicate_target_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir": &vfst.Dir{
						Perm: 0o777 &^ chezmoitest.Umask,
					},
					"exact_dir": &vfst.Dir{
						Perm: 0o777 &^ chezmoitest.Umask,
					},
				},
			},
			expectedError: "dir: inconsistent state (dir, exact_dir)",
		},
		{
			name: "duplicate_target_script",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"run_script":      "#!/bin/sh\n",
					"run_once_script": "#!/bin/sh\n",
				},
			},
			expectedError: "script: inconsistent state (run_once_script, run_script)",
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
					NewRelPath(".file"): &SourceStateFile{
						origin:        "executable_dot_file",
						sourceRelPath: NewSourceRelPath("executable_dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
							Executable: true,
						},
						lazyContents: newLazyContents([]byte("# contents of .file\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o777 &^ chezmoitest.Umask,
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
					NewRelPath("script"): &SourceStateFile{
						origin:        "run_script",
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
						},
						lazyContents: newLazyContents([]byte("# contents of .script\n")),
						targetStateEntry: &TargetStateScript{
							name:         NewRelPath("script"),
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
					NewRelPath("script"): &SourceStateFile{
						origin:        "run_script",
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
						},
						lazyContents: newLazyContents([]byte("# contents of script\n")),
						targetStateEntry: &TargetStateScript{
							name:         NewRelPath("script"),
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
					NewRelPath(".symlink"): &SourceStateFile{
						origin:        "symlink_dot_symlink",
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
					NewRelPath("dir"): &SourceStateDir{
						origin:        "dir",
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777 &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file"): &SourceStateFile{
						origin:        "dir/file",
						sourceRelPath: NewSourceRelPath("dir/file"),
						Attr: FileAttr{
							TargetName: "file",
							Type:       SourceFileTypeFile,
						},
						lazyContents: newLazyContents([]byte("# contents of .dir/file\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o666 &^ chezmoitest.Umask,
							lazyContents: newLazyContents([]byte("# contents of .dir/file\n")),
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
					NewRelPath("dir"): &SourceStateDir{
						origin:        "exact_dir",
						sourceRelPath: NewSourceRelDirPath("exact_dir"),
						Attr: DirAttr{
							TargetName: "dir",
							Exact:      true,
						},
						targetStateEntry: &TargetStateDir{
							perm: 0o777 &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file1"): &SourceStateFile{
						origin:        "exact_dir/file1",
						sourceRelPath: NewSourceRelPath("exact_dir/file1"),
						Attr: FileAttr{
							TargetName: "file1",
							Type:       SourceFileTypeFile,
						},
						lazyContents: newLazyContents([]byte("# contents of dir/file1\n")),
						targetStateEntry: &TargetStateFile{
							perm:         0o666 &^ chezmoitest.Umask,
							lazyContents: newLazyContents([]byte("# contents of dir/file1\n")),
						},
					},
					NewRelPath("dir/file2"): &SourceStateRemove{
						sourceRelPath: NewSourceRelDirPath("exact_dir"),
						targetRelPath: NewRelPath("dir/file2"),
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
					NewRelPath("file"): &SourceStateRemove{
						sourceRelPath: NewSourceRelPath(".chezmoiremove"),
						targetRelPath: NewRelPath("file"),
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
					NewRelPath("file1"): &SourceStateRemove{
						sourceRelPath: NewSourceRelPath(".chezmoiremove"),
						targetRelPath: NewRelPath("file1"),
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
			expectedSourceState: NewSourceState(),
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
			expectedError: "source state requires version 2.3.4 or later, chezmoi is version 1.2.3",
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
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				s := NewSourceState(
					WithBaseSystem(system),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
					WithVersion(semver.Version{
						Major: 1,
						Minor: 2,
						Patch: 3,
					}),
				)
				err := s.Read(ctx, nil)
				if tc.expectedError != "" {
					assert.Error(t, err)
					assert.Equal(t, tc.expectedError, err.Error())
					return
				}
				require.NoError(t, err)
				requireEvaluateAll(t, s, system)
				tc.expectedSourceState.destDirAbsPath = NewAbsPath("/home/user")
				tc.expectedSourceState.sourceDirAbsPath = NewAbsPath("/home/user/.local/share/chezmoi")
				requireEvaluateAll(t, tc.expectedSourceState, system)
				s.baseSystem = nil
				s.system = nil
				s.templateData = nil
				s.version = semver.Version{}
				assert.Equal(t, tc.expectedSourceState, s)
			})
		})
	}
}

func TestSourceStateReadExternal(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("data"))
		require.NoError(t, err)
	}))
	defer httpServer.Close()

	for _, tc := range []struct {
		name              string
		root              interface{}
		expectedExternals map[RelPath]External
	}{
		{
			name: "external_yaml",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiexternal.yaml": chezmoitest.JoinLines(
						`file:`,
						`    type: "file"`,
						`    url: "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath]External{
				NewRelPath("file"): {
					Type: "file",
					URL:  httpServer.URL + "/file",
				},
			},
		},
		{
			name: "external_toml",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					".chezmoiexternal.toml": chezmoitest.JoinLines(
						`[file]`,
						`    type = "file"`,
						`    url = "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath]External{
				NewRelPath("file"): {
					Type: "file",
					URL:  httpServer.URL + "/file",
				},
			},
		},
		{
			name: "external_in_subdir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/dot_dir": map[string]interface{}{
					".chezmoiexternal.yaml": chezmoitest.JoinLines(
						`file:`,
						`    type: "file"`,
						`    url: "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath]External{
				NewRelPath(".dir/file"): {
					Type: "file",
					URL:  httpServer.URL + "/file",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				s := NewSourceState(
					WithBaseSystem(system),
					WithCacheDir(NewAbsPath("/home/user/.cache/chezmoi")),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
				)
				require.NoError(t, s.Read(ctx, nil))
				assert.Equal(t, tc.expectedExternals, s.externals)
			})
		})
	}
}

func TestSourceStateReadExternalCache(t *testing.T) {
	buffer := &bytes.Buffer{}
	tarWriterSystem := NewTarWriterSystem(buffer, tar.Header{})
	require.NoError(t, tarWriterSystem.WriteFile(NewAbsPath("file"), []byte("# contents of file\n"), 0o666))
	require.NoError(t, tarWriterSystem.Close())
	archiveData := buffer.Bytes()

	httpRequests := 0
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequests++
		_, err := w.Write(archiveData)
		require.NoError(t, err)
	}))
	defer httpServer.Close()

	now := time.Now()
	readOptions := &ReadOptions{
		TimeNow: func() time.Time {
			return now
		},
	}

	chezmoitest.WithTestFS(t, map[string]interface{}{
		"/home/user/.local/share/chezmoi": map[string]interface{}{
			".chezmoiexternal.yaml": chezmoitest.JoinLines(
				`.dir:`,
				`    type: "archive"`,
				`    url: "`+httpServer.URL+`/archive.tar"`,
				`    refreshPeriod: "1m"`,
			),
		},
	}, func(fileSystem vfs.FS) {
		ctx := context.Background()
		system := NewRealSystem(fileSystem)

		readSourceState := func() {
			s := NewSourceState(
				WithBaseSystem(system),
				WithCacheDir(NewAbsPath("/home/user/.cache/chezmoi")),
				WithDestDir(NewAbsPath("/home/user")),
				WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
				WithSystem(system),
			)
			require.NoError(t, s.Read(ctx, readOptions))
			assert.Equal(t, map[RelPath]External{
				NewRelPath(".dir"): {
					Type:          "archive",
					URL:           httpServer.URL + "/archive.tar",
					RefreshPeriod: 1 * time.Minute,
				},
			}, s.externals)
		}

		readSourceState()
		assert.Equal(t, 1, httpRequests)

		now = now.Add(10 * time.Second)
		readSourceState()
		assert.Equal(t, 1, httpRequests)

		now = now.Add(1 * time.Minute)
		readSourceState()
		assert.Equal(t, 2, httpRequests)
	})
}

func TestSourceStateTargetRelPaths(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		root                   interface{}
		expectedTargetRelPaths []RelPath
	}{
		{
			name:                   "empty",
			root:                   nil,
			expectedTargetRelPaths: []RelPath{},
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
			expectedTargetRelPaths: []RelPath{
				NewRelPath("1before"),
				NewRelPath("2before"),
				NewRelPath("3before"),
				NewRelPath("1"),
				NewRelPath("2"),
				NewRelPath("3"),
				NewRelPath("1after"),
				NewRelPath("2after"),
				NewRelPath("3after"),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				s := NewSourceState(
					WithBaseSystem(system),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
				)
				require.NoError(t, s.Read(ctx, nil))
				assert.Equal(t, tc.expectedTargetRelPaths, s.TargetRelPaths())
			})
		})
	}
}

func TestWalkSourceDir(t *testing.T) {
	sourceDirAbsPath := NewAbsPath("/home/user/.local/share/chezmoi")
	root := map[string]interface{}{
		sourceDirAbsPath.String(): map[string]interface{}{
			".chezmoi.toml.tmpl":    "",
			".chezmoidata.json":     "",
			".chezmoidata.toml":     "",
			".chezmoidata.yaml":     "",
			".chezmoiexternal.yaml": "",
			".chezmoiignore":        "",
			".chezmoiremove":        "",
			".chezmoitemplates":     &vfst.Dir{Perm: 0o777},
			".chezmoiversion":       "",
			"dot_file":              "",
		},
	}
	expectedAbsPaths := []AbsPath{
		sourceDirAbsPath,
		sourceDirAbsPath.JoinString(".chezmoiversion"),
		sourceDirAbsPath.JoinString(".chezmoidata.json"),
		sourceDirAbsPath.JoinString(".chezmoidata.toml"),
		sourceDirAbsPath.JoinString(".chezmoidata.yaml"),
		sourceDirAbsPath.JoinString(".chezmoitemplates"),
		sourceDirAbsPath.JoinString(".chezmoi.toml.tmpl"),
		sourceDirAbsPath.JoinString(".chezmoiexternal.yaml"),
		sourceDirAbsPath.JoinString(".chezmoiignore"),
		sourceDirAbsPath.JoinString(".chezmoiremove"),
		sourceDirAbsPath.JoinString("dot_file"),
	}

	var actualAbsPaths []AbsPath
	chezmoitest.WithTestFS(t, root, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
		require.NoError(t, WalkSourceDir(system, sourceDirAbsPath, func(absPath AbsPath, fileInfo fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			actualAbsPaths = append(actualAbsPaths, absPath)
			return nil
		}))
	})
	assert.Equal(t, expectedAbsPaths, actualAbsPaths)
}

// applyAll updates targetDir in targetSystem to match s.
func (s *SourceState) applyAll(targetSystem, destSystem System, persistentState PersistentState, targetDir AbsPath, options ApplyOptions) error {
	for _, targetRelPath := range s.TargetRelPaths() {
		switch err := s.Apply(targetSystem, destSystem, persistentState, targetDir, targetRelPath, options); {
		case errors.Is(err, Skip):
			continue
		case err != nil:
			return err
		}
	}
	return nil
}

// requireEvaluateAll requires that every target state entry in s evaluates
// without error.
func requireEvaluateAll(t *testing.T, s *SourceState, destSystem System) {
	t.Helper()
	require.NoError(t, s.root.ForEach(EmptyRelPath, func(targetRelPath RelPath, sourceStateEntry SourceStateEntry) error {
		if err := sourceStateEntry.Evaluate(); err != nil {
			return err
		}
		destAbsPath := s.destDirAbsPath.Join(targetRelPath)
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(destSystem, destAbsPath)
		if err != nil {
			return err
		}
		return targetStateEntry.Evaluate()
	}))
}

func withEntries(sourceEntries map[RelPath]SourceStateEntry) SourceStateOption {
	return func(s *SourceState) {
		s.root = sourceStateEntryTreeNode{}
		for targetRelPath, sourceStateEntry := range sourceEntries {
			s.root.Set(targetRelPath, sourceStateEntry)
		}
	}
}

func withIgnore(ignore *patternSet) SourceStateOption {
	return func(s *SourceState) {
		s.ignore = ignore
	}
}

// withUserTemplateData adds template data.
func withUserTemplateData(templateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		RecursiveMerge(s.userTemplateData, templateData)
	}
}

func withTemplates(templates map[string]*template.Template) SourceStateOption {
	return func(s *SourceState) {
		s.templates = templates
	}
}
