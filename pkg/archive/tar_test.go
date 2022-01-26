package archive

import (
	"archive/tar"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTarFile(t *testing.T) {
	data, err := NewTar(map[string]interface{}{
		"file": "# contents of .file\n",
	})
	require.NoError(t, err)

	tarReader := tar.NewReader(bytes.NewBuffer(data))

	header, err := tarReader.Next()
	require.NoError(t, err)
	assert.Equal(t, int(tar.TypeReg), int(header.Typeflag))
	assert.Equal(t, "file", header.Name)
	assert.Equal(t, len("# contents of .file\n"), int(header.Size))
	assert.Equal(t, 0o666, int(header.Mode))
}
