package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestDumpCmd(t *testing.T) {
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
		dump: dumpCmdConfig{
			format:    "json",
			recursive: true,
		},
		stdout: stdout,
	}
	assert.NoError(t, c.runDumpCmd(fs, nil))
	fmt.Println(stdout.String())
	var actual interface{}
	assert.NoError(t, json.NewDecoder(stdout).Decode(&actual))
	expected := []interface{}{
		map[string]interface{}{
			"type":       "dir",
			"sourcePath": "/home/user/.local/share/chezmoi/dir",
			"targetPath": "dir",
			"exact":      false,
			"perm":       float64(0777),
			"entries": []interface{}{
				map[string]interface{}{
					"type":       "file",
					"sourcePath": "/home/user/.local/share/chezmoi/dir/file",
					"targetPath": "dir/file",
					"empty":      false,
					"encrypted":  false,
					"perm":       float64(0666),
					"template":   false,
					"contents":   "contents",
				},
			},
		},
		map[string]interface{}{
			"type":       "symlink",
			"sourcePath": "/home/user/.local/share/chezmoi/symlink_symlink",
			"targetPath": "symlink",
			"template":   false,
			"linkname":   "target",
		},
	}
	assert.Equal(t, expected, actual)
}
