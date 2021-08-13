package chezmoi

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// An AGEEncryption uses age for encryption and decryption. See
// https://age-encryption.org.
type AGEEncryption struct {
	Command         string
	Args            []string
	Identity        string
	Identities      []string
	Passphrase      bool
	Recipient       string
	Recipients      []string
	RecipientsFile  AbsPath
	RecipientsFiles []AbsPath
	Suffix          string
	Symmetric       bool
}

// Decrypt implements Encyrption.Decrypt.
func (e *AGEEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.decryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	plaintext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *AGEEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.decryptArgs(), "--output", string(plaintextAbsPath)), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// Encrypt implements Encryption.Encrypt.
func (e *AGEEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.encryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stderr = os.Stderr
	ciphertext, err := chezmoilog.LogCmdOutput(log.Logger, cmd)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (e *AGEEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.encryptArgs(), e.Args...), string(plaintextAbsPath))...)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *AGEEncryption) EncryptedSuffix() string {
	return e.Suffix
}

// decryptArgs returns the arguments for decryption.
func (e *AGEEncryption) decryptArgs() []string {
	var args []string
	args = append(args, "--decrypt")
	if !e.Passphrase {
		args = append(args, e.identityArgs()...)
	}
	return args
}

// encryptArgs returns the arguments for encryption.
func (e *AGEEncryption) encryptArgs() []string {
	var args []string
	args = append(args,
		"--armor",
		"--encrypt",
	)
	switch {
	case e.Passphrase:
		args = append(args, "--passphrase")
	case e.Symmetric:
		args = append(args, e.identityArgs()...)
	default:
		if e.Recipient != "" {
			args = append(args, "--recipient", e.Recipient)
		}
		for _, recipient := range e.Recipients {
			args = append(args, "--recipient", recipient)
		}
		if e.RecipientsFile != "" {
			args = append(args, "--recipients-file", string(e.RecipientsFile))
		}
		for _, recipientsFile := range e.RecipientsFiles {
			args = append(args, "--recipients-file", string(recipientsFile))
		}
	}
	return args
}

func (e *AGEEncryption) identityArgs() []string {
	args := make([]string, 0, 2+2*len(e.Identities))
	if e.Identity != "" {
		args = append(args, "--identity", e.Identity)
	}
	for _, identity := range e.Identities {
		args = append(args, "--identity", identity)
	}
	return args
}
