package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-vfs/v4/vfst"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestVerifyCmd(t *testing.T) {
	for _, tc := range []struct {
		name        string
		root        any
		expectedErr error
	}{
		{
			name: "empty",
			root: map[string]any{
				"/home/user/.local/share/chezmoi": &vfst.Dir{
					Perm: 0o777 &^ chezmoitest.Umask,
				},
			},
		},
		{
			name: "file",
			root: map[string]any{
				"/home/user": map[string]any{
					".bashrc": &vfst.File{
						Contents: []byte("# contents of .bashrc\n"),
						Perm:     0o666 &^ chezmoitest.Umask,
					},
					".local/share/chezmoi/dot_bashrc": "# contents of .bashrc\n",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				assert.Equal(t, tc.expectedErr, newTestConfig(t, fileSystem).execute([]string{"verify"}))
			})
		})
	}
}
