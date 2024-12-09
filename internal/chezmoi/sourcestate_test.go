package chezmoi

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"text/template"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestSourceStateAdd(t *testing.T) {
	for _, tc := range []struct {
		name         string
		destAbsPaths []AbsPath
		addOptions   AddOptions
		extraRoot    any
		tests        []any
	}{
		{
			name: "dir",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "dir_change_attributes",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi/exact_dot_dir/file": "# contents of .dir/file\n",
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/exact_dot_dir",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .dir/file\n"),
				),
			},
		},
		{
			name: "dir_file",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/file"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dot_dir": &vfst.Dir{Perm: fs.ModePerm},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "dir_subdir_file",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.dir/subdir/file"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestDoesNotExist(),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dot_dir/subdir": &vfst.Dir{Perm: fs.ModePerm},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/subdir/file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.readonly_dir": &vfst.Dir{Perm: 0o555},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/readonly_dot_readonly_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
			},
		},
		{
			name: "empty",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.empty"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_empty",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "empty_with_empty",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.empty"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/empty_dot_empty",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/executable_dot_executable",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_executable",
					vfst.TestModeIsRegular(),
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
				Create: true,
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/create_dot_create",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/executable_dot_file": "# contents of .file\n",
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/executable_dot_file",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "file_replace_contents",
			destAbsPaths: []AbsPath{
				NewAbsPath("/home/user/.file"),
			},
			addOptions: AddOptions{
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dot_file": "# old contents of .file\n",
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/private_dot_private",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_private",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.readonly": &vfst.File{
					Perm:     0o444,
					Contents: []byte("# contents of .readonly\n"),
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/readonly_dot_readonly",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/symlink_dot_symlink",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user": map[string]any{
					".symlink_windows": &vfst.Symlink{Target: ".dir\\subdir\\file"},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/symlink_dot_symlink_windows",
					vfst.TestModeIsRegular(),
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
				Filter:       NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_template.tmpl",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file",
					vfst.TestModeIsRegular(),
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
				Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
			},
			extraRoot: map[string]any{
				"/home/user/.local/share/chezmoi/dot_dir/exact_subdir": &vfst.Dir{Perm: fs.ModePerm},
			},
			tests: []any{
				vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/exact_subdir/file",
					vfst.TestModeIsRegular(),
					vfst.TestContentsString("# contents of .dir/subdir/file\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)

			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user": map[string]any{
					".create": "# contents of .create\n",
					".dir": map[string]any{
						"file": "# contents of .dir/file\n",
						"subdir": map[string]any{
							"file": "# contents of .dir/subdir/file\n",
						},
					},
					".empty": "",
					".executable": &vfst.File{
						Perm:     fs.ModePerm,
						Contents: []byte("# contents of .executable\n"),
					},
					".file": "# contents of .file\n",
					".local": map[string]any{
						"share": map[string]any{
							"chezmoi": &vfst.Dir{Perm: fs.ModePerm},
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
					assert.NoError(t, vfst.NewBuilder().Build(system.UnderlyingFS(), tc.extraRoot))
				}

				s := NewSourceState(
					WithBaseSystem(system),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
					withUserTemplateData(map[string]any{
						"variable": "value",
					}),
				)
				assert.NoError(t, s.Read(ctx, nil))
				requireEvaluateAll(t, s, system)

				destAbsPathInfos := make(map[AbsPath]fs.FileInfo)
				for _, destAbsPath := range tc.destAbsPaths {
					assert.NoError(t, s.AddDestAbsPathInfos(destAbsPathInfos, system, destAbsPath, nil))
				}
				assert.NoError(t, s.Add(system, persistentState, system, destAbsPathInfos, &tc.addOptions))

				vfst.RunTests(t, fileSystem, "", tc.tests...)
			})
		})
	}
}

func TestSourceStateAddInExternal(t *testing.T) {
	buffer := &bytes.Buffer{}
	tarWriterSystem := NewTarWriterSystem(buffer, tar.Header{})
	assert.NoError(t, tarWriterSystem.Mkdir(NewAbsPath("dir"), fs.ModePerm))
	assert.NoError(t, tarWriterSystem.WriteFile(NewAbsPath("dir/file"), []byte("# contents of dir/file\n"), 0o666))
	assert.NoError(t, tarWriterSystem.Close())
	archiveData := buffer.Bytes()

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write(archiveData)
		assert.NoError(t, err)
	}))
	defer httpServer.Close()

	root := map[string]any{
		"/home/user": map[string]any{
			".dir/file2": "# contents of .dir/file2\n",
			".local/share/chezmoi": map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`[".dir"]`,
					`    type = "archive"`,
					`    url = "`+httpServer.URL+`/archive.tar"`,
					`    stripComponents = 1`,
				),
				"dot_dir": &vfst.Dir{Perm: fs.ModePerm},
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
		assert.NoError(t, s.Read(ctx, nil))

		destAbsPath := NewAbsPath("/home/user/.dir/file2")
		fileInfo, err := system.Stat(destAbsPath)
		assert.NoError(t, err)
		destAbsPathInfos := map[AbsPath]fs.FileInfo{
			destAbsPath: fileInfo,
		}
		assert.NoError(t, s.Add(system, persistentState, system, destAbsPathInfos, &AddOptions{
			Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
		}))

		vfst.RunTests(t, fileSystem, "",
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir",
				vfst.TestIsDir(),
				vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
			),
			vfst.TestPath("/home/user/.local/share/chezmoi/dot_dir/file2",
				vfst.TestModeIsRegular(),
				vfst.TestModePerm(0o666&^chezmoitest.Umask),
				vfst.TestContentsString("# contents of .dir/file2\n"),
			),
		)
	})
}

