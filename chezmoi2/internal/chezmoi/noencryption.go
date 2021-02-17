package chezmoi

import "errors"

var errNoEncryption = errors.New("no encryption")

// NoEncryption returns an error when any method is called.
type NoEncryption struct{}

// Decrypt implements Encryption.Decrypt.
func (NoEncryption) Decrypt([]byte) ([]byte, error) { return nil, errNoEncryption }

// DecryptToFile implements Encryption.DecryptToFile.
func (NoEncryption) DecryptToFile(string, []byte) error { return errNoEncryption }

// Encrypt implements Encryption.Encrypt.
func (NoEncryption) Encrypt([]byte) ([]byte, error) { return nil, errNoEncryption }

// EncryptFile implements Encryption.EncryptFile.
func (NoEncryption) EncryptFile(string) ([]byte, error) { return nil, errNoEncryption }
