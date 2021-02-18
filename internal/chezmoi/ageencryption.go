package chezmoi

import (
	"bytes"
	"os/exec"

	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
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
func (e *AGEEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.decryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	plaintext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *AGEEncryption) DecryptToFile(filename string, ciphertext []byte) error {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.decryptArgs(), "--output", filename), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// Encrypt implements Encryption.Encrypt.
func (e *AGEEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.encryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(plaintext)
	ciphertext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (e *AGEEncryption) EncryptFile(filename string) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.encryptArgs(), e.Args...), filename)...)
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *AGEEncryption) EncryptedSuffix() string {
	return e.Suffix
}

func (e *AGEEncryption) decryptArgs() []string {
	args := make([]string, 0, 1+2*(1+len(e.Identities)))
	args = append(args, "--decrypt")
	if e.Identity != "" {
		args = append(args, "--identity", e.Identity)
	}
	for _, identity := range e.Identities {
		args = append(args, "--identity", identity)
	}
	return args
}

func (e *AGEEncryption) encryptArgs() []string {
	args := make([]string, 0, 1+2*(1+len(e.Recipients))+2*(1+len(e.RecipientsFiles)))
	args = append(args, "--armor")
	if e.Recipient != "" {
		args = append(args, "--recipient", e.Recipient)
	}
	for _, recipient := range e.Recipients {
		args = append(args, "--recipient", recipient)
	}
	if e.RecipientsFile != "" {
		args = append(args, "--recipients-file", e.RecipientsFile)
	}
	for _, recipientsFile := range e.RecipientsFiles {
		args = append(args, "--recipients-file", recipientsFile)
	}
	return args
}