func TestSourceStateApplyAll(t *testing.T) {
	for _, tc := range []struct {
		name               string
		root               any
		sourceStateOptions []SourceStateOption
		tests              []any
	}{
		{
			name: "empty",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": &vfst.Dir{Perm: fs.ModePerm},
				},
			},
		},
		{
			name: "dir",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"dot_dir": &vfst.Dir{Perm: fs.ModePerm},
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
			},
		},
		{
			name: "dir_exact",
			root: map[string]any{
				"/home/user": map[string]any{
					".dir": map[string]any{
						"file": "# contents of .dir/file\n",
					},
					".local/share/chezmoi": map[string]any{
						"exact_dot_dir": &vfst.Dir{Perm: fs.ModePerm},
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.dir",
					vfst.TestIsDir(),
					vfst.TestModePerm(fs.ModePerm&^chezmoitest.Umask),
				),
				vfst.TestPath("/home/user/.dir/file",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "file",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"dot_file": "# contents of .file\n",
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.file",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .file\n"),
				),
			},
		},
		{
			name: "file_remove_empty",
			root: map[string]any{
				"/home/user": map[string]any{
					".empty": "# contents of .empty\n",
					".local/share/chezmoi": map[string]any{
						"dot_empty": "",
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.empty",
					vfst.TestDoesNotExist(),
				),
			},
		},
		{
			name: "file_create_empty",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"empty_dot_empty": "",
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.empty",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContents(nil),
				),
			},
		},
		{
			name: "file_template",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"dot_template.tmpl": "key = {{ .variable }}\n",
					},
				},
			},
			sourceStateOptions: []SourceStateOption{
				withUserTemplateData(map[string]any{
					"variable": "value",
				}),
			},
			tests: []any{
				vfst.TestPath("/home/user/.template",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("key = value\n"),
				),
			},
		},
		{
			name: "create",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"create_dot_create": "# contents of .create\n",
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# contents of .create\n"),
				),
			},
		},
		{
			name: "create_no_replace",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"create_dot_create": "# contents of .create\n",
					},
					".create": "# existing contents of .create\n",
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.create",
					vfst.TestModeIsRegular(),
					vfst.TestModePerm(0o666&^chezmoitest.Umask),
					vfst.TestContentsString("# existing contents of .create\n"),
				),
			},
		},
		{
			name: "symlink",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"symlink_dot_symlink": ".dir/subdir/file\n",
					},
				},
			},
			tests: []any{
				vfst.TestPath("/home/user/.symlink",
					vfst.TestModeType(fs.ModeSymlink),
					vfst.TestSymlinkTarget(filepath.FromSlash(".dir/subdir/file")),
				),
			},
		},
		{
			name: "symlink_template",
			root: map[string]any{
				"/home/user": map[string]any{
					".local/share/chezmoi": map[string]any{
						"symlink_dot_symlink.tmpl": `{{ ".dir/subdir/file" }}` + "\n",
					},
				},
			},
			tests: []any{
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
				assert.NoError(t, s.Read(ctx, nil))
				requireEvaluateAll(t, s, system)
				err := s.applyAll(system, system, persistentState, NewAbsPath("/home/user"), ApplyOptions{
					Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
					Umask:  chezmoitest.Umask,
				})
				assert.NoError(t, err)

				vfst.RunTests(t, fileSystem, "", tc.tests...)
			})
		})
	}
}

