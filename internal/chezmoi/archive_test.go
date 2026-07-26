package chezmoi

import (
	"io"
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/v2/internal/archivetest"
)

func TestWalkArchive(t *testing.T) {
	nestedRoot := map[string]any{
		"dir1": map[string]any{
			"subdir1": map[string]any{
				"file1": "",
				"file2": "",
			},
			"subdir2": map[string]any{
				"file1": "",
				"file2": "",
			},
		},
		"dir2": map[string]any{
			"subdir1": map[string]any{
				"file1": "",
				"file2": "",
			},
			"subdir2": map[string]any{
				"file1": "",
				"file2": "",
			},
		},
		"file1":    "",
		"file2":    "",
		"symlink1": &archivetest.Symlink{Target: "file1"},
		"symlink2": &archivetest.Symlink{Target: "file2"},
	}
	flatRoot := map[string]any{
		"dir1/subdir1/file1": "",
		"dir1/subdir1/file2": "",
		"dir1/subdir2/file1": "",
		"dir1/subdir2/file2": "",
		"dir2/subdir1/file1": "",
		"dir2/subdir1/file2": "",
		"dir2/subdir2/file1": "",
		"dir2/subdir2/file2": "",
		"file1":              "",
		"file2":              "",
		"symlink1":           &archivetest.Symlink{Target: "file1"},
		"symlink2":           &archivetest.Symlink{Target: "file2"},
	}
	for _, tc := range []struct {
		name          string
		root          map[string]any
		dataFunc      func(map[string]any) ([]byte, error)
		archiveFormat ArchiveFormat
	}{
		{
			name:          "tar",
			root:          nestedRoot,
			dataFunc:      archivetest.NewTar,
			archiveFormat: ArchiveFormatTar,
		},
		{
			name:          "zip",
			root:          nestedRoot,
			dataFunc:      archivetest.NewZip,
			archiveFormat: ArchiveFormatZip,
		},
		{
			name:          "zip-flat",
			root:          flatRoot,
			dataFunc:      archivetest.NewZip,
			archiveFormat: ArchiveFormatZip,
		},
		{
			name:          "tar-flat",
			root:          flatRoot,
			dataFunc:      archivetest.NewTar,
			archiveFormat: ArchiveFormatTar,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := tc.dataFunc(tc.root)
			assert.NoError(t, err)

			expectedNames := []RelPath{
				NewRelPath("dir1"),
				NewRelPath("dir1/subdir1"),
				NewRelPath("dir1/subdir1/file1"),
				NewRelPath("dir1/subdir1/file2"),
				NewRelPath("dir1/subdir2"),
				NewRelPath("dir2"),
				NewRelPath("file1"),
				NewRelPath("file2"),
				NewRelPath("symlink1"),
			}

			var actualNames []RelPath
			walkArchiveFunc := func(name RelPath, info fs.FileInfo, r io.Reader, linkname string) error {
				actualNames = append(actualNames, name)
				switch name {
				case NewRelPath("dir1/subdir2"):
					return fs.SkipDir
				case NewRelPath("dir2"):
					return fs.SkipDir
				case NewRelPath("symlink1"):
					return fs.SkipAll
				default:
					return nil
				}
			}
			assert.NoError(t, WalkArchive(data, tc.archiveFormat, walkArchiveFunc))
			assert.Equal(t, expectedNames, actualNames)
		})
	}
}
