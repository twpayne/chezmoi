package chezmoi

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
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
	var plaintext []byte
	if err := withPrivateTempDir(func(tempDir string) error {
		ciphertextFilename := filepath.Join(tempDir, "ciphertext"+e.EncryptedSuffix())
		if err := os.WriteFile(ciphertextFilename, ciphertext, 0o600); err != nil {
			return err
		}
		plaintextFilename := filepath.Join(tempDir, "plaintext")

		args := e.decryptArgs(plaintextFilename, ciphertextFilename)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		plaintext, err = os.ReadFile(plaintextFilename)
		return err
	}); err != nil {
		return nil, err
	}
	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *GPGEncryption) DecryptToFile(plaintextFilename string, ciphertext []byte) error {
	return withPrivateTempDir(func(tempDir string) error {
		ciphertextFilename := filepath.Join(tempDir, "ciphertext"+e.EncryptedSuffix())
		if err := os.WriteFile(ciphertextFilename, ciphertext, 0o600); err != nil {
			return err
		}
		args := e.decryptArgs(plaintextFilename, ciphertextFilename)
		return e.run(args)
	})
}

// Encrypt implements Encryption.Encrypt.
func (e *GPGEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	var ciphertext []byte
	if err := withPrivateTempDir(func(tempDir string) error {
		plaintextFilename := filepath.Join(tempDir, "plaintext")
		if err := os.WriteFile(plaintextFilename, plaintext, 0o600); err != nil {
			return err
		}
		ciphertextFilename := filepath.Join(tempDir, "ciphertext"+e.EncryptedSuffix())

		args := e.encryptArgs(plaintextFilename, ciphertextFilename)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		ciphertext, err = os.ReadFile(ciphertextFilename)
		return err
	}); err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (e *GPGEncryption) EncryptFile(plaintextFilename string) ([]byte, error) {
	var ciphertext []byte
	if err := withPrivateTempDir(func(tempDir string) error {
		ciphertextFilename := filepath.Join(tempDir, "ciphertext"+e.EncryptedSuffix())

		args := e.encryptArgs(plaintextFilename, ciphertextFilename)
		if err := e.run(args); err != nil {
			return err
		}

		var err error
		ciphertext, err = os.ReadFile(ciphertextFilename)
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

func (e *GPGEncryption) decryptArgs(plaintextFilename, ciphertextFilename string) []string {
	args := []string{"--output", plaintextFilename}
	args = append(args, e.Args...)
	args = append(args, "--decrypt", ciphertextFilename)
	return args
}

func (e *GPGEncryption) encryptArgs(plaintextFilename, ciphertextFilename string) []string {
	args := []string{
		"--armor",
		"--output", ciphertextFilename,
	}
	if e.Symmetric {
		args = append(args, "--symmetric")
	} else if e.Recipient != "" {
		args = append(args, "--recipient", e.Recipient)
	}
	args = append(args, e.Args...)
	if !e.Symmetric {
		args = append(args, "--encrypt")
	}
	args = append(args, plaintextFilename)
	return args
}

func (e *GPGEncryption) run(args []string) error {
	//nolint:gosec
	cmd := exec.Command(e.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// withPrivateTempDir creates a private temporary and calls f.
func withPrivateTempDir(f func(tempDir string) error) error {
	tempDir, err := os.MkdirTemp("", "chezmoi-encryption")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempDir, 0o700); err != nil {
			return err
		}
	}

	return f(tempDir)
}