func TestSourceStateExecuteTemplateData(t *testing.T) {
	for _, tc := range []struct {
		name        string
		dataStr     string
		expectedStr string
	}{
		{
			name: "line_ending_lf",
			dataStr: "" +
				"unix\n" +
				"\n" +
				"windows\r\n" +
				"\r\n" +
				"# chezmoi:template:line-ending=lf\n",
			expectedStr: chezmoitest.JoinLines(
				"unix",
				"",
				"windows",
				"",
			),
		},
		{
			name: "line_endings_lf",
			dataStr: "" +
				"unix\n" +
				"\n" +
				"windows\r\n" +
				"\r\n" +
				"# chezmoi:template:line-endings=lf\n",
			expectedStr: chezmoitest.JoinLines(
				"unix",
				"",
				"windows",
				"",
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSourceState()
			actual, err := s.ExecuteTemplateData(ExecuteTemplateDataOptions{
				Name: tc.name,
				Data: []byte(tc.dataStr),
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStr, string(actual))
		})
	}
}

func TestSourceStateRead(t *testing.T) {
	for _, tc := range []struct {
		name                string
		root                any
		expectedError       string
		expectedSourceState *SourceState
	}{
		{
			name: "empty",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": &vfst.Dir{Perm: fs.ModePerm},
			},
			expectedSourceState: NewSourceState(),
		},
		{
			name: "dir",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dir": &vfst.Dir{
						Perm: fs.ModePerm &^ chezmoitest.Umask,
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/dir")),
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
				}),
			),
		},
		{
			name: "file",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dot_file": "# contents of .file\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath(".file"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/dot_file")),
						sourceRelPath: NewSourceRelPath("dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of .file\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .file\n"))),
						targetStateEntry: &TargetStateFile{
							contentsFunc:       eagerNoErr([]byte("# contents of .file\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .file\n"))),
							perm:               0o666 &^ chezmoitest.Umask,
						},
					},
				}),
			),
		},
		{
			name: "duplicate_target_file",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dot_file":      "# contents of .file\n",
					"dot_file.tmpl": "# contents of .file\n",
				},
			},
			expectedError: ".file: inconsistent state (/home/user/.local/share/chezmoi/dot_file, /home/user/.local/share/chezmoi/dot_file.tmpl)",
		},
		{
			name: "duplicate_target_dir",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dir": &vfst.Dir{
						Perm: fs.ModePerm &^ chezmoitest.Umask,
					},
					"exact_dir": &vfst.Dir{
						Perm: fs.ModePerm &^ chezmoitest.Umask,
					},
				},
			},
			expectedError: "dir: inconsistent state (/home/user/.local/share/chezmoi/dir, /home/user/.local/share/chezmoi/exact_dir)",
		},
		{
			name: "duplicate_target_script",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"run_script":      "#!/bin/sh\n",
					"run_once_script": "#!/bin/sh\n",
				},
			},
			expectedError: "script: inconsistent state (/home/user/.local/share/chezmoi/run_once_script, /home/user/.local/share/chezmoi/run_script)",
		},
		{
			name: "symlink_with_attr",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".file":               "# contents of .file\n",
					"executable_dot_file": &vfst.Symlink{Target: ".file"},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath(".file"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/executable_dot_file")),
						sourceRelPath: NewSourceRelPath("executable_dot_file"),
						Attr: FileAttr{
							TargetName: ".file",
							Type:       SourceFileTypeFile,
							Executable: true,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of .file\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .file\n"))),
						targetStateEntry: &TargetStateFile{
							contentsFunc:       eagerNoErr([]byte("# contents of .file\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .file\n"))),
							perm:               fs.ModePerm &^ chezmoitest.Umask,
						},
					},
				}),
			),
		},
		{
			name: "symlink_script",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".script":    "# contents of .script\n",
					"run_script": &vfst.Symlink{Target: ".script"},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("script"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/run_script")),
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
							Condition:  ScriptConditionAlways,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of .script\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .script\n"))),
						targetStateEntry: &TargetStateScript{
							name:               NewRelPath("script"),
							contentsFunc:       eagerNoErr([]byte("# contents of .script\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .script\n"))),
							condition:          ScriptConditionAlways,
							sourceAttr: SourceAttr{
								Condition: ScriptConditionAlways,
							},
							sourceRelPath: NewSourceRelPath("run_script"),
						},
					},
				}),
			),
		},
		{
			name: "script",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"run_script": "# contents of script\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("script"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/run_script")),
						sourceRelPath: NewSourceRelPath("run_script"),
						Attr: FileAttr{
							TargetName: "script",
							Type:       SourceFileTypeScript,
							Condition:  ScriptConditionAlways,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of script\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of script\n"))),
						targetStateEntry: &TargetStateScript{
							name:               NewRelPath("script"),
							contentsFunc:       eagerNoErr([]byte("# contents of script\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of script\n"))),
							condition:          ScriptConditionAlways,
							sourceAttr: SourceAttr{
								Condition: ScriptConditionAlways,
							},
							sourceRelPath: NewSourceRelPath("run_script"),
						},
					},
				}),
			),
		},
		{
			name: "symlink",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"symlink_dot_symlink": ".dir/subdir/file",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath(".symlink"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/symlink_dot_symlink")),
						sourceRelPath: NewSourceRelPath("symlink_dot_symlink"),
						Attr: FileAttr{
							TargetName: ".symlink",
							Type:       SourceFileTypeSymlink,
						},
						contentsFunc:       eagerNoErr([]byte(".dir/subdir/file")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte(".dir/subdir/file"))),
						targetStateEntry: &TargetStateSymlink{
							linknameFunc: eagerNoErr(".dir/subdir/file"),
						},
					},
				}),
			),
		},
		{
			name: "file_in_dir",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"dir": map[string]any{
						"file": "# contents of .dir/file\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/dir")),
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/dir/file")),
						sourceRelPath: NewSourceRelPath("dir/file"),
						Attr: FileAttr{
							TargetName: "file",
							Type:       SourceFileTypeFile,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of .dir/file\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .dir/file\n"))),
						targetStateEntry: &TargetStateFile{
							contentsFunc:       eagerNoErr([]byte("# contents of .dir/file\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of .dir/file\n"))),
							perm:               0o666 &^ chezmoitest.Umask,
						},
					},
				}),
			),
		},
		{
			name: "chezmoiignore",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiignore": "README.md\n",
				},
			},
			expectedSourceState: NewSourceState(
				withIgnore(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"README.md": patternSetInclude,
					}),
				),
			),
		},
		{
			name: "chezmoiignore_ignore_file",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiignore": "README.md\n",
					"README.md":      "",
				},
			},
			expectedSourceState: NewSourceState(
				withIgnore(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"README.md": patternSetInclude,
					}),
				),
				withIgnoredRelPathStrs(
					"README.md",
				),
			),
		},
		{
			name: "chezmoiignore_exact_dir",
			root: map[string]any{
				"/home/user/dir": map[string]any{
					"file1": "# contents of dir/file1\n",
					"file2": "# contents of dir/file2\n",
					"file3": "# contents of dir/file3\n",
				},
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiignore": "dir/file3\n",
					"exact_dir": map[string]any{
						"file1": "# contents of dir/file1\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/exact_dir")),
						sourceRelPath: NewSourceRelDirPath("exact_dir"),
						Attr: DirAttr{
							TargetName: "dir",
							Exact:      true,
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file1"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/exact_dir/file1")),
						sourceRelPath: NewSourceRelPath("exact_dir/file1"),
						Attr: FileAttr{
							TargetName: "file1",
							Type:       SourceFileTypeFile,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of dir/file1\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of dir/file1\n"))),
						targetStateEntry: &TargetStateFile{
							contentsFunc:       eagerNoErr([]byte("# contents of dir/file1\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of dir/file1\n"))),
							perm:               0o666 &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file2"): &SourceStateRemove{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/exact_dir")),
						sourceRelPath: NewSourceRelDirPath("exact_dir"),
						targetRelPath: NewRelPath("dir/file2"),
					},
				}),
				withIgnore(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"dir/file3": patternSetInclude,
					}),
				),
				withIgnoredRelPathStrs(
					"dir/file3",
				),
			),
		},
		{
			name: "chezmoiremove",
			root: map[string]any{
				"/home/user/file": "",
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiremove": "file\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("file"): &SourceStateRemove{
						origin:        SourceStateOriginRemove{},
						sourceRelPath: NewSourceRelPath(".chezmoiremove"),
						targetRelPath: NewRelPath("file"),
					},
				}),
				withRemove(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"file": patternSetInclude,
					}),
				),
			),
		},
		{
			name: "chezmoiremove_and_ignore",
			root: map[string]any{
				"/home/user": map[string]any{
					"file1": "",
					"file2": "",
				},
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiignore": "file2\n",
					".chezmoiremove": "file*\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("file1"): &SourceStateRemove{
						origin:        SourceStateOriginRemove{},
						sourceRelPath: NewSourceRelPath(".chezmoiremove"),
						targetRelPath: NewRelPath("file1"),
					},
				}),
				withIgnore(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"file2": patternSetInclude,
					}),
				),
				withIgnoredRelPathStrs(
					"file2",
				),
				withRemove(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"file*": patternSetInclude,
					}),
				),
			),
		},
		{
			name: "chezmoiremove_and_ignore_in_subdir",
			root: map[string]any{
				"/home/user": map[string]any{
					"dir": map[string]any{
						"file1": "",
						"file2": "",
					},
				},
				"/home/user/.local/share/chezmoi": map[string]any{
					"dir/.chezmoiignore": "file2\n",
					"dir/.chezmoiremove": "file*\n",
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/dir")),
						sourceRelPath: NewSourceRelDirPath("dir"),
						Attr: DirAttr{
							TargetName: "dir",
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/file1"): &SourceStateRemove{
						origin:        SourceStateOriginRemove{},
						sourceRelPath: NewSourceRelPath(".chezmoiremove"),
						targetRelPath: NewRelPath("dir/file1"),
					},
				}),
				withIgnore(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"dir/file2": patternSetInclude,
					}),
				),
				withIgnoredRelPathStrs(
					"dir/file2",
				),
				withRemove(
					mustNewPatternSet(t, map[string]patternSetIncludeType{
						"dir/file*": patternSetInclude,
					}),
				),
			),
		},
		{
			name: "external",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					"external_dir": map[string]any{
						"dot_file": "# contents of dir/dot_file\n",
						"subdir": map[string]any{
							"empty_file": "",
						},
						"symlink": &vfst.Symlink{Target: "dot_file"},
					},
				},
			},
			expectedSourceState: NewSourceState(
				withEntries(map[RelPath]SourceStateEntry{
					NewRelPath("dir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/external_dir")),
						sourceRelPath: NewSourceRelDirPath("external_dir"),
						Attr: DirAttr{
							TargetName: "dir",
							External:   true,
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/dot_file"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/external_dir/dot_file")),
						sourceRelPath: NewSourceRelPath("external_dir/dot_file"),
						Attr: FileAttr{
							TargetName: "dot_file",
							Type:       SourceFileTypeFile,
							Empty:      true,
						},
						contentsFunc:       eagerNoErr([]byte("# contents of dir/dot_file\n")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of dir/dot_file\n"))),
						targetStateEntry: &TargetStateFile{
							contentsFunc:       eagerNoErr([]byte("# contents of dir/dot_file\n")),
							contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("# contents of dir/dot_file\n"))),
							empty:              true,
							perm:               0o666 &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/subdir"): &SourceStateDir{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/external_dir/subdir")),
						sourceRelPath: NewSourceRelDirPath("external_dir/subdir"),
						Attr: DirAttr{
							TargetName: "subdir",
							Exact:      true,
						},
						targetStateEntry: &TargetStateDir{
							perm: fs.ModePerm &^ chezmoitest.Umask,
						},
					},
					NewRelPath("dir/subdir/empty_file"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/external_dir/subdir/empty_file")),
						sourceRelPath: NewSourceRelPath("external_dir/subdir/empty_file"),
						Attr: FileAttr{
							TargetName: "empty_file",
							Type:       SourceFileTypeFile,
							Empty:      true,
						},
						contentsFunc:       eagerNoErr([]byte{}),
						contentsSHA256Func: eagerNoErr(sha256.Sum256(nil)),
						targetStateEntry: &TargetStateFile{
							empty:              true,
							perm:               0o666 &^ chezmoitest.Umask,
							contentsFunc:       eagerZeroNoErr[[]byte](),
							contentsSHA256Func: eagerNoErr(sha256.Sum256(nil)),
						},
					},
					NewRelPath("dir/symlink"): &SourceStateFile{
						origin:        SourceStateOriginAbsPath(NewAbsPath("/home/user/.local/share/chezmoi/external_dir/symlink")),
						sourceRelPath: NewSourceRelPath("external_dir/symlink"),
						Attr: FileAttr{
							TargetName: "symlink",
							Type:       SourceFileTypeFile,
						},
						contentsFunc:       eagerNoErr([]byte("dot_file")),
						contentsSHA256Func: eagerNoErr(sha256.Sum256([]byte("dot_file"))),
						targetStateEntry: &TargetStateSymlink{
							linknameFunc: eagerNoErr("dot_file"),
						},
					},
				}),
			),
		},
		{
			name: "chezmoitemplates",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoitemplates": map[string]any{
						"template": "# contents of .chezmoitemplates/template\n",
					},
				},
			},
			expectedSourceState: NewSourceState(
				withTemplates(
					map[string]*Template{
						"template": {
							name: "template",
							template: template.Must(
								template.New("template").
									Option("missingkey=error").
									Parse("# contents of .chezmoitemplates/template\n"),
							),
							options: TemplateOptions{
								Options: []string{"missingkey=error"},
							},
						},
					},
				),
			),
		},
		{
			name: "chezmoiversion",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiversion": "1.2.3\n",
				},
			},
			expectedSourceState: NewSourceState(),
		},
		{
			name: "chezmoiversion_multiple",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiversion": "1.2.3\n",
					"dir": map[string]any{
						".chezmoiversion": "2.3.4\n",
					},
				},
			},
			expectedError: "source state requires chezmoi version 2.3.4 or later, chezmoi is version 1.2.3",
		},
		{
			name: "ignore_dir",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".dir": map[string]any{
						"file": "# contents of .dir/file\n",
					},
				},
			},
			expectedSourceState: NewSourceState(),
		},
		{
			name: "ignore_file",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
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
				assert.NoError(t, err)
				requireEvaluateAll(t, s, system)
				tc.expectedSourceState.destDirAbsPath = NewAbsPath("/home/user")
				tc.expectedSourceState.sourceDirAbsPath = NewAbsPath(
					"/home/user/.local/share/chezmoi",
				)
				requireEvaluateAll(t, tc.expectedSourceState, system)
				s.templateData = nil
				s.version = semver.Version{}
				assert.Equal(t, tc.expectedSourceState, s, assert.Exclude[System]())
			})
		})
	}
}

