package chezmoi

import (
	"context"
	"io/fs"
	"sort"
	"sync"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-vfs/v4/vfst"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestConcurrentWalkSourceDir(t *testing.T) {
	sourceDirAbsPath := NewAbsPath("/home/user/.local/share/chezmoi")
	root := map[string]any{
		sourceDirAbsPath.String(): map[string]any{
			".chezmoiversion": "# contents of .chezmoiversion\n",
			"dot_dir/file":    "# contents of .dir/file\n",
		},
	}
	expectedSourceAbsPaths := AbsPaths{
		sourceDirAbsPath.JoinString(".chezmoiversion"),
		sourceDirAbsPath.JoinString("dot_dir"),
		sourceDirAbsPath.JoinString("dot_dir/file"),
	}

	var actualSourceAbsPaths AbsPaths
	chezmoitest.WithTestFS(t, root, func(fileSystem vfs.FS) {
		ctx := context.Background()
		system := NewRealSystem(fileSystem)
		var mutex sync.Mutex
		walkFunc := func(ctx context.Context, sourceAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
			mutex.Lock()
			actualSourceAbsPaths = append(actualSourceAbsPaths, sourceAbsPath)
			mutex.Unlock()
			return nil
		}
		assert.NoError(t, concurrentWalkSourceDir(ctx, system, sourceDirAbsPath, walkFunc))
	})
	sort.Sort(actualSourceAbsPaths)
	assert.Equal(t, expectedSourceAbsPaths, actualSourceAbsPaths)
}

func TestWalkSourceDir(t *testing.T) {
	sourceDirAbsPath := NewAbsPath("/home/user/.local/share/chezmoi")
	root := map[string]any{
		sourceDirAbsPath.String(): map[string]any{
			".chezmoi.toml.tmpl":    "",
			".chezmoidata.json":     "",
			".chezmoidata.toml":     "",
			".chezmoidata.yaml":     "",
			".chezmoiexternal.yaml": "",
			".chezmoiignore":        "",
			".chezmoiremove":        "",
			".chezmoitemplates":     &vfst.Dir{Perm: fs.ModePerm},
			".chezmoiversion":       "",
			"dot_file":              "",
		},
	}
	expectedSourceDirAbsPaths := []AbsPath{
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

	var actualSourceDirAbsPaths []AbsPath
	chezmoitest.WithTestFS(t, root, func(fileSystem vfs.FS) {
		system := NewRealSystem(fileSystem)
		err := WalkSourceDir(system, sourceDirAbsPath, func(absPath AbsPath, fileInfo fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			actualSourceDirAbsPaths = append(actualSourceDirAbsPaths, absPath)
			return nil
		})
		assert.NoError(t, err)
	})
	assert.Equal(t, expectedSourceDirAbsPaths, actualSourceDirAbsPaths)
}
