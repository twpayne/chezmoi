package chezmoi

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

// A GPGEncryption uses gpg for encryption and decryption. See https://gnupg.org/.
type GPGEncryption struct {
	Command   string
	Args      []string
	Recipient string
	Symmetric bool
	Suffix    string
}

// Decrypt implements Encyrption.Decrypt.
func (e *GPGEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	args := append([]string{"--decrypt"}, e.Args...)
	//nolint:gosec
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *GPGEncryption) DecryptToFile(filename string, ciphertext []byte) error {
	args := append([]string{
		"--decrypt",
		"--output", filename,
		"--yes",
	}, e.Args...)
	//nolint:gosec
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// Encrypt implements Encryption.Encrypt.
func (e *GPGEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	args := append(e.encryptArgs(), e.Args...)
	//nolint:gosec
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// EncryptFile implements Encryption.EncryptFile.
func (e *GPGEncryption) EncryptFile(filename string) (ciphertext []byte, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	args := append(e.encryptArgs(), e.Args...)
	//nolint:gosec
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = f
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *GPGEncryption) EncryptedSuffix() string {
	return e.Suffix
}

func (e *GPGEncryption) encryptArgs() []string {
	args := []string{
		"--armor",
		"--encrypt",
	}
	if e.Recipient != "" {
		args = append(args, "--recipient", e.Recipient)
	}
	if e.Symmetric {
		args = append(args, "--symmetric")
	}
	return args
}