func TestSourceStateReadExternal(t *testing.T) {
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("data"))
		assert.NoError(t, err)
	}))
	defer httpServer.Close()

	for _, tc := range []struct {
		name              string
		root              any
		expectedExternals map[RelPath][]*External
	}{
		{
			name: "external_yaml",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiexternal.yaml": chezmoitest.JoinLines(
						`file:`,
						`    type: "file"`,
						`    url: "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath][]*External{
				NewRelPath("file"): {
					{
						Type:          "file",
						URL:           httpServer.URL + "/file",
						sourceAbsPath: NewAbsPath("/home/user/.local/share/chezmoi/.chezmoiexternal.yaml"),
					},
				},
			},
		},
		{
			name: "external_toml",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiexternal.toml": chezmoitest.JoinLines(
						`[file]`,
						`    type = "file"`,
						`    url = "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath][]*External{
				NewRelPath("file"): {
					{
						Type:          "file",
						URL:           httpServer.URL + "/file",
						sourceAbsPath: NewAbsPath("/home/user/.local/share/chezmoi/.chezmoiexternal.toml"),
					},
				},
			},
		},
		{
			name: "external_in_subdir",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_dir": map[string]any{
					".chezmoiexternal.yaml": chezmoitest.JoinLines(
						`file:`,
						`    type: "file"`,
						`    url: "`+httpServer.URL+`/file"`,
					),
				},
			},
			expectedExternals: map[RelPath][]*External{
				NewRelPath(".dir/file"): {
					{
						Type:          "file",
						URL:           httpServer.URL + "/file",
						sourceAbsPath: NewAbsPath("/home/user/.local/share/chezmoi/dot_dir/.chezmoiexternal.yaml"),
					},
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
				assert.NoError(t, s.Read(ctx, nil))
				assert.Equal(t, tc.expectedExternals, s.externals)
			})
		})
	}
}

