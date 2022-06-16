package chezmoi

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestGPGEncryption(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping gpg tests on Windows")
	}
	command := lookPathOrSkip(t, "gpg")

	tempDir := t.TempDir()
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
			testEncryption(t, &GPGEncryption{
				Command: command,
				Args: []string{
					"--homedir", tempDir,
					"--no-tty",
					"--passphrase", passphrase,
					"--pinentry-mode", "loopback",
				},
				Recipient: key,
				Symmetric: tc.symmetric,
			})
		})
	}
}

func TestGPGEncryptionMultipleRecipients(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping gpg tests on Windows")
	}
	command := lookPathOrSkip(t, "gpg")

	tempDir := t.TempDir()

	// Generate two GPG private keys for testing in the same GNUPGHOME.
	// NOTE: GPGGenerateKey returns a fixed password, so using the same variable for both keys is acceptable.
	key1, passphrase, err1 := chezmoitest.GPGGenerateKey(command, tempDir)
	require.NoError(t, err1)

	key2, _, err2 := chezmoitest.GPGGenerateKey(command, tempDir)
	require.NoError(t, err2)

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
			testEncryption(t, &GPGEncryption{
				Command: command,
				Args: []string{
					"--homedir", tempDir,
					"--no-tty",
					"--passphrase", passphrase,
					"--pinentry-mode", "loopback",
				},
				Recipients: []string{
					key1,
					key2,
				},
				Symmetric: tc.symmetric,
			})
		})
	}
}
