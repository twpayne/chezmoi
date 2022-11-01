package chezmoi

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"io/fs"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

var _ System = &TarWriterSystem{}

func TestTarWriterSystem(t *testing.T) {
	chezmoitest.WithTestFS(t, map[string]any{
		"/home/user/.local/share/chezmoi": map[string]any{
			".chezmoiignore":  "README.md\n",
			".chezmoiremove":  "*.txt\n",
			".chezmoiversion": "1.2.3\n",
			".chezmoitemplates": map[string]any{
				"template": "# contents of .chezmoitemplates/template\n",
			},
			"README.md": "",
			"dot_dir": map[string]any{
				"file": "# contents of .dir/file\n",
			},
			"run_script":      "# contents of script\n",
			"symlink_symlink": ".dir/subdir/file\n",
		},
	}, func(fileSystem vfs.FS) {
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
		require.NoError(t, s.Read(ctx, nil))
		requireEvaluateAll(t, s, system)

		b := &bytes.Buffer{}
		tarWriterSystem := NewTarWriterSystem(b, tar.Header{})
		persistentState := NewMockPersistentState()
		require.NoError(t, s.applyAll(tarWriterSystem, system, persistentState, EmptyAbsPath, ApplyOptions{
			Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
		}))
		require.NoError(t, tarWriterSystem.Close())

		r := tar.NewReader(b)
		for _, tc := range []struct {
			expectedTypeflag byte
			expectedName     string
			expectedMode     int64
			expectedLinkname string
			expectedContents []byte
		}{
			{
				expectedTypeflag: tar.TypeDir,
				expectedName:     ".dir/",
				expectedMode:     int64(fs.ModePerm &^ chezmoitest.Umask),
			},
			{
				expectedTypeflag: tar.TypeReg,
				expectedName:     ".dir/file",
				expectedContents: []byte("# contents of .dir/file\n"),
				expectedMode:     0o666 &^ int64(chezmoitest.Umask),
			},
			{
				expectedTypeflag: tar.TypeReg,
				expectedName:     "script",
				expectedContents: []byte("# contents of script\n"),
				expectedMode:     0o700,
			},
			{
				expectedTypeflag: tar.TypeSymlink,
				expectedName:     "symlink",
				expectedLinkname: ".dir/subdir/file",
			},
		} {
			t.Run(tc.expectedName, func(t *testing.T) {
				header, err := r.Next()
				require.NoError(t, err)
				assert.Equal(t, tc.expectedTypeflag, header.Typeflag)
				assert.Equal(t, tc.expectedName, header.Name)
				assert.Equal(t, tc.expectedMode, header.Mode)
				assert.Equal(t, tc.expectedLinkname, header.Linkname)
				assert.Equal(t, int64(len(tc.expectedContents)), header.Size)
				if tc.expectedContents != nil {
					actualContents, err := io.ReadAll(r)
					require.NoError(t, err)
					assert.Equal(t, tc.expectedContents, actualContents)
				}
			})
		}
		_, err := r.Next()
		assert.Equal(t, io.EOF, err)
	})
}
