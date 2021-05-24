package chezmoi

// An Encryption encrypts and decrypts files and data.
type Encryption interface {
	Decrypt(ciphertext []byte) ([]byte, error)
	DecryptToFile(plaintextFilename string, ciphertext []byte) error
	Encrypt(plaintext []byte) ([]byte, error)
	EncryptFile(plaintextFilename string) ([]byte, error)
	EncryptedSuffix() string
}
