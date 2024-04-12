package cmd

import (
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	vfs "github.com/twpayne/go-vfs/v5"
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

func TestConfigFileFormatRoundTrip(t *testing.T) {
	for _, format := range []chezmoi.Format{
		chezmoi.FormatJSON,
		chezmoi.FormatYAML,
	} {
		t.Run(format.Name(), func(t *testing.T) {
			configFile := ConfigFile{
				Color:        autoBool{auto: true},
				Data:         map[string]any{},
				Env:          map[string]string{},
				Hooks:        map[string]hookConfig{},
				Interpreters: map[string]chezmoi.Interpreter{},
				Mode:         chezmoi.ModeFile,
				PINEntry: pinEntryConfig{
					Args:    []string{},
					Options: []string{},
				},
				ScriptEnv: map[string]string{},
				Template: templateConfig{
					Options: []string{},
				},
				TextConv:      []*textConvElement{},
				UseBuiltinAge: autoBool{value: false},
				UseBuiltinGit: autoBool{value: true},
				Dashlane: dashlaneConfig{
					Args: []string{},
				},
				Doppler: dopplerConfig{
					Args: []string{},
				},
				HCPVaultSecrets: hcpVaultSecretConfig{
					Args: []string{},
				},
				Keepassxc: keepassxcConfig{
					Args: []string{},
				},
				Keeper: keeperConfig{
					Args: []string{},
				},
				Passhole: passholeConfig{
					Args: []string{},
				},
				Secret: secretConfig{
					Args: []string{},
				},
				Age: chezmoi.AgeEncryption{
					Args:            []string{},
					Identity:        chezmoi.NewAbsPath("/identity.txt"),
					Identities:      []chezmoi.AbsPath{},
					Recipients:      []string{},
					RecipientsFiles: []chezmoi.AbsPath{},
				},
				GPG: chezmoi.GPGEncryption{
					Args:       []string{},
					Recipients: []string{},
				},
				Add: addCmdConfig{
					Secrets: severityError,
				},
				CD: cdCmdConfig{
					Args: []string{},
				},
				Diff: diffCmdConfig{
					Args: []string{},
				},
				Edit: editCmdConfig{
					Args: []string{},
				},
				Merge: mergeCmdConfig{
					Args: []string{},
				},
				Update: updateCmdConfig{
					Args: []string{},
				},
			}
			data, err := format.Marshal(configFile)
			assert.NoError(t, err)
			var actualConfigFile ConfigFile
			assert.NoError(t, format.Unmarshal(data, &actualConfigFile))
			assert.Equal(t, configFile, actualConfigFile)
		})
	}
}

func TestParseCommand(t *testing.T) {
	for i, tc := range []struct {
		command         string
		args            []string
		expectedCommand string
		expectedArgs    []string
		expectedErr     bool
	}{
		{
			command:         "chezmoi-editor",
			expectedCommand: "chezmoi-editor",
		},
		{
			command:         `chezmoi-editor -f --nomru -c "au VimLeave * !open -a Terminal"`,
			expectedCommand: "chezmoi-editor",
			expectedArgs:    []string{"-f", "--nomru", "-c", "au VimLeave * !open -a Terminal"},
		},
		{
			command:         `"chezmoi editor" $CHEZMOI_TEST_VAR`,
			args:            []string{"extra-arg"},
			expectedCommand: "chezmoi editor",
			expectedArgs:    []string{"chezmoi-test-value", "extra-arg"},
		},
		{
			command:     `"chezmoi editor`,
			expectedErr: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Setenv("CHEZMOI_TEST_VAR", "chezmoi-test-value")
			actualCommand, actualArgs, err := parseCommand(tc.command, tc.args)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommand, actualCommand)
				assert.Equal(t, tc.expectedArgs, actualArgs)
			}
		})
	}
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
			chezmoitest.WithTestFS(t, map[string]any{
				"/home/user/.config/chezmoi/" + tc.filename: tc.contents,
			}, func(fileSystem vfs.FS) {
				c := newTestConfig(t, fileSystem)
				assert.NoError(t, c.execute([]string{"init"}))
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
			withVersionInfo(VersionInfo{
				Version: "2.0.0",
			}),
		}, options...)...,
	)
	assert.NoError(t, err)
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

func withNoTTY(noTTY bool) configOption { //nolint:unparam
	return func(c *Config) error {
		c.noTTY = noTTY
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
		t.Setenv(env, config.homeDir)
		var err error
		config.homeDirAbsPath, err = chezmoi.NormalizePath(config.homeDir)
		if err != nil {
			t.Fatal(err)
		}
		config.CacheDirAbsPath = config.homeDirAbsPath.JoinString(".cache", "chezmoi")
		config.SourceDirAbsPath = config.homeDirAbsPath.JoinString(".local", "share", "chezmoi")
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
