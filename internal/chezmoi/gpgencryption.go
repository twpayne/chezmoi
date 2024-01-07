package chezmoi

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A GPGEncryption uses gpg for encryption and decryption. See https://gnupg.org/.
type GPGEncryption struct {
	Command    string   `json:"command"    mapstructure:"command"    yaml:"command"`
	Recipient  string   `json:"recipient"  mapstructure:"recipient"  yaml:"recipient"`
	Suffix     string   `json:"suffix"     mapstructure:"suffix"     yaml:"suffix"`
	Args       []string `json:"args"       mapstructure:"args"       yaml:"args"`
	Recipients []string `json:"recipients" mapstructure:"recipients" yaml:"recipients"`
	Symmetric  bool     `json:"symmetric"  mapstructure:"symmetric"  yaml:"symmetric"`
}

// Decrypt implements Encryption.Decrypt.
func (e *GPGEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	var plaintext []byte
	if err := withPrivateTempDir(func(tempDirAbsPath AbsPath) error {
		ciphertextAbsPath := tempDirAbsPath.JoinString("ciphertext" + e.EncryptedSuffix())
		if err := os.WriteFile(ciphertextAbsPath.String(), ciphertext, 0o600); err != nil {
			return err
		}
		plaintextAbsPath := tempDirAbsPath.JoinString("plaintext")

		args := e.decryptArgs(plaintextAbsPath, ciphertextAbsPath)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		plaintext, err = os.ReadFile(plaintextAbsPath.String())
		return err
	}); err != nil {
		return nil, err
	}
	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *GPGEncryption) DecryptToFile(plaintextFilename AbsPath, ciphertext []byte) error {
	return withPrivateTempDir(func(tempDirAbsPath AbsPath) error {
		ciphertextAbsPath := tempDirAbsPath.JoinString("ciphertext" + e.EncryptedSuffix())
		if err := os.WriteFile(ciphertextAbsPath.String(), ciphertext, 0o600); err != nil {
			return err
		}
		args := e.decryptArgs(plaintextFilename, ciphertextAbsPath)
		return e.run(args)
	})
}

// Encrypt implements Encryption.Encrypt.
func (e *GPGEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	var ciphertext []byte
	if err := withPrivateTempDir(func(tempDirAbsPath AbsPath) error {
		plaintextAbsPath := tempDirAbsPath.JoinString("plaintext")
		if err := os.WriteFile(plaintextAbsPath.String(), plaintext, 0o600); err != nil {
			return err
		}
		ciphertextAbsPath := tempDirAbsPath.JoinString("ciphertext" + e.EncryptedSuffix())

		args := e.encryptArgs(plaintextAbsPath, ciphertextAbsPath)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		ciphertext, err = os.ReadFile(ciphertextAbsPath.String())
		return err
	}); err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (e *GPGEncryption) EncryptFile(plaintextFilename AbsPath) ([]byte, error) {
	var ciphertext []byte
	if err := withPrivateTempDir(func(tempDirAbsPath AbsPath) error {
		ciphertextAbsPath := tempDirAbsPath.JoinString("ciphertext" + e.EncryptedSuffix())

		args := e.encryptArgs(plaintextFilename, ciphertextAbsPath)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		ciphertext, err = os.ReadFile(ciphertextAbsPath.String())
		return err
	}); err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *GPGEncryption) EncryptedSuffix() string {
	return e.Suffix
}

// decryptArgs returns the arguments for decryption.
func (e *GPGEncryption) decryptArgs(plaintextFilename, ciphertextFilename AbsPath) []string {
	args := []string{"--output", plaintextFilename.String()}
	args = append(args, e.Args...)
	args = append(args, "--decrypt", ciphertextFilename.String())
	return args
}

// encryptArgs returns the arguments for encryption.
func (e *GPGEncryption) encryptArgs(plaintextFilename, ciphertextFilename AbsPath) []string {
	args := []string{
		"--armor",
		"--output", ciphertextFilename.String(),
	}
	if e.Symmetric {
		args = append(args, "--symmetric")
	} else {
		if e.Recipient != "" {
			args = append(args, "--recipient", e.Recipient)
		}
		for _, recipient := range e.Recipients {
			args = append(args, "--recipient", recipient)
		}
	}
	args = append(args, e.Args...)
	if !e.Symmetric {
		args = append(args, "--encrypt")
	}
	args = append(args, plaintextFilename.String())
	return args
}

// run runs the command with args.
func (e *GPGEncryption) run(args []string) error {
	cmd := exec.Command(e.Command, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(cmd)
}

// withPrivateTempDir creates a private temporary and calls f.
func withPrivateTempDir(f func(tempDirAbsPath AbsPath) error) (err error) {
	var tempDir string
	if tempDir, err = os.MkdirTemp("", "chezmoi-encryption"); err != nil {
		return
	}
	defer chezmoierrors.CombineFunc(&err, func() error {
		return os.RemoveAll(tempDir)
	})
	if runtime.GOOS != "windows" {
		if err = os.Chmod(tempDir, 0o700); err != nil {
			return
		}
	}

	err = f(NewAbsPath(tempDir))
	return
}