func TestSourceStateReadScriptsConcurrent(t *testing.T) {
	for _, tc := range []struct {
		name string
		root any
	}{
		{
			name: "with_ignore",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
					".chezmoiignore": ".chezmoiscripts/linux/**\n",
					".chezmoiscripts": map[string]any{
						"linux":  manyScripts(1000),
						"darwin": manyScripts(1000),
					},
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

				assert.NoError(t, s.Read(ctx, nil))
			})
		})
	}
}

func TestSourceStateReadExternalCache(t *testing.T) {
	buffer := &bytes.Buffer{}
	tarWriterSystem := NewTarWriterSystem(buffer, tar.Header{})
	assert.NoError(t, tarWriterSystem.WriteFile(NewAbsPath("file"), []byte("# contents of file\n"), 0o666))
	assert.NoError(t, tarWriterSystem.Close())
	archiveData := buffer.Bytes()

	httpRequests := 0
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequests++
		_, err := w.Write(archiveData)
		assert.NoError(t, err)
	}))
	defer httpServer.Close()

	now := time.Now()

	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user/.local/share/chezmoi": map[string]any{
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

		readSourceState := func(refreshExternals RefreshExternals) {
			s := NewSourceState(
				WithBaseSystem(system),
				WithCacheDir(NewAbsPath("/home/user/.cache/chezmoi")),
				WithDestDir(NewAbsPath("/home/user")),
				WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
				WithSystem(system),
			)
			assert.NoError(t, s.Read(ctx, &ReadOptions{
				RefreshExternals: refreshExternals,
				TimeNow: func() time.Time {
					return now
				},
			}))
			assert.Equal(t, map[RelPath][]*External{
				NewRelPath(".dir"): {
					{
						Type:          "archive",
						URL:           httpServer.URL + "/archive.tar",
						RefreshPeriod: Duration(1 * time.Minute),
						sourceAbsPath: NewAbsPath("/home/user/.local/share/chezmoi/.chezmoiexternal.yaml"),
					},
				},
			}, s.externals)
		}

		readSourceState(RefreshExternalsAuto)
		assert.Equal(t, 1, httpRequests)

		now = now.Add(10 * time.Second)
		readSourceState(RefreshExternalsAuto)
		assert.Equal(t, 1, httpRequests)

		now = now.Add(1 * time.Minute)
		readSourceState(RefreshExternalsAuto)
		assert.Equal(t, 2, httpRequests)

		now = now.Add(10 * time.Second)
		readSourceState(RefreshExternalsAlways)
		assert.Equal(t, 3, httpRequests)

		now = now.Add(5 * time.Minute)
		readSourceState(RefreshExternalsNever)
		assert.Equal(t, 3, httpRequests)
	})
}

