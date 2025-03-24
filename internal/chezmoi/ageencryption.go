package chezmoi

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"filippo.io/age"
	"filippo.io/age/armor"

	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// An AgeEncryption uses age for encryption and decryption. See
// https://age-encryption.org.
type AgeEncryption struct {
	UseBuiltin      bool      `json:"useBuiltin"      mapstructure:"useBuiltin"      yaml:"useBuiltin"`
	Command         string    `json:"command"         mapstructure:"command"         yaml:"command"`
	Args            []string  `json:"args"            mapstructure:"args"            yaml:"args"`
	Identity        AbsPath   `json:"identity"        mapstructure:"identity"        yaml:"identity"`
	Identities      []AbsPath `json:"identities"      mapstructure:"identities"      yaml:"identities"`
	Passphrase      bool      `json:"passphrase"      mapstructure:"passphrase"      yaml:"passphrase"`
	Recipient       string    `json:"recipient"       mapstructure:"recipient"       yaml:"recipient"`
	Recipients      []string  `json:"recipients"      mapstructure:"recipients"      yaml:"recipients"`
	RecipientsFile  AbsPath   `json:"recipientsFile"  mapstructure:"recipientsFile"  yaml:"recipientsFile"`
	RecipientsFiles []AbsPath `json:"recipientsFiles" mapstructure:"recipientsFiles" yaml:"recipientsFiles"`
	Suffix          string    `json:"suffix"          mapstructure:"suffix"          yaml:"suffix"`
	Symmetric       bool      `json:"symmetric"       mapstructure:"symmetric"       yaml:"symmetric"`
}

// Decrypt implements Encryption.Decrypt.
func (e *AgeEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if e.UseBuiltin {
		return e.builtinDecrypt(ciphertext)
	}

	cmd := exec.Command(e.Command, append(e.decryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(slog.Default(), cmd)
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *AgeEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	if e.UseBuiltin {
		plaintext, err := e.builtinDecrypt(ciphertext)
		if err != nil {
			return err
		}
		return os.WriteFile(plaintextAbsPath.String(), plaintext, 0o644)
	}

	args := append(append(e.decryptArgs(), "--output", plaintextAbsPath.String()), e.Args...)
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = bytes.NewReader(ciphertext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(slog.Default(), cmd)
}

// Encrypt implements Encryption.Encrypt.
func (e *AgeEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	if e.UseBuiltin {
		return e.builtinEncrypt(plaintext)
	}

	cmd := exec.Command(e.Command, append(e.encryptArgs(), e.Args...)...)
	cmd.Stdin = bytes.NewReader(plaintext)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(slog.Default(), cmd)
}

// EncryptFile implements Encryption.EncryptFile.
func (e *AgeEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	if e.UseBuiltin {
		plaintext, err := os.ReadFile(plaintextAbsPath.String())
		if err != nil {
			return nil, err
		}
		return e.builtinEncrypt(plaintext)
	}

	args := append(append(e.encryptArgs(), e.Args...), plaintextAbsPath.String())
	cmd := exec.Command(e.Command, args...)
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdOutput(slog.Default(), cmd)
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
	var ciphertextReader io.Reader = bytes.NewReader(ciphertext)
	if bytes.HasPrefix(ciphertext, []byte(armor.Header)) {
		ciphertextReader = armor.NewReader(ciphertextReader)
	}
	plaintextReader, err := age.Decrypt(ciphertextReader, identities...)
	if err != nil {
		return nil, err
	}
	plaintextBuffer := &bytes.Buffer{}
	if _, err := io.Copy(plaintextBuffer, plaintextReader); err != nil {
		return nil, err
	}
	return plaintextBuffer.Bytes(), nil
}

// builtinEncrypt encrypts ciphertext using the builtin age.
func (e *AgeEncryption) builtinEncrypt(plaintext []byte) ([]byte, error) {
	recipients, err := e.builtinRecipients()
	if err != nil {
		return nil, err
	}
	ciphertextBuffer := &bytes.Buffer{}
	armoredCiphertextWriter := armor.NewWriter(ciphertextBuffer)
	ciphertextWriteCloser, err := age.Encrypt(armoredCiphertextWriter, recipients...)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(ciphertextWriteCloser, bytes.NewReader(plaintext)); err != nil {
		return nil, err
	}
	if err := ciphertextWriteCloser.Close(); err != nil {
		return nil, err
	}
	if err := armoredCiphertextWriter.Close(); err != nil {
		return nil, err
	}
	return ciphertextBuffer.Bytes(), nil
}

// builtinIdentities returns the identities for decryption using the builtin
// age.
func (e *AgeEncryption) builtinIdentities() ([]age.Identity, error) {
	var identities []age.Identity
	if !e.Identity.IsEmpty() {
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
	if !e.RecipientsFile.IsEmpty() {
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
		if !e.RecipientsFile.IsEmpty() {
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
	if !e.Identity.IsEmpty() {
		args = append(args, "--identity", e.Identity.String())
	}
	for _, identity := range e.Identities {
		args = append(args, "--identity", identity.String())
	}
	return args
}

// parseIdentityFile parses the identities from identityFile using the builtin
// age.
func parseIdentityFile(identityFile AbsPath) (identities []age.Identity, err error) {
	var file *os.File
	if file, err = os.Open(identityFile.String()); err != nil {
		return nil, err
	}
	defer chezmoierrors.CombineFunc(&err, file.Close)
	return age.ParseIdentities(file)
}

// parseRecipientsFile parses the recipients from recipientsFile using the
// builtin age.
func parseRecipientsFile(recipientsFile AbsPath) (recipients []age.Recipient, err error) {
	var file *os.File
	if file, err = os.Open(recipientsFile.String()); err != nil {
		return nil, err
	}
	defer chezmoierrors.CombineFunc(&err, file.Close)
	return age.ParseRecipients(file)
}
