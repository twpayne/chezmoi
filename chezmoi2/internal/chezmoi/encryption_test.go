package chezmoi

import (
	"io/ioutil"
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

func (t *xorEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	return t.xorWithKey(ciphertext), nil
}

func (t *xorEncryption) DecryptToFile(filename string, ciphertext []byte) error {
	return ioutil.WriteFile(filename, t.xorWithKey(ciphertext), 0o666)
}

func (t *xorEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	return t.xorWithKey(plaintext), nil
}

func (t *xorEncryption) EncryptFile(filename string) ([]byte, error) {
	plaintext, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return t.xorWithKey(plaintext), nil
}

func (t *xorEncryption) xorWithKey(input []byte) []byte {
	output := make([]byte, 0, len(input))
	for _, b := range input {
		output = append(output, b^t.key)
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

		tempDir, err := ioutil.TempDir("", "chezmoi-test-encryption")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tempDir))
		}()
		filename := filepath.Join(tempDir, "filename")

		require.NoError(t, encryption.DecryptToFile(filename, actualCiphertext))

		actualPlaintext, err := ioutil.ReadFile(filename)
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

		tempDir, err := ioutil.TempDir("", "chezmoi-test-encryption")
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, os.RemoveAll(tempDir))
		}()
		filename := filepath.Join(tempDir, "filename")
		require.NoError(t, ioutil.WriteFile(filename, expectedPlaintext, 0o666))

		actualCiphertext, err := encryption.EncryptFile(filename)
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
