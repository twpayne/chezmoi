package chezmoi

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestGPGEncryption(t *testing.T) {
	if chezmoitest.GitHubActionsOnWindows() {
		t.Skip("gpg is broken on Windows in GitHub Actions")
	}

	command, err := chezmoitest.GPGCommand()
	if errors.Is(err, exec.ErrNotFound) {
		t.Skip("gpg not found in $PATH")
	}
	require.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "chezmoi-test-gpg")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()

	key, passphrase, err := chezmoitest.GPGGenerateKey(command, tempDir)
	require.NoError(t, err)

	for _, tc := range []struct {
		name      string
		symmetric bool
	}{
		{
			name:      "asymmetric",
			symmetric: false,
		},
		{
			name:      "symmetric",
			symmetric: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gpgEncryption := &GPGEncryption{
				Command: command,
				Args: []string{
					"--homedir", tempDir,
					"--no-tty",
					"--passphrase", passphrase,
					"--pinentry-mode", "loopback",
				},
				Recipient: key,
				Symmetric: tc.symmetric,
			}

			testEncryptionDecryptToFile(t, gpgEncryption)
			testEncryptionEncryptDecrypt(t, gpgEncryption)
			testEncryptionEncryptFile(t, gpgEncryption)
		})
	}
}