func TestSourceStateTargetRelPaths(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		root                   any
		expectedTargetRelPaths []RelPath
	}{
		{
			name:                   "empty",
			root:                   nil,
			expectedTargetRelPaths: []RelPath{},
		},
		{
			name: "scripts",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": map[string]any{
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
				assert.NoError(t, s.Read(ctx, nil))
				assert.Equal(t, tc.expectedTargetRelPaths, s.TargetRelPaths())
			})
		})
	}
}

func TestTemplateOptionsParseDirectives(t *testing.T) {
	for _, tc := range []struct {
		name            string
		dataStr         string
		expected        TemplateOptions
		expectedDataStr string
	}{
		{
			name: "empty",
		},
		{
			name:    "unquoted",
			dataStr: "chezmoi:template:left-delimiter=[[ right-delimiter=]]",
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
		},
		{
			name:    "quoted",
			dataStr: `chezmoi:template:left-delimiter="# {{" right-delimiter="}}"`,
			expected: TemplateOptions{
				LeftDelimiter:  "# {{",
				RightDelimiter: "}}",
			},
		},
		{
			name:    "left_only",
			dataStr: "chezmoi:template:left-delimiter=[[",
			expected: TemplateOptions{
				LeftDelimiter: "[[",
			},
		},
		{
			name:    "left_quoted_only",
			dataStr: `chezmoi:template:left-delimiter="# [["`,
			expected: TemplateOptions{
				LeftDelimiter: "# [[",
			},
		},
		{
			name:    "right_quoted_only",
			dataStr: `chezmoi:template:right-delimiter="]]"`,
			expected: TemplateOptions{
				RightDelimiter: "]]",
			},
		},
		{
			name:    "line_with_leading_data",
			dataStr: "# chezmoi:template:left-delimiter=[[ right-delimiter=]]",
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
		},
		{
			name: "line_before",
			dataStr: chezmoitest.JoinLines(
				"# before",
				"# chezmoi:template:left-delimiter=[[ right-delimiter=]]",
			),
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
			expectedDataStr: chezmoitest.JoinLines(
				"# before",
			),
		},
		{
			name: "line_after",
			dataStr: chezmoitest.JoinLines(
				"# chezmoi:template:left-delimiter=[[ right-delimiter=]]",
				"# after",
			),
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
			expectedDataStr: chezmoitest.JoinLines(
				"# after",
			),
		},
		{
			name: "line_before_and_after",
			dataStr: chezmoitest.JoinLines(
				"# before",
				"# chezmoi:template:left-delimiter=[[ right-delimiter=]]",
				"# after",
			),
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
			expectedDataStr: chezmoitest.JoinLines(
				"# before",
				"# after",
			),
		},
		{
			name: "multiple_lines",
			dataStr: chezmoitest.JoinLines(
				"# before",
				"# chezmoi:template:left-delimiter=<<",
				"# during",
				"# chezmoi:template:left-delimiter=[[",
				"# chezmoi:template:right-delimiter=]]",
				"# after",
			),
			expected: TemplateOptions{
				LeftDelimiter:  "[[",
				RightDelimiter: "]]",
			},
			expectedDataStr: chezmoitest.JoinLines(
				"# before",
				"# during",
				"# after",
			),
		},
		{
			name: "duplicate_directives",
			dataStr: chezmoitest.JoinLines(
				"# chezmoi:template:left-delimiter=<<",
				"# chezmoi:template:left-delimiter=[[",
			),
			expected: TemplateOptions{
				LeftDelimiter: "[[",
			},
		},
		{
			name:    "missing_key",
			dataStr: "chezmoi:template:missing-key=zero",
			expected: TemplateOptions{
				Options: []string{"missingkey=zero"},
			},
		},
		{
			name:    "line_ending_crlf",
			dataStr: "chezmoi:template:line-ending=crlf",
			expected: TemplateOptions{
				LineEnding: "\r\n",
			},
		},
		{
			name:    "line_endings_crlf",
			dataStr: "chezmoi:template:line-endings=crlf",
			expected: TemplateOptions{
				LineEnding: "\r\n",
			},
		},
		{
			name:    "line_ending_quoted",
			dataStr: `chezmoi:template:line-ending="\n"`,
			expected: TemplateOptions{
				LineEnding: "\n",
			},
		},
		{
			name:    "line_endings_quoted",
			dataStr: `chezmoi:template:line-endings="\n"`,
			expected: TemplateOptions{
				LineEnding: "\n",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var actual TemplateOptions
			actualData := actual.parseAndRemoveDirectives([]byte(tc.dataStr))
			assert.Equal(t, tc.expected, actual)
			assert.Equal(t, tc.expectedDataStr, string(actualData))
		})
	}
}

