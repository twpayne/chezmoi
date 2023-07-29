package chezmoi

import (
	"github.com/rs/zerolog"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A DebugEncryption logs all calls to an Encryption.
type DebugEncryption struct {
	logger     *zerolog.Logger
	encryption Encryption
}

// NewDebugEncryption returns a new DebugEncryption that logs methods on
// encryption to logger.
func NewDebugEncryption(encryption Encryption, logger *zerolog.Logger) *DebugEncryption {
	return &DebugEncryption{
		logger:     logger,
		encryption: encryption,
	}
}

// Decrypt implements Encryption.Decrypt.
func (e *DebugEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := e.encryption.Decrypt(ciphertext)
	e.logger.Err(err).
		Bytes("ciphertext", chezmoilog.Output(ciphertext, err)).
		Bytes("plaintext", chezmoilog.Output(plaintext, err)).
		Msg("Decrypt")
	return plaintext, err
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *DebugEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	err := e.encryption.DecryptToFile(plaintextAbsPath, ciphertext)
	e.logger.Err(err).
		Stringer("plaintextAbsPath", plaintextAbsPath).
		Bytes("ciphertext", chezmoilog.Output(ciphertext, err)).
		Msg("DecryptToFile")
	return err
}

// Encrypt implements Encryption.Encrypt.
func (e *DebugEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := e.encryption.Encrypt(plaintext)
	e.logger.Err(err).
		Bytes("plaintext", chezmoilog.Output(plaintext, err)).
		Bytes("ciphertext", chezmoilog.Output(ciphertext, err)).
		Msg("Encrypt")
	return ciphertext, err
}

// EncryptFile implements Encryption.EncryptFile.
func (e *DebugEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	ciphertext, err := e.encryption.EncryptFile(plaintextAbsPath)
	e.logger.Err(err).
		Stringer("plaintextAbsPath", plaintextAbsPath).
		Bytes("ciphertext", chezmoilog.Output(ciphertext, err)).
		Msg("EncryptFile")
	return ciphertext, err
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *DebugEncryption) EncryptedSuffix() string {
	return e.encryption.EncryptedSuffix()
}
