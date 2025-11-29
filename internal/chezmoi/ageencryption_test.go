package chezmoi

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

var ageCommands = []string{
	"age",
	"rage",
}

func TestAgeEncryption(t *testing.T) {
	forEachAgeCommand(t, func(t *testing.T, command string) {
		t.Helper()

		identityFile := filepath.Join(t.TempDir(), "chezmoi-test-age-key.txt")
		recipient, err := chezmoitest.AgeGenerateKey(command, identityFile)
		assert.NoError(t, err)

		testEncryption(t, &AgeEncryption{
			Command:   command,
			Identity:  NewAbsPath(identityFile),
			Recipient: recipient.String(),
		})
	})
}

func TestAgeEncryptionMarshalUnmarshal(t *testing.T) {
	for _, format := range []Format{
		FormatJSON,
		FormatYAML,
	} {
		t.Run(format.Name(), func(t *testing.T) {
			expected := AgeEncryption{
				UseBuiltin: true,
				Command:    "command",
				Args: []string{
					"arg1",
					"arg2",
				},
				Identity: NewAbsPath("/identity"),
				Identities: []AbsPath{
					NewAbsPath("/identity1"),
					NewAbsPath("/identity2"),
				},
				Passphrase:     true,
				Recipient:      "recipient",
				RecipientsFile: NewAbsPath("/recipients-file"),
				RecipientsFiles: []AbsPath{
					NewAbsPath("/recipients-file1"),
					NewAbsPath("/recipients-file2"),
				},
				Suffix:    "suffix",
				Symmetric: true,
			}
			data, err := format.Marshal(expected)
			assert.NoError(t, err)
			var actual AgeEncryption
			assert.NoError(t, format.Unmarshal(data, &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestAgeEncryptionMarshalUnmarshalField(t *testing.T) {
	type ConfigFile struct {
		Age AgeEncryption `json:"age" yaml:"age"`
	}
	for _, format := range []Format{
		FormatJSON,
		FormatYAML,
	} {
		t.Run(format.Name(), func(t *testing.T) {
			expected := ConfigFile{
				Age: AgeEncryption{
					UseBuiltin: true,
					Command:    "command",
					Args: []string{
						"arg1",
						"arg2",
					},
					Identity: NewAbsPath("/identity"),
					Identities: []AbsPath{
						NewAbsPath("/identity1"),
						NewAbsPath("/identity2"),
					},
					Passphrase:     true,
					Recipient:      "recipient",
					RecipientsFile: NewAbsPath("/recipients-file"),
					RecipientsFiles: []AbsPath{
						NewAbsPath("/recipients-file1"),
						NewAbsPath("/recipients-file2"),
					},
					Suffix:    "suffix",
					Symmetric: true,
				},
			}
			data, err := format.Marshal(expected)
			assert.NoError(t, err)
			var actual ConfigFile
			assert.NoError(t, format.Unmarshal(data, &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestAgeEncryptionMarshalUnmarshalFieldEmbedded(t *testing.T) {
	type ConfigFile struct {
		Age AgeEncryption `json:"age" yaml:"age"`
	}
	type Config struct {
		ConfigFile
	}
	for _, format := range []Format{
		FormatJSON,
		FormatYAML,
	} {
		t.Run(format.Name(), func(t *testing.T) {
			expected := Config{
				ConfigFile: ConfigFile{
					Age: AgeEncryption{
						UseBuiltin: true,
						Command:    "command",
						Args: []string{
							"arg1",
							"arg2",
						},
						Identity: NewAbsPath("/identity"),
						Identities: []AbsPath{
							NewAbsPath("/identity1"),
							NewAbsPath("/identity2"),
						},
						Passphrase:     true,
						Recipient:      "recipient",
						RecipientsFile: NewAbsPath("/recipients-file"),
						RecipientsFiles: []AbsPath{
							NewAbsPath("/recipients-file1"),
							NewAbsPath("/recipients-file2"),
						},
						Suffix:    "suffix",
						Symmetric: true,
					},
				},
			}
			data, err := format.Marshal(expected)
			assert.NoError(t, err)
			var actual Config
			assert.NoError(t, format.Unmarshal(data, &actual))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestAgeMultipleIdentitiesAndMultipleRecipients(t *testing.T) {
	forEachAgeCommand(t, func(t *testing.T, command string) {
		t.Helper()

		tempDir := t.TempDir()

		identityFile1 := filepath.Join(tempDir, "chezmoi-test-age-key1.txt")
		recipient1, err := chezmoitest.AgeGenerateKey(command, identityFile1)
		assert.NoError(t, err)

		identityFile2 := filepath.Join(tempDir, "chezmoi-test-age-key2.txt")
		recipient2, err := chezmoitest.AgeGenerateKey(command, identityFile2)
		assert.NoError(t, err)

		testEncryption(t, &AgeEncryption{
			Command: command,
			Identities: []AbsPath{
				NewAbsPath(identityFile1),
				NewAbsPath(identityFile2),
			},
			Recipients: []string{
				recipient1.String(),
				recipient2.String(),
			},
		})
	})
}

func TestAgeRecipientsFile(t *testing.T) {
	t.Helper()

	forEachAgeCommand(t, func(t *testing.T, command string) {
		t.Helper()

		tempDir := t.TempDir()

		identityFile := filepath.Join(tempDir, "chezmoi-test-age-key.txt")
		recipient, err := chezmoitest.AgeGenerateKey(command, identityFile)
		assert.NoError(t, err)

		recipientsFile := filepath.Join(t.TempDir(), "chezmoi-test-age-recipients.txt")
		assert.NoError(t, os.WriteFile(recipientsFile, []byte(recipient.String()), 0o666))

		testEncryption(t, &AgeEncryption{
			Command:        command,
			Identity:       NewAbsPath(identityFile),
			RecipientsFile: NewAbsPath(recipientsFile),
		})

		testEncryption(t, &AgeEncryption{
			Command:  command,
			Identity: NewAbsPath(identityFile),
			RecipientsFiles: []AbsPath{
				NewAbsPath(recipientsFile),
			},
		})
	})
}

func TestBuiltinAgeEncryption(t *testing.T) {
	recipientStringer, identityAbsPath := builtinAgeGenerateKey(t)

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		Identity:   identityAbsPath,
		Recipient:  recipientStringer.String(),
	})
}

func TestBuiltinAgeMultipleIdentitiesAndMultipleRecipients(t *testing.T) {
	recipient1, identityAbsPath1 := builtinAgeGenerateKey(t)
	recipient2, identityAbsPath2 := builtinAgeGenerateKey(t)

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		Identities: []AbsPath{
			identityAbsPath1,
			identityAbsPath2,
		},
		Recipients: []string{
			recipient1.String(),
			recipient2.String(),
		},
	})
}

func TestBuiltinAgeRecipientsFile(t *testing.T) {
	recipient, identityAbsPath := builtinAgeGenerateKey(t)
	recipientsFile := filepath.Join(t.TempDir(), "chezmoi-builtin-age-recipients.txt")
	assert.NoError(t, os.WriteFile(recipientsFile, []byte(recipient.String()), 0o666))

	testEncryption(t, &AgeEncryption{
		UseBuiltin:     true,
		Identity:       identityAbsPath,
		RecipientsFile: NewAbsPath(recipientsFile),
	})

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		Identity:   identityAbsPath,
		RecipientsFiles: []AbsPath{
			NewAbsPath(recipientsFile),
		},
	})
}

func builtinAgeGenerateKey(t *testing.T) (*age.X25519Recipient, AbsPath) {
	t.Helper()
	identity, err := age.GenerateX25519Identity()
	assert.NoError(t, err)
	identityFile := filepath.Join(t.TempDir(), "chezmoi-test-builtin-age-key.txt")
	assert.NoError(t, os.WriteFile(identityFile, []byte(identity.String()), 0o600))
	return identity.Recipient(), NewAbsPath(identityFile)
}

func forEachAgeCommand(t *testing.T, f func(*testing.T, string)) {
	t.Helper()
	for _, command := range ageCommands {
		t.Run(command, func(t *testing.T) {
			f(t, lookPathOrSkip(t, command))
		})
	}
}