func TestSourceStateExternalErrors(t *testing.T) {
	for _, tc := range []struct {
		name        string
		shareDir    map[string]any
		expectedErr string
	}{
		{
			name: "missing_type",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`["dir"]`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "dir: missing external type",
		},
		{
			name: "empty_rel_path",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`[""]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: empty path",
		},
		{
			name: "relative_empty_rel_path",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`["."]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: .: empty relative path",
		},
		{
			name: "parent_root_rel_path",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`[".."]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: ..: relative path in parent",
		},
		{
			name: "relative_parent_root_rel_path",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`["./.."]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: ./..: relative path in parent",
		},
		{
			name: "relative_empty",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`["a/../b/.."]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: a/../b/..: empty relative path",
		},
		{
			name: "relative_parent",
			shareDir: map[string]any{
				".chezmoiexternal.toml": chezmoitest.JoinLines(
					`["a/../b/../.."]`,
					`    type = "file"`,
					`    url = "http://example.com/"`,
				),
			},
			expectedErr: "/home/user/.local/share/chezmoi/.chezmoiexternal.toml: a/../b/../..: relative path in parent",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user/.local/share/chezmoi": tc.shareDir,
			}, func(fileSystem vfs.FS) {
				ctx := context.Background()
				system := NewRealSystem(fileSystem)
				s := NewSourceState(
					WithBaseSystem(system),
					WithCacheDir(NewAbsPath("/home/user/.cache/chezmoi")),
					WithDestDir(NewAbsPath("/home/user")),
					WithSourceDir(NewAbsPath("/home/user/.local/share/chezmoi")),
					WithSystem(system),
				)
				err := s.Read(ctx, nil)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), tc.expectedErr)
			})
		})
	}
}

// applyAll updates targetDirAbsPath in targetSystem to match s.
func (s *SourceState) applyAll(
	targetSystem, destSystem System,
	persistentState PersistentState,
	targetDirAbsPath AbsPath,
	options ApplyOptions,
) error {
	for _, targetRelPath := range s.TargetRelPaths() {
		switch err := s.Apply(targetSystem, destSystem, persistentState, targetDirAbsPath, targetRelPath, options); {
		case errors.Is(err, fs.SkipDir):
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
	err := s.root.forEach(EmptyRelPath, func(targetRelPath RelPath, sourceStateEntry SourceStateEntry) error {
		assert.NoError(t, sourceStateEntry.Evaluate())
		if sourceStateFile, ok := sourceStateEntry.(*SourceStateFile); ok {
			contents, err := sourceStateFile.Contents()
			assert.NoError(t, err)
			contentsSHA256, err := sourceStateFile.ContentsSHA256()
			assert.NoError(t, err)
			assert.Equal(t, sha256.Sum256(contents), contentsSHA256)
		}
		destAbsPath := s.destDirAbsPath.Join(targetRelPath)
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(destSystem, destAbsPath)
		assert.NoError(t, err)
		assert.NoError(t, targetStateEntry.Evaluate())
		switch targetStateEntry := targetStateEntry.(type) {
		case *TargetStateFile:
			contents, err := targetStateEntry.Contents()
			assert.NoError(t, err)
			contentsSHA256, err := targetStateEntry.ContentsSHA256()
			assert.NoError(t, err)
			assert.Equal(t, sha256.Sum256(contents), contentsSHA256)
		case *TargetStateScript:
			contents, err := targetStateEntry.Contents()
			assert.NoError(t, err)
			contentsSHA256, err := targetStateEntry.ContentsSHA256()
			assert.NoError(t, err)
			assert.Equal(t, sha256.Sum256(contents), contentsSHA256)
		}
		return nil
	})
	assert.NoError(t, err)
}

func withEntries(sourceEntries map[RelPath]SourceStateEntry) SourceStateOption {
	return func(s *SourceState) {
		s.root = sourceStateEntryTreeNode{}
		for targetRelPath, sourceStateEntry := range sourceEntries {
			s.root.set(targetRelPath, sourceStateEntry)
		}
	}
}

func withIgnore(ignore *patternSet) SourceStateOption {
	return func(s *SourceState) {
		s.ignore = ignore
	}
}

func withIgnoredRelPathStrs(relPathStrs ...string) SourceStateOption {
	return func(s *SourceState) {
		for _, relPathStr := range relPathStrs {
			s.ignoredRelPaths.Add(NewRelPath(relPathStr))
		}
	}
}

func withRemove(remove *patternSet) SourceStateOption {
	return func(s *SourceState) {
		s.remove = remove
	}
}

// withUserTemplateData adds template data.
func withUserTemplateData(templateData map[string]any) SourceStateOption {
	return func(s *SourceState) {
		RecursiveMerge(s.userTemplateData, templateData)
	}
}

func withTemplates(templates map[string]*Template) SourceStateOption {
	return func(s *SourceState) {
		s.templates = templates
	}
}

func manyScripts(amount int) map[string]any {
	scripts := map[string]any{}
	for i := range amount {
		scripts[fmt.Sprintf("run_onchange_before_%d.sh", i)] = ""
	}
	return scripts
}
