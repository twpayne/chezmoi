package chezmoi

import (
	"log/slog"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A DebugEncryption logs all calls to an Encryption.
type DebugEncryption struct {
	logger     *slog.Logger
	encryption Encryption
}

// NewDebugEncryption returns a new DebugEncryption that logs methods on
// encryption to logger.
func NewDebugEncryption(encryption Encryption, logger *slog.Logger) *DebugEncryption {
	return &DebugEncryption{
		logger:     logger,
		encryption: encryption,
	}
}

// Decrypt implements Encryption.Decrypt.
func (e *DebugEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := e.encryption.Decrypt(ciphertext)
	chezmoilog.InfoOrError(e.logger, "Decrypt", err,
		chezmoilog.FirstFewBytes("ciphertext", ciphertext),
		chezmoilog.FirstFewBytes("plaintext", plaintext),
	)
	return plaintext, err
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *DebugEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	err := e.encryption.DecryptToFile(plaintextAbsPath, ciphertext)
	chezmoilog.InfoOrError(e.logger, "DecryptToFile", err,
		chezmoilog.Stringer("plaintextAbsPath", plaintextAbsPath),
		chezmoilog.FirstFewBytes("ciphertext", ciphertext),
	)
	return err
}

// Encrypt implements Encryption.Encrypt.
func (e *DebugEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	ciphertext, err := e.encryption.Encrypt(plaintext)
	chezmoilog.InfoOrError(e.logger, "Encrypt", err,
		chezmoilog.FirstFewBytes("plaintext", plaintext),
		chezmoilog.FirstFewBytes("ciphertext", ciphertext),
	)
	return ciphertext, err
}

// EncryptFile implements Encryption.EncryptFile.
func (e *DebugEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	ciphertext, err := e.encryption.EncryptFile(plaintextAbsPath)
	chezmoilog.InfoOrError(e.logger, "EncryptFile", err,
		chezmoilog.Stringer("plaintextAbsPath", plaintextAbsPath),
		chezmoilog.FirstFewBytes("ciphertext", ciphertext),
	)
	return ciphertext, err
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *DebugEncryption) EncryptedSuffix() string {
	return e.encryption.EncryptedSuffix()
}
