package chezmoi

import (
	"runtime"
	"testing"

	"github.com/alecthomas/assert/v2"

	"chezmoi.io/chezmoi/internal/chezmoitest"
)

func TestGPGEncryption(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping gpg tests on Windows")
	}
	command := lookPathOrSkip(t, "gpg")

	tempDir := t.TempDir()
	key, passphrase, err := chezmoitest.GPGGenerateKey(command, tempDir)
	assert.NoError(t, err)

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
