package chezmoi

import (
	"github.com/rs/zerolog/log"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A DebugEncryption logs all calls to an Encryption.
type DebugEncryption struct {
	encryption Encryption
}

// NewDebugEncryption returns a new DebugEncryption.
func NewDebugEncryption(encryption Encryption) *DebugEncryption {
	return &DebugEncryption{
		encryption: encryption,
	}
}

// Decrypt implements Encryption.Decrypt.
func (e *DebugEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := e.encryption.Decrypt(ciphertext)
	log.Debug().
		Bytes("ciphertext", chezmoilog.FirstFewBytes(ciphertext)).
		Err(err).
		Bytes("plaintext", chezmoilog.FirstFewBytes(plaintext)).
		Msg("Decrypt")
	return plaintext, err
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *DebugEncryption) DecryptToFile(filename string, ciphertext []byte) error {
	err := e.encryption.DecryptToFile(filename, ciphertext)
	log.Debug().
		Str("filename", filename).
		Bytes("ciphertext", chezmoilog.FirstFewBytes(ciphertext)).
		Err(err).
		Msg("DecryptToFile")
	return err
}

// Encrypt implements Encryption.Encrypt.
func (e *DebugEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := e.encryption.Encrypt(plaintext)
	log.Debug().
		Bytes("plaintext", chezmoilog.FirstFewBytes(plaintext)).
		Err(err).
		Bytes("ciphertext", chezmoilog.FirstFewBytes(ciphertext)).
		Msg("Encrypt")
	return ciphertext, err
}

// EncryptFile implements Encryption.EncryptFile.
func (e *DebugEncryption) EncryptFile(filename string) ([]byte, error) {
	ciphertext, err := e.encryption.EncryptFile(filename)
	log.Debug().
		Str("filename", filename).
		Err(err).
		Bytes("ciphertext", chezmoilog.FirstFewBytes(ciphertext)).
		Msg("EncryptFile")
	return ciphertext, err
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *DebugEncryption) EncryptedSuffix() string {
	return e.encryption.EncryptedSuffix()
}
