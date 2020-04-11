package cmd

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestManagedCmd(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.local/share/chezmoi/dir/file1":        "contents",
		"/home/user/.local/share/chezmoi/dir/subdir/file2": "contents",
		"/home/user/.local/share/chezmoi/symlink_symlink":  "target",
	})
	require.NoError(t, err)
	defer cleanup()
	stdout := &bytes.Buffer{}
	c := newTestConfig(
		fs,
		withStdout(stdout),
	)
	assert.NoError(t, c.runManagedCmd(nil, nil))
	fmt.Print(stdout.String())
	actual := strings.TrimSuffix(strings.Replace(stdout.String(), "\r\n", "\n", -1), "\n")
	actualSlice := strings.Split(actual, "\n")
	assert.Greater(t, len(actualSlice), 0)
	expected := [3]string{
		filepath.Join(filepath.VolumeName(actualSlice[0]), "/", "home", "user", "dir", "file1"),
		filepath.Join(filepath.VolumeName(actualSlice[0]), "/", "home", "user", "dir", "subdir", "file2"),
		filepath.Join(filepath.VolumeName(actualSlice[0]), "/", "home", "user", "symlink"),
	}
	assert.ElementsMatch(t, expected, actualSlice)
}
