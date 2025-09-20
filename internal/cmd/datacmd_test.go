package cmd

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/internal/chezmoi"
	"chezmoi.io/chezmoi/internal/chezmoitest"
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
					`  "mode": "symlink",`,
					`  "sourceDir": "/tmp/source",`,
					`  "encryption": "age",`,
					`  "age": {`,
					`    "args": [`,
					`      "arg"`,
					`    ],`,
					`    "identity": "/my-age-identity"`,
					`  },`,
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
					`mode: symlink`,
					`sourceDir: /tmp/source`,
					`encryption: age`,
					`age:`,
					`  args:`,
					`  - arg`,
					`  identity: /my-age-identity`,
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
				stdout := strings.Builder{}
				config := newTestConfig(t, fileSystem, withStdout(&stdout))
				assert.NoError(t, config.execute(args))

				var data struct {
					Chezmoi struct {
						Config struct {
							Age struct {
								Args     []string `json:"args"     yaml:"args"`
								Identity string   `json:"identity" yaml:"identity"`
							} `json:"age"  yaml:"age"`
							Mode string `json:"mode" yaml:"mode"`
						} `json:"config"    yaml:"config"`
						SourceDir string `json:"sourceDir" yaml:"sourceDir"`
					} `json:"chezmoi" yaml:"chezmoi"`
					Test bool `json:"test"    yaml:"test"`
				}
				assert.NoError(t, tc.format.Unmarshal([]byte(stdout.String()), &data))
				assert.Equal(t, []string{"arg"}, data.Chezmoi.Config.Age.Args)
				normalizedAgeIdentity, err := chezmoi.NormalizePath("/my-age-identity")
				assert.NoError(t, err)
				assert.Equal(t, normalizedAgeIdentity.String(), data.Chezmoi.Config.Age.Identity)
				assert.Equal(t, "symlink", data.Chezmoi.Config.Mode)
				normalizedSourceDir, err := chezmoi.NormalizePath("/tmp/source")
				assert.NoError(t, err)
				assert.Equal(t, normalizedSourceDir.String(), data.Chezmoi.SourceDir)
				assert.True(t, data.Test)
			})
		})
	}
}
