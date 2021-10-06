package cmd

import (
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v4"
	xdg "github.com/twpayne/go-xdg/v6"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestAddTemplateFuncPanic(t *testing.T) {
	chezmoitest.WithTestFS(t, nil, func(fileSystem vfs.FS) {
		config := newTestConfig(t, fileSystem)
		assert.NotPanics(t, func() {
			config.addTemplateFunc("func", nil)
		})
		assert.Panics(t, func() {
			config.addTemplateFunc("func", nil)
		})
	})
}

func TestParseConfig(t *testing.T) {
	for _, tc := range []struct {
		name          string
		filename      string
		contents      string
		expectedColor bool
	}{
		{
			name:     "json_bool",
			filename: "chezmoi.json",
			contents: chezmoitest.JoinLines(
				`{`,
				`  "color":true`,
				`}`,
			),
			expectedColor: true,
		},
		{
			name:     "json_string",
			filename: "chezmoi.json",
			contents: chezmoitest.JoinLines(
				`{`,
				`  "color":"on"`,
				`}`,
			),
			expectedColor: true,
		},
		{
			name:     "toml_bool",
			filename: "chezmoi.toml",
			contents: chezmoitest.JoinLines(
				`color = true`,
			),
			expectedColor: true,
		},
		{
			name:     "toml_string",
			filename: "chezmoi.toml",
			contents: chezmoitest.JoinLines(
				`color = "y"`,
			),
			expectedColor: true,
		},
		{
			name:     "yaml_bool",
			filename: "chezmoi.yaml",
			contents: chezmoitest.JoinLines(
				`color: true`,
			),
			expectedColor: true,
		},
		{
			name:     "yaml_string",
			filename: "chezmoi.yaml",
			contents: chezmoitest.JoinLines(
				`color: "yes"`,
			),
			expectedColor: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, map[string]interface{}{
				"/home/user/.config/chezmoi/" + tc.filename: tc.contents,
			}, func(fileSystem vfs.FS) {
				c := newTestConfig(t, fileSystem)
				require.NoError(t, c.execute([]string{"init"}))
				assert.Equal(t, tc.expectedColor, c.Color.Value(c.colorAutoFunc))
			})
		})
	}
}

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, expected := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, expected, upperSnakeCaseToCamelCase(s))
	}
}

func TestValidateKeys(t *testing.T) {
	for _, tc := range []struct {
		data        interface{}
		expectedErr bool
	}{
		{
			data:        nil,
			expectedErr: false,
		},
		{
			data: map[string]interface{}{
				"foo":                    "bar",
				"a":                      0,
				"_x9":                    false,
				"ThisVariableIsExported": nil,
				"αβ":                     "",
			},
			expectedErr: false,
		},
		{
			data: map[string]interface{}{
				"foo-foo": "bar",
			},
			expectedErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar-bar": "baz",
				},
			},
			expectedErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar-bar": "baz",
					},
				},
			},
			expectedErr: true,
		},
	} {
		if tc.expectedErr {
			assert.Error(t, validateKeys(tc.data, identifierRx))
		} else {
			assert.NoError(t, validateKeys(tc.data, identifierRx))
		}
	}
}

func newTestConfig(t *testing.T, fileSystem vfs.FS, options ...configOption) *Config {
	t.Helper()
	system := chezmoi.NewRealSystem(fileSystem)
	config, err := newConfig(
		append([]configOption{
			withBaseSystem(system),
			withDestSystem(system),
			withSourceSystem(system),
			withTestFS(fileSystem),
			withTestUser(t, "user"),
			withUmask(chezmoitest.Umask),
		}, options...)...,
	)
	require.NoError(t, err)
	return config
}

func withBaseSystem(baseSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.baseSystem = baseSystem
		return nil
	}
}

func withDestSystem(destSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.destSystem = destSystem
		return nil
	}
}

func withSourceSystem(sourceSystem chezmoi.System) configOption {
	return func(c *Config) error {
		c.sourceSystem = sourceSystem
		return nil
	}
}

func withStdin(stdin io.Reader) configOption {
	return func(c *Config) error {
		c.stdin = stdin
		return nil
	}
}

func withStdout(stdout io.Writer) configOption {
	return func(c *Config) error {
		c.stdout = stdout
		return nil
	}
}

func withTestFS(fileSystem vfs.FS) configOption {
	return func(c *Config) error {
		c.fileSystem = fileSystem
		return nil
	}
}

func withTestUser(t *testing.T, username string) configOption {
	t.Helper()
	return func(config *Config) error {
		var env string
		switch runtime.GOOS {
		case "plan9":
			config.homeDir = filepath.Join("/", "home", username)
			env = "home"
		case "windows":
			config.homeDir = filepath.Join("C:\\", "home", username)
			env = "USERPROFILE"
		default:
			config.homeDir = filepath.Join("/", "home", username)
			env = "HOME"
		}
		testSetenv(t, env, config.homeDir)
		var err error
		config.homeDirAbsPath, err = chezmoi.NormalizePath(config.homeDir)
		if err != nil {
			panic(err)
		}
		config.CacheDirAbsPath = config.homeDirAbsPath.Join(".cache", "chezmoi")
		config.SourceDirAbsPath = config.homeDirAbsPath.Join(".local", "share", "chezmoi")
		config.DestDirAbsPath = config.homeDirAbsPath
		config.Umask = 0o22
		configHome := filepath.Join(config.homeDir, ".config")
		dataHome := filepath.Join(config.homeDir, ".local", "share")
		config.bds = &xdg.BaseDirectorySpecification{
			ConfigHome: configHome,
			ConfigDirs: []string{configHome},
			DataHome:   dataHome,
			DataDirs:   []string{dataHome},
			CacheHome:  filepath.Join(config.homeDir, ".cache"),
			RuntimeDir: filepath.Join(config.homeDir, ".run"),
		}
		return nil
	}
}

func withUmask(umask fs.FileMode) configOption {
	return func(c *Config) error {
		c.Umask = umask
		return nil
	}
}
