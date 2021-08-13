package chezmoi

// An Encryption encrypts and decrypts files and data.
type Encryption interface {
	Decrypt(ciphertext []byte) ([]byte, error)
	DecryptToFile(plaintextFilename AbsPath, ciphertext []byte) error
	Encrypt(plaintext []byte) ([]byte, error)
	EncryptFile(plaintextFilename AbsPath) ([]byte, error)
	EncryptedSuffix() string
}
