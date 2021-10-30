package chezmoi

// FIXME add builtin support for --passphrase
// FIXME add builtin support for --symmetric
// FIXME add builtin support for SSH keys if recommended

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"filippo.io/age"
	"filippo.io/age/armor"
	"go.uber.org/multierr"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// An AgeEncryption uses age for encryption and decryption. See
// https://age-encryption.org.
type AgeEncryption struct {
	UseBuiltin      bool
	BaseSystem      System
	Command         string
	Args            []string
	Identity        AbsPath
	Identities      []AbsPath
	Passphrase      bool
	Recipient       string
	Recipients      []string
	RecipientsFile  AbsPath
	RecipientsFiles []AbsPath
	Suffix          string
	Symmetric       bool
}

// Decrypt implements Encyrption.Decrypt.
func (e *AgeEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if e.UseBuiltin {
		return e.builtinDecrypt(ciphertext)
	}

	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.decryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(cmd)
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *AgeEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	if e.UseBuiltin {
		plaintext, err := e.builtinDecrypt(ciphertext)
		if err != nil {
			return err
		}
		return e.BaseSystem.WriteFile(plaintextAbsPath, plaintext, 0o644) // FIXME encrypted executables
	}

	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.decryptArgs(), "--output", plaintextAbsPath.String()), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(cmd)
}

// Encrypt implements Encryption.Encrypt.
func (e *AgeEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	if e.UseBuiltin {
		return e.builtinEncrypt(plaintext)
	}

	//nolint:gosec
	cmd := exec.Command(e.Command, append(e.encryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(cmd)
}

// EncryptFile implements Encryption.EncryptFile.
func (e *AgeEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	if e.UseBuiltin {
		plaintext, err := e.BaseSystem.ReadFile(plaintextAbsPath)
		if err != nil {
			return nil, err
		}
		return e.builtinEncrypt(plaintext)
	}

	//nolint:gosec
	cmd := exec.Command(e.Command, append(append(e.encryptArgs(), e.Args...), plaintextAbsPath.String())...)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(cmd)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *AgeEncryption) EncryptedSuffix() string {
	return e.Suffix
}

// builtinDecrypt decrypts ciphertext using the builtin age.
func (e *AgeEncryption) builtinDecrypt(ciphertext []byte) ([]byte, error) {
	identities, err := e.builtinIdentities()
	if err != nil {
		return nil, err
	}
	r, err := age.Decrypt(armor.NewReader(bytes.NewReader(ciphertext)), identities...)
	if err != nil {
		return nil, err
	}
	buffer := &bytes.Buffer{}
	if _, err = io.Copy(buffer, r); err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}

// builtinEncrypt encrypts ciphertext using the builtin age.
func (e *AgeEncryption) builtinEncrypt(plaintext []byte) ([]byte, error) {
	recipients, err := e.builtinRecipients()
	if err != nil {
		return nil, err
	}
	output := &bytes.Buffer{}
	armorWriter := armor.NewWriter(output)
	writer, err := age.Encrypt(armorWriter, recipients...)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(writer, bytes.NewReader(plaintext)); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if err := armorWriter.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// builtinIdentities returns the identities for decryption using the builtin
// age.
func (e *AgeEncryption) builtinIdentities() ([]age.Identity, error) {
	var identities []age.Identity
	if !e.Identity.Empty() {
		parsedIdentities, err := parseIdentityFile(e.Identity)
		if err != nil {
			return nil, err
		}
		identities = append(identities, parsedIdentities...)
	}
	for _, identityAbsPath := range e.Identities {
		parsedIdentities, err := parseIdentityFile(identityAbsPath)
		if err != nil {
			return nil, err
		}
		identities = append(identities, parsedIdentities...)
	}
	return identities, nil
}

// builtinRecipients returns the recipients for encryption using the builtin
// age.
func (e *AgeEncryption) builtinRecipients() ([]age.Recipient, error) {
	recipients := make([]age.Recipient, 0, 1+len(e.Recipients))
	if e.Recipient != "" {
		parsedRecipient, err := age.ParseX25519Recipient(e.Recipient)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, parsedRecipient)
	}
	for _, recipient := range e.Recipients {
		parsedRecipient, err := age.ParseX25519Recipient(recipient)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, parsedRecipient)
	}
	if !e.RecipientsFile.Empty() {
		parsedRecipients, err := parseRecipientsFile(e.RecipientsFile)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, parsedRecipients...)
	}
	for _, recipientsFile := range e.RecipientsFiles {
		parsedRecipients, err := parseRecipientsFile(recipientsFile)
		if err != nil {
			return nil, err
		}
		recipients = append(recipients, parsedRecipients...)
	}
	return recipients, nil
}

// decryptArgs returns the arguments for decryption.
func (e *AgeEncryption) decryptArgs() []string {
	var args []string
	args = append(args, "--decrypt")
	if !e.Passphrase {
		args = append(args, e.identityArgs()...)
	}
	return args
}

// encryptArgs returns the arguments for encryption.
func (e *AgeEncryption) encryptArgs() []string {
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
		if !e.RecipientsFile.Empty() {
			args = append(args, "--recipients-file", e.RecipientsFile.String())
		}
		for _, recipientsFile := range e.RecipientsFiles {
			args = append(args, "--recipients-file", recipientsFile.String())
		}
	}
	return args
}

// identityArgs returns the arguments for identity.
func (e *AgeEncryption) identityArgs() []string {
	args := make([]string, 0, 2+2*len(e.Identities))
	if !e.Identity.Empty() {
		args = append(args, "--identity", e.Identity.String())
	}
	for _, identity := range e.Identities {
		args = append(args, "--identity", identity.String())
	}
	return args
}

// parseIdentityFile parses the identities from indentityFile using the builtin
// age.
func parseIdentityFile(identityFile AbsPath) (identities []age.Identity, err error) {
	var file *os.File
	if file, err = os.Open(identityFile.String()); err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, file.Close())
	}()
	identities, err = age.ParseIdentities(file)
	return
}

// parseRecipientFile parses the recipients from recipientFile using the builtin
// age.
func parseRecipientsFile(recipientsFile AbsPath) (recipients []age.Recipient, err error) {
	var file *os.File
	if file, err = os.Open(recipientsFile.String()); err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, file.Close())
	}()
	recipients, err = age.ParseRecipients(file)
	return
}
