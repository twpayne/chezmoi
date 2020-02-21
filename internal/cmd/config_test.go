package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
)

func TestAutoCommitCommitMessage(t *testing.T) {
	commitMessageText, err := getAsset("assets/templates/COMMIT_MESSAGE.tmpl")
	require.NoError(t, err)
	commitMessageTmpl, err := template.New("commit_message").Funcs(sprig.HermeticTxtFuncMap()).Parse(string(commitMessageText))
	require.NoError(t, err)
	for _, tc := range []struct {
		name            string
		statusStr       string
		wantErr         bool
		expectedMessage string
	}{
		{
			name:            "add",
			statusStr:       "1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 cea5c3500651a923bacd80f960dd20f04f71d509 main.go\n",
			expectedMessage: "Add main.go\n",
		},
		{
			name:            "remove",
			statusStr:       "1 D. N... 100644 000000 000000 cea5c3500651a923bacd80f960dd20f04f71d509 0000000000000000000000000000000000000000 main.go\n",
			expectedMessage: "Remove main.go\n",
		},
		{
			name:            "update",
			statusStr:       "1 M. N... 100644 100644 100644 353dbbb3c29a80fb44d4e26dac111739d25294db 353dbbb3c29a80fb44d4e26dac111739d25294db main.go\n",
			expectedMessage: "Update main.go\n",
		},
		{
			name:            "rename",
			statusStr:       "2 R. N... 100644 100644 100644 9d06c86ecba40e1c695e69b55a40843df6a79cef 9d06c86ecba40e1c695e69b55a40843df6a79cef R100 chezmoi_rename.go chezmoi.go\n",
			expectedMessage: "Rename chezmoi.go to chezmoi_rename.go\n",
		},
		{
			name:      "unsupported_xy",
			statusStr: "1 MM N... 100644 100644 100644 353dbbb3c29a80fb44d4e26dac111739d25294db 353dbbb3c29a80fb44d4e26dac111739d25294db main.go\n",
			wantErr:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, err := gitVCS{}.ParseStatusOutput([]byte(tc.statusStr))
			require.NoError(t, err)
			b := &bytes.Buffer{}
			err = commitMessageTmpl.Execute(b, status)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedMessage, b.String())
			}
		})
	}
}

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, want := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, want, upperSnakeCaseToCamelCase(s))
	}
}

func TestValidateKeys(t *testing.T) {
	for _, tc := range []struct {
		data    interface{}
		wantErr bool
	}{
		{
			data:    nil,
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo":                    "bar",
				"a":                      0,
				"_x9":                    false,
				"ThisVariableIsExported": nil,
				"αβ":                     "",
			},
			wantErr: false,
		},
		{
			data: map[string]interface{}{
				"foo-foo": "bar",
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": map[string]interface{}{
					"bar-bar": "baz",
				},
			},
			wantErr: true,
		},
		{
			data: map[string]interface{}{
				"foo": []interface{}{
					map[string]interface{}{
						"bar-bar": "baz",
					},
				},
			},
			wantErr: true,
		},
	} {
		if tc.wantErr {
			assert.Error(t, validateKeys(tc.data, identifierRegexp))
		} else {
			assert.NoError(t, validateKeys(tc.data, identifierRegexp))
		}
	}
}

func newTestConfig(fs vfs.FS, options ...configOption) *Config {
	return newConfig(append(
		[]configOption{
			withTestFS(fs),
			withTestUser("user"),
		},
		options...,
	)...)
}

func withAddCmdConfig(add addCmdConfig) configOption {
	return func(c *Config) {
		c.add = add
	}
}

func withData(data map[string]interface{}) configOption {
	return func(c *Config) {
		c.Data = data
	}
}

func withDestDir(destDir string) configOption {
	return func(c *Config) {
		c.DestDir = destDir
	}
}

func withDumpCmdConfig(dumpCmdConfig dumpCmdConfig) configOption {
	return func(c *Config) {
		c.dump = dumpCmdConfig
	}
}

func withFollow(follow bool) configOption {
	return func(c *Config) {
		c.Follow = follow
	}
}

func withGenericSecretCmdConfig(genericSecretCmdConfig genericSecretCmdConfig) configOption {
	return func(c *Config) {
		c.GenericSecret = genericSecretCmdConfig
	}
}

func withMutator(mutator chezmoi.Mutator) configOption {
	return func(c *Config) {
		c.mutator = mutator
	}
}

func withRemove(remove bool) configOption {
	return func(c *Config) {
		c.Remove = remove
	}
}

func withRemoveCmdConfig(remove removeCmdConfig) configOption {
	return func(c *Config) {
		c.remove = remove
	}
}

func withStdin(stdin io.Reader) configOption {
	return func(c *Config) {
		c.Stdin = stdin
	}
}

func withStdout(stdout io.Writer) configOption {
	return func(c *Config) {
		c.Stdout = stdout
	}
}

func withTestFS(fs vfs.FS) configOption {
	return func(c *Config) {
		c.fs = fs
		c.mutator = chezmoi.NewVerboseMutator(os.Stdout, chezmoi.NewFSMutator(fs), false, 0)
		c.Verbose = true
	}
}

func withTestUser(username string) configOption {
	return func(c *Config) {
		homeDir := filepath.Join("/", "home", username)
		c.SourceDir = filepath.Join(homeDir, ".local", "share", "chezmoi")
		c.DestDir = homeDir
		c.Umask = 022
		c.bds = &xdg.BaseDirectorySpecification{
			ConfigHome: filepath.Join(homeDir, ".config"),
			DataHome:   filepath.Join(homeDir, ".local"),
			CacheHome:  filepath.Join(homeDir, ".cache"),
			RuntimeDir: filepath.Join(homeDir, ".run"),
		}
	}
}
