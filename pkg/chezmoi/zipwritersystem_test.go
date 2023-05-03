package chezmoi

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/coreos/go-semver/semver"
	"github.com/klauspost/compress/zip"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

var _ System = &ZIPWriterSystem{}

func TestZIPWriterSystem(t *testing.T) {
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
		assert.NoError(t, s.Read(ctx, nil))
		requireEvaluateAll(t, s, system)

		b := &bytes.Buffer{}
		zipWriterSystem := NewZIPWriterSystem(b, time.Now().UTC())
		persistentState := NewMockPersistentState()
		assert.NoError(t, s.applyAll(zipWriterSystem, system, persistentState, EmptyAbsPath, ApplyOptions{
			Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
		}))
		assert.NoError(t, zipWriterSystem.Close())

		r, err := zip.NewReader(bytes.NewReader(b.Bytes()), int64(b.Len()))
		assert.NoError(t, err)
		expectedFiles := []struct {
			name     string
			method   uint16
			mode     fs.FileMode
			contents []byte
		}{
			{
				name: ".dir",
				mode: (fs.ModeDir | fs.ModePerm) &^ chezmoitest.Umask,
			},
			{
				name:     ".dir/file",
				method:   zip.Deflate,
				mode:     0o666 &^ chezmoitest.Umask,
				contents: []byte("# contents of .dir/file\n"),
			},
			{
				name:     "script",
				method:   zip.Deflate,
				mode:     0o700 &^ chezmoitest.Umask,
				contents: []byte("# contents of script\n"),
			},
			{
				name:     "symlink",
				mode:     fs.ModeSymlink,
				contents: []byte(".dir/subdir/file"),
			},
		}
		assert.Equal(t, len(expectedFiles), len(r.File))
		for i, expectedFile := range expectedFiles {
			t.Run(expectedFile.name, func(t *testing.T) {
				actualFile := r.File[i]
				assert.Equal(t, expectedFile.name, actualFile.Name)
				assert.Equal(t, expectedFile.method, actualFile.Method)
				assert.Equal(t, expectedFile.mode, actualFile.Mode())
				if expectedFile.contents != nil {
					rc, err := actualFile.Open()
					assert.NoError(t, err)
					actualContents, err := io.ReadAll(rc)
					assert.NoError(t, err)
					assert.Equal(t, expectedFile.contents, actualContents)
				}
			})
		}
	})
}
