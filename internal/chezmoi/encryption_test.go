package chezmoi

import (
	"math/rand"
	"os"
	"path/filepath"
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

func (e *xorEncryption) DecryptToFile(plaintextFilename string, ciphertext []byte) error {
	return os.WriteFile(plaintextFilename, e.xorWithKey(ciphertext), 0o666)
}

func (e *xorEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	return e.xorWithKey(plaintext), nil
}

func (e *xorEncryption) EncryptFile(plaintextFilename string) ([]byte, error) {
	plaintext, err := os.ReadFile(plaintextFilename)
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
		plaintextFilename := filepath.Join(tempDir, "plaintext")

		require.NoError(t, encryption.DecryptToFile(plaintextFilename, actualCiphertext))

		actualPlaintext, err := os.ReadFile(plaintextFilename)
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
		plaintextFilename := filepath.Join(tempDir, "plaintext")
		require.NoError(t, os.WriteFile(plaintextFilename, expectedPlaintext, 0o666))

		actualCiphertext, err := encryption.EncryptFile(plaintextFilename)
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
