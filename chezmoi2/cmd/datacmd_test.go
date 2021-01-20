package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

func TestDataCmd(t *testing.T) {
	for _, tc := range []struct {
		format chezmoi.Format
		root   map[string]interface{}
	}{
		{
			format: chezmoi.JSONFormat,
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.json": chezmoitest.JoinLines(
					`{`,
					`  "sourceDir": "/tmp/source",`,
					`  "data": {`,
					`    "test": true`,
					`  }`,
					`}`,
				),
			},
		},
		{
			format: chezmoi.YAMLFormat,
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.yaml": chezmoitest.JoinLines(
					`sourceDir: /tmp/source`,
					`data:`,
					`  test: true`,
				),
			},
		},
	} {
		t.Run(tc.format.Name(), func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				args := []string{
					"data",
					"--format", tc.format.Name(),
				}
				c := newTestConfig(t, fs)
				var sb strings.Builder
				c.stdout = &sb
				require.NoError(t, c.execute(args))

				var data struct {
					Chezmoi struct {
						SourceDir string `json:"sourceDir" yaml:"sourceDir"`
					} `json:"chezmoi" yaml:"chezmoi"`
					Test bool `json:"test" yaml:"test"`
				}
				assert.NoError(t, tc.format.Unmarshal([]byte(sb.String()), &data))
				normalizedSourceDir, err := chezmoi.NormalizePath("/tmp/source")
				require.NoError(t, err)
				assert.Equal(t, string(normalizedSourceDir), data.Chezmoi.SourceDir)
				assert.True(t, data.Test)
			})
		})
	}
}
