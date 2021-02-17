package chezmoi

// An Encryption encrypts and decrypts files and data.
type Encryption interface {
	Decrypt(ciphertext []byte) ([]byte, error)
	DecryptToFile(filename string, ciphertext []byte) error
	Encrypt(plaintext []byte) ([]byte, error)
	EncryptFile(filename string) ([]byte, error)
}
