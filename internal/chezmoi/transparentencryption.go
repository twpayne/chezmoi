package chezmoi

import (
	"os"

	"github.com/google/renameio/v2/maybe"
)

// A TransparentEncryption assumes that encryption is handled transparently.
// Encrypted files still get the encrypted_ attribute, but no encryption occurs
// in chezmoi.
type TransparentEncryption struct{}

// Decrypt implements Encryption.Decrypt.
func (t TransparentEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (t TransparentEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	return maybe.WriteFile(plaintextAbsPath.String(), ciphertext, 0o600)
}

// Encrypt implements Encryption.Encrypt.
func (t TransparentEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	return plaintext, nil
}

// EncryptFile implements Encryption.EncryptFile.
func (t TransparentEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	return os.ReadFile(plaintextAbsPath.String())
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (t TransparentEncryption) EncryptedSuffix() string {
	return ""
}
