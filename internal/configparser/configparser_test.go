package configparser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestExtensions(t *testing.T) {
	expectedExtensions := []string{
		".json",
		".toml",
		".yaml",
		".yml",
	}
	assert.Equal(t, expectedExtensions, Extensions())
}

func TestParseConfig(t *testing.T) {
	type Config struct {
		Format string `json:"format" toml:"format" yaml:"format"`
	}
	for _, tc := range []struct {
		name                     string
		root                     interface{}
		filename                 string
		expectedFindConfigError  bool
		expectedParseConfigError bool
		expectedConfig           *Config
	}{
		{
			name:           "no_config",
			root:           nil,
			filename:       "/home/user/.config/chezmoi/chezmoi",
			expectedConfig: &Config{},
		},
		{
			name: "json_file",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.json": `{"format":"json"}`,
			},
			filename: "/home/user/.config/chezmoi/chezmoi.json",
			expectedConfig: &Config{
				Format: "json",
			},
		},
		{
			name: "json",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.json": `{"format":"json"}`,
			},
			filename: "/home/user/.config/chezmoi/chezmoi",
			expectedConfig: &Config{
				Format: "json",
			},
		},
		{
			name: "toml",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.toml": "format = \"toml\"\n",
			},
			filename: "/home/user/.config/chezmoi/chezmoi",
			expectedConfig: &Config{
				Format: "toml",
			},
		},
		{
			name: "yaml",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.yaml": "format: yaml\n",
			},
			filename: "/home/user/.config/chezmoi/chezmoi",
			expectedConfig: &Config{
				Format: "yaml",
			},
		},
		{
			name: "yml",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.yml": "format: yaml\n",
			},
			filename: "/home/user/.config/chezmoi/chezmoi",
			expectedConfig: &Config{
				Format: "yaml",
			},
		},
		{
			name:           "file_does_not_exist",
			filename:       "/home/user/.config/chezmoi/chezmoi.json",
			expectedConfig: &Config{},
		},
		{
			name: "json_file_is_directory",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.json/foo": "bar",
			},
			filename:       "/home/user/.config/chezmoi/chezmoi.json",
			expectedConfig: &Config{},
		},
		{
			name: "unsupported_format",
			root: map[string]interface{}{
				"/home/user/.config/chezmoi/chezmoi.properties": `<?xml version="1.0" encoding="UTF-8"?>`,
			},
			filename:                 "/home/user/.config/chezmoi/chezmoi.properties",
			expectedParseConfigError: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			configFileName, err := FindConfig(fs, tc.filename)
			if tc.expectedFindConfigError {
				require.Error(t, err)
				return
			}
			actualConfig := &Config{}
			if configFileName != "" {
				configFile, err := fs.Open(configFileName)
				require.NoError(t, err)
				defer configFile.Close()
				err = ParseConfig(configFile, actualConfig)
				if tc.expectedParseConfigError {
					require.Error(t, err)
					return
				}
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedConfig, actualConfig)
		})
	}
}
