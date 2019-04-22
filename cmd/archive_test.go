package cmd

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestArchiveCmd(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user/.local/share/chezmoi/dir/file":        "contents",
		"/home/user/.local/share/chezmoi/symlink_symlink": "target",
	})
	require.NoError(t, err)
	defer cleanup()
	stdout := &bytes.Buffer{}
	c := &Config{
		SourceDir: "/home/user/.local/share/chezmoi",
		Umask:     022,
		stdout:    stdout,
	}
	assert.NoError(t, c.runArchiveCmd(fs, nil))
	r := tar.NewReader(stdout)

	h, err := r.Next()
	assert.NoError(t, err)
	assert.Equal(t, "dir", h.Name)

	h, err = r.Next()
	assert.NoError(t, err)
	assert.Equal(t, "dir/file", h.Name)
	data, err := ioutil.ReadAll(r)
	assert.NoError(t, err)
	assert.Equal(t, []byte("contents"), data)

	h, err = r.Next()
	assert.NoError(t, err)
	assert.Equal(t, "symlink", h.Name)
	assert.Equal(t, "target", h.Linkname)

	_, err = r.Next()
	assert.Equal(t, err, io.EOF)
}
