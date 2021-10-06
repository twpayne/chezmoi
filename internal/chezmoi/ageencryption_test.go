package chezmoi

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"filippo.io/age"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestAgeEncryption(t *testing.T) {
	command, err := exec.LookPath("age")
	if errors.Is(err, exec.ErrNotFound) {
		t.Skip("age not found in $PATH")
	}
	require.NoError(t, err)

	publicKey, privateKeyFile, err := chezmoitest.AgeGenerateKey("")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(filepath.Dir(privateKeyFile)))
	}()

	ageEncryption := &AgeEncryption{
		Command:   command,
		Identity:  NewAbsPath(privateKeyFile),
		Recipient: publicKey,
	}

	testEncryptionDecryptToFile(t, ageEncryption)
	testEncryptionEncryptDecrypt(t, ageEncryption)
	testEncryptionEncryptFile(t, ageEncryption)
}

func TestBuiltinAgeEncryption(t *testing.T) {
	recipientStringer, identityAbsPath := builtinAgeGenerateKey(t)

	ageEncryption := &AgeEncryption{
		UseBuiltin: true,
		BaseSystem: NewRealSystem(vfs.OSFS),
		Identity:   identityAbsPath,
		Recipient:  recipientStringer.String(),
	}

	testEncryptionDecryptToFile(t, ageEncryption)
	testEncryptionEncryptDecrypt(t, ageEncryption)
	testEncryptionEncryptFile(t, ageEncryption)
}

func builtinAgeGenerateKey(t *testing.T) (fmt.Stringer, AbsPath) {
	identity, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	privateKeyFile := filepath.Join(t.TempDir(), "chezmoi-builtin-age-key.txt")
	require.NoError(t, os.WriteFile(privateKeyFile, []byte(identity.String()), 0o600))
	return identity.Recipient(), NewAbsPath(privateKeyFile)
}
