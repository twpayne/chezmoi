// +build !windows

package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
)

func TestDiffDoesNotRunScript(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chezmoi")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()
	fs := vfs.NewPathFS(vfs.OSFS, tempDir)
	require.NoError(t, vfst.NewBuilder().Build(
		fs,
		map[string]interface{}{
			"/home/user/.local/share/chezmoi/run_true": "#!/bin/sh\necho foo >>" + filepath.Join(tempDir, "evidence") + "\n",
		},
	))
	c := newConfig(
		withTestFS(fs),
		withTestUser("user"),
	)
	assert.NoError(t, c.runDiffCmd(nil, nil))
	vfst.RunTests(t, vfs.OSFS, "",
		vfst.TestPath(filepath.Join(tempDir, "evidence"),
			vfst.TestDoesNotExist,
		),
	)
}
