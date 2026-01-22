package chezmoi

import "errors"

var errEncryptionNotConfigured = errors.New("encryption not configured")

// NoEncryption returns an error when any method is called.
type NoEncryption struct{}

// Decrypt implements Encryption.Decrypt.
func (NoEncryption) Decrypt([]byte) ([]byte, error) { return nil, errEncryptionNotConfigured }

// DecryptToFile implements Encryption.DecryptToFile.
func (NoEncryption) DecryptToFile(AbsPath, []byte) error { return errEncryptionNotConfigured }

// Encrypt implements Encryption.Encrypt.
func (NoEncryption) Encrypt([]byte) ([]byte, error) { return nil, errEncryptionNotConfigured }

// EncryptFile implements Encryption.EncryptFile.
func (NoEncryption) EncryptFile(AbsPath) ([]byte, error) { return nil, errEncryptionNotConfigured }

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (NoEncryption) EncryptedSuffix() string { return "" }
