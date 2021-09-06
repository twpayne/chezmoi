package chezmoi

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		Identity:  privateKeyFile,
		Recipient: publicKey,
	}

	testEncryptionDecryptToFile(t, ageEncryption)
	testEncryptionEncryptDecrypt(t, ageEncryption)
	testEncryptionEncryptFile(t, ageEncryption)
}
