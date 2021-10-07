package chezmoi

import (
	"os"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestAgeEncryption(t *testing.T) {
	command := lookPathOrSkip(t, "age")

	identityFile := filepath.Join(t.TempDir(), "chezmoi-test-age-key.txt")
	recipient, err := chezmoitest.AgeGenerateKey(identityFile)
	require.NoError(t, err)

	testEncryption(t, &AgeEncryption{
		Command:   command,
		Identity:  NewAbsPath(identityFile),
		Recipient: recipient,
	})
}

func TestAgeMultipleIdentitiesAndMultipleRecipients(t *testing.T) {
	command := lookPathOrSkip(t, "age")

	tempDir := t.TempDir()
	identityFile1 := filepath.Join(tempDir, "chezmoi-test-age-key1.txt")
	recipient1, err := chezmoitest.AgeGenerateKey(identityFile1)
	require.NoError(t, err)
	identityFile2 := filepath.Join(tempDir, "chezmoi-test-age-key2.txt")
	recipient2, err := chezmoitest.AgeGenerateKey(identityFile2)
	require.NoError(t, err)

	testEncryption(t, &AgeEncryption{
		Command: command,
		Identities: []AbsPath{
			NewAbsPath(identityFile1),
			NewAbsPath(identityFile2),
		},
		Recipients: []string{
			recipient1,
			recipient2,
		},
	})
}

func TestAgeRecipientsFile(t *testing.T) {
	command := lookPathOrSkip(t, "age")

	tempDir := t.TempDir()
	identityFile := filepath.Join(tempDir, "chezmoi-test-age-key.txt")
	recipient, err := chezmoitest.AgeGenerateKey(identityFile)
	require.NoError(t, err)
	recipientsFile := filepath.Join(t.TempDir(), "chezmoi-test-age-recipients.txt")
	require.NoError(t, os.WriteFile(recipientsFile, []byte(recipient), 0o666))

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
}

func TestBuiltinAgeEncryption(t *testing.T) {
	recipientStringer, identityAbsPath := builtinAgeGenerateKey(t)

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		BaseSystem: NewRealSystem(vfs.OSFS),
		Identity:   identityAbsPath,
		Recipient:  recipientStringer.String(),
	})
}

func TestBuiltinAgeMultipleIdentitiesAndMultipleRecipients(t *testing.T) {
	recipient1, identityAbsPath1 := builtinAgeGenerateKey(t)
	recipient2, identityAbsPath2 := builtinAgeGenerateKey(t)

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		BaseSystem: NewRealSystem(vfs.OSFS),
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
	baseSystem := NewRealSystem(vfs.OSFS)
	recipient, identityAbsPath := builtinAgeGenerateKey(t)
	recipientsFile := filepath.Join(t.TempDir(), "chezmoi-builtin-age-recipients.txt")
	require.NoError(t, os.WriteFile(recipientsFile, []byte(recipient.String()), 0o666))

	testEncryption(t, &AgeEncryption{
		UseBuiltin:     true,
		BaseSystem:     baseSystem,
		Identity:       identityAbsPath,
		RecipientsFile: NewAbsPath(recipientsFile),
	})

	testEncryption(t, &AgeEncryption{
		UseBuiltin: true,
		BaseSystem: baseSystem,
		Identity:   identityAbsPath,
		RecipientsFiles: []AbsPath{
			NewAbsPath(recipientsFile),
		},
	})
}

func builtinAgeGenerateKey(t *testing.T) (*age.X25519Recipient, AbsPath) {
	t.Helper()
	identity, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	identityFile := filepath.Join(t.TempDir(), "chezmoi-test-builtin-age-key.txt")
	require.NoError(t, os.WriteFile(identityFile, []byte(identity.String()), 0o600))
	return identity.Recipient(), NewAbsPath(identityFile)
}
