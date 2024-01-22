package chezmoi

import (
	"context"
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

var _ System = &DumpSystem{}

func TestDumpSystem(t *testing.T) {
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

		dumpSystem := NewDumpSystem()
		persistentState := NewMockPersistentState()
		err := s.applyAll(dumpSystem, system, persistentState, EmptyAbsPath, ApplyOptions{
			Filter: NewEntryTypeFilter(EntryTypesAll, EntryTypesNone),
		})
		assert.NoError(t, err)
		expectedData := map[string]any{
			".dir": &dirData{
				Type: dataTypeDir,
				Name: NewAbsPath(".dir"),
				Perm: fs.ModePerm &^ chezmoitest.Umask,
			},
			".dir/file": &fileData{
				Type:     dataTypeFile,
				Name:     NewAbsPath(".dir/file"),
				Contents: "# contents of .dir/file\n",
				Perm:     0o666 &^ chezmoitest.Umask,
			},
			"script": &scriptData{
				Type:      dataTypeScript,
				Name:      NewAbsPath("script"),
				Contents:  "# contents of script\n",
				Condition: "always",
			},
			"symlink": &symlinkData{
				Type:     dataTypeSymlink,
				Name:     NewAbsPath("symlink"),
				Linkname: ".dir/subdir/file",
			},
		}
		actualData := dumpSystem.Data()
		assert.Equal(t, expectedData, actualData.(map[string]any))
	})
}
