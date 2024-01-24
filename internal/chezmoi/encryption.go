package chezmoi

// An Encryption encrypts and decrypts files and data.
type Encryption interface {
	Decrypt(ciphertext []byte) ([]byte, error)
	DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error
	Encrypt(plaintext []byte) ([]byte, error)
	EncryptFile(plaintextAbsPath AbsPath) ([]byte, error)
	EncryptedSuffix() string
}
