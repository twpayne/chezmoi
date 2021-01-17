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
		format string
		root   map[string]interface{}
	}{
		{
			format: "json",
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
			format: "toml",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.toml": chezmoitest.JoinLines(
					`sourceDir = "/tmp/source"`,
					`[data]`,
					`  test = true`,
				),
			},
		},
		{
			format: "yaml",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.yaml": chezmoitest.JoinLines(
					`sourceDir: /tmp/source`,
					`data:`,
					`  test: true`,
				),
			},
		},
	} {
		t.Run(tc.format, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fs vfs.FS) {
				args := []string{
					"data",
					"--format", tc.format,
				}
				c := newTestConfig(t, fs)
				var sb strings.Builder
				c.stdout = &sb
				require.NoError(t, c.execute(args))

				var data struct {
					Chezmoi struct {
						SourceDir string `json:"sourceDir" toml:"sourceDir" yaml:"sourceDir"`
					} `json:"chezmoi" toml:"chezmoi" yaml:"chezmoi"`
					Test bool `json:"test" toml:"test" yaml:"test"`
				}
				assert.NoError(t, chezmoi.Formats[tc.format].Unmarshal([]byte(sb.String()), &data))
				normalizedSourceDir, err := chezmoi.NormalizePath("/tmp/source")
				require.NoError(t, err)
				assert.Equal(t, string(normalizedSourceDir), data.Chezmoi.SourceDir)
				assert.True(t, data.Test)
			})
		})
	}
}
