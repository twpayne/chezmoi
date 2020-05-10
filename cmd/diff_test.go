package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestIssue740(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"/home/user": map[string]interface{}{
			"dir": map[string]interface{}{
				"foo": "foo",
			},
			".local/share/chezmoi": map[string]interface{}{
				"exact_dir": map[string]interface{}{
					"foo": "foo",
					"bar": "bar",
				},
			},
		},
	})
	require.NoError(t, err)
	defer cleanup()

	c := newTestConfig(fs)
	c.Diff.Format = "git"
	assert.NoError(t, c.runDiffCmd(nil, nil))
}
