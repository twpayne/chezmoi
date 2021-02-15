package chezmoi

import (
	"bytes"
	"os/exec"

	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoilog"
)

// An AGEEncryption uses age for encryption and decryption. See
// https://github.com/FiloSottile/age.
type AGEEncryption struct {
	Command         string
	Args            []string
	Identity        string
	Identities      []string
	Recipient       string
	Recipients      []string
	RecipientsFile  string
	RecipientsFiles []string
	Suffix          string
}

// Decrypt implements Encyrption.Decrypt.
func (t *AGEEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(t.Command, append(t.decryptArgs(), t.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	plaintext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (t *AGEEncryption) DecryptToFile(filename string, ciphertext []byte) error {
	//nolint:gosec
	cmd := exec.Command(t.Command, append(append(t.decryptArgs(), "--output", filename), t.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// Encrypt implements Encryption.Encrypt.
func (t *AGEEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(t.Command, append(t.encryptArgs(), t.Args...)...)
	cmd.Stdin = bytes.NewReader(plaintext)
	ciphertext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (t *AGEEncryption) EncryptFile(filename string) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(t.Command, append(append(t.encryptArgs(), t.Args...), filename)...)
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (t *AGEEncryption) EncryptedSuffix() string {
	return t.Suffix
}

func (t *AGEEncryption) decryptArgs() []string {
	args := make([]string, 0, 1+2*(1+len(t.Identities)))
	args = append(args, "--decrypt")
	if t.Identity != "" {
		args = append(args, "--identity", t.Identity)
	}
	for _, identity := range t.Identities {
		args = append(args, "--identity", identity)
	}
	return args
}

func (t *AGEEncryption) encryptArgs() []string {
	args := make([]string, 0, 1+2*(1+len(t.Recipients))+2*(1+len(t.RecipientsFiles)))
	args = append(args, "--armor")
	if t.Recipient != "" {
		args = append(args, "--recipient", t.Recipient)
	}
	for _, recipient := range t.Recipients {
		args = append(args, "--recipient", recipient)
	}
	if t.RecipientsFile != "" {
		args = append(args, "--recipients-file", t.RecipientsFile)
	}
	for _, recipientsFile := range t.RecipientsFiles {
		args = append(args, "--recipients-file", recipientsFile)
	}
	return args
}
