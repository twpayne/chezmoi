package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestDataCmd(t *testing.T) {
	for _, tc := range []struct {
		format chezmoi.Format
		root   map[string]any
	}{
		{
			format: chezmoi.FormatJSON,
			root: map[string]any{
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
			format: chezmoi.FormatYAML,
			root: map[string]any{
				"/home/user/.config/chezmoi/chezmoi.yaml": chezmoitest.JoinLines(
					`sourceDir: /tmp/source`,
					`data:`,
					`  test: true`,
				),
			},
		},
	} {
		t.Run(tc.format.Name(), func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				args := []string{
					"data",
					"--format", tc.format.Name(),
				}
				config := newTestConfig(t, fileSystem)
				builder := strings.Builder{}
				config.stdout = &builder
				assert.NoError(t, config.execute(args))

				var data struct {
					Chezmoi struct {
						SourceDir string `json:"sourceDir" yaml:"sourceDir"`
					} `json:"chezmoi" yaml:"chezmoi"`
					Test bool `json:"test" yaml:"test"`
				}
				assert.NoError(t, tc.format.Unmarshal([]byte(builder.String()), &data))
				normalizedSourceDir, err := chezmoi.NormalizePath("/tmp/source")
				assert.NoError(t, err)
				assert.Equal(t, normalizedSourceDir.String(), data.Chezmoi.SourceDir)
				assert.True(t, data.Test)
			})
		})
	}
}
