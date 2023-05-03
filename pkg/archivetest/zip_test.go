package archivetest

import (
	"bytes"
	"io/fs"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/klauspost/compress/zip"
	"github.com/stretchr/testify/require"
)

func TestNewZip(t *testing.T) {
	data, err := NewZip(map[string]any{
		"dir": map[string]any{
			"file1": "# contents of dir/file1\n",
			"file2": []byte("# contents of dir/file2\n"),
			"subdir": &Dir{
				Perm: 0o700,
				Entries: map[string]any{
					"file": &File{
						Perm:     fs.ModePerm,
						Contents: []byte("# contents of dir/subdir/file\n"),
					},
					"symlink": &Symlink{
						Target: "file",
					},
				},
			},
		},
	})
	require.NoError(t, err)

	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	fileIndex := 0
	nextFile := func() *zip.File {
		require.LessOrEqual(t, fileIndex, len(zipReader.File))
		zipFile := zipReader.File[fileIndex]
		fileIndex++
		return zipFile
	}

	zipFile := nextFile()
	assert.Equal(t, "dir", zipFile.Name)
	assert.Equal(t, fs.ModeDir, zipFile.FileInfo().Mode().Type())
	assert.Equal(t, fs.ModePerm, zipFile.FileInfo().Mode().Perm())

	zipFile = nextFile()
	assert.Equal(t, "dir/file1", zipFile.Name)
	assert.Equal(t, fs.FileMode(0), zipFile.FileInfo().Mode().Type())
	assert.Equal(t, fs.FileMode(0o666), zipFile.FileInfo().Mode().Perm())
	assert.Equal(t, uint64(len("# contents of dir/file1\n")), zipFile.UncompressedSize64)

	zipFile = nextFile()
	assert.Equal(t, "dir/file2", zipFile.Name)
	assert.Equal(t, fs.FileMode(0), zipFile.FileInfo().Mode().Type())
	assert.Equal(t, fs.FileMode(0o666), zipFile.FileInfo().Mode().Perm())
	assert.Equal(t, uint64(len("# contents of dir/file2\n")), zipFile.UncompressedSize64)

	zipFile = nextFile()
	assert.Equal(t, "dir/subdir", zipFile.Name)
	assert.Equal(t, fs.ModeDir, zipFile.FileInfo().Mode().Type())
	assert.Equal(t, fs.FileMode(0o700), zipFile.FileInfo().Mode().Perm())

	zipFile = nextFile()
	assert.Equal(t, "dir/subdir/file", zipFile.Name)
	assert.Equal(t, fs.FileMode(0), zipFile.FileInfo().Mode().Type())
	assert.Equal(t, fs.ModePerm, zipFile.FileInfo().Mode().Perm())
	assert.Equal(t, uint64(len("# contents of dir/subdir/file\n")), zipFile.UncompressedSize64)

	zipFile = nextFile()
	assert.Equal(t, "dir/subdir/symlink", zipFile.Name)
	assert.Equal(t, fs.ModeSymlink, zipFile.FileInfo().Mode().Type())
	assert.Equal(t, uint64(len("file")), zipFile.UncompressedSize64)

	assert.Equal(t, fileIndex, len(zipReader.File))
}
