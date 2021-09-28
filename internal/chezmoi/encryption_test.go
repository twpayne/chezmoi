package chezmoi

import (
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type xorEncryption struct {
	key byte
}

var _ Encryption = &xorEncryption{}

func (e *xorEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	return e.xorWithKey(ciphertext), nil
}

func (e *xorEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	return os.WriteFile(plaintextAbsPath.String(), e.xorWithKey(ciphertext), 0o666)
}

func (e *xorEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	return e.xorWithKey(plaintext), nil
}

func (e *xorEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	plaintext, err := os.ReadFile(plaintextAbsPath.String())
	if err != nil {
		return nil, err
	}
	return e.xorWithKey(plaintext), nil
}

func (e *xorEncryption) EncryptedSuffix() string {
	return ".xor"
}

func (e *xorEncryption) xorWithKey(input []byte) []byte {
	output := make([]byte, 0, len(input))
	for _, b := range input {
		output = append(output, b^e.key)
	}
	return output
}

func testEncryptionDecryptToFile(t *testing.T, encryption Encryption) {
	t.Helper()
	t.Run("DecryptToFile", func(t *testing.T) {
		expectedPlaintext := []byte("plaintext\n")

		actualCiphertext, err := encryption.Encrypt(expectedPlaintext)
		require.NoError(t, err)
		require.NotEmpty(t, actualCiphertext)
		assert.NotEqual(t, expectedPlaintext, actualCiphertext)

		tempDir, err := os.MkdirTemp("", "chezmoi-test-encryption")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tempDir))
		}()
		plaintextAbsPath := NewAbsPath(tempDir).Join("plaintext")

		require.NoError(t, encryption.DecryptToFile(plaintextAbsPath, actualCiphertext))

		actualPlaintext, err := os.ReadFile(plaintextAbsPath.String())
		require.NoError(t, err)
		require.NotEmpty(t, actualPlaintext)
		assert.Equal(t, expectedPlaintext, actualPlaintext)
	})
}

func testEncryptionEncryptDecrypt(t *testing.T, encryption Encryption) {
	t.Helper()
	t.Run("EncryptDecrypt", func(t *testing.T) {
		expectedPlaintext := []byte("plaintext\n")

		actualCiphertext, err := encryption.Encrypt(expectedPlaintext)
		require.NoError(t, err)
		require.NotEmpty(t, actualCiphertext)
		assert.NotEqual(t, expectedPlaintext, actualCiphertext)

		actualPlaintext, err := encryption.Decrypt(actualCiphertext)
		require.NoError(t, err)
		require.NotEmpty(t, actualPlaintext)
		assert.Equal(t, expectedPlaintext, actualPlaintext)
	})
}

func testEncryptionEncryptFile(t *testing.T, encryption Encryption) {
	t.Helper()
	t.Run("EncryptFile", func(t *testing.T) {
		expectedPlaintext := []byte("plaintext\n")

		tempDir, err := os.MkdirTemp("", "chezmoi-test-encryption")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tempDir))
		}()
		plaintextAbsPath := NewAbsPath(tempDir).Join("plaintext")
		require.NoError(t, os.WriteFile(plaintextAbsPath.String(), expectedPlaintext, 0o666))

		actualCiphertext, err := encryption.EncryptFile(plaintextAbsPath)
		require.NoError(t, err)
		require.NotEmpty(t, actualCiphertext)
		assert.NotEqual(t, expectedPlaintext, actualCiphertext)

		actualPlaintext, err := encryption.Decrypt(actualCiphertext)
		require.NoError(t, err)
		require.NotEmpty(t, actualPlaintext)
		assert.Equal(t, expectedPlaintext, actualPlaintext)
	})
}

func TestXOREncryption(t *testing.T) {
	xorEncryption := &xorEncryption{
		key: byte(rand.Int() + 1),
	}
	testEncryptionDecryptToFile(t, xorEncryption)
	testEncryptionEncryptDecrypt(t, xorEncryption)
	testEncryptionEncryptFile(t, xorEncryption)
}
