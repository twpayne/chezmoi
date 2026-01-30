package chezmoi

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strings"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// SSHEncryption implements Encryption using an SSH agent.
type SSHEncryption struct {
	Identity string `json:"identity" mapstructure:"identity" yaml:"identity"`

	agentFactory func() (agent.Agent, error) // For testing
}

const (
	sshEncryptionHeaderMagic     = "CHEZMOISSHv1"
	sshEncryptionSaltSize        = 32
	sshEncryptionChallengeSize   = 32
	sshEncryptionNonceSize       = chacha20poly1305.NonceSizeX
	sshEncryptionFingerprintSize = 32 // SHA256 of public key
	sshEncryptionSignaturePrefix = "chezmoi-ssh-signature-v1:"
)

var sshEncryptionSupportedAlgorithms = []string{
	ssh.KeyAlgoED25519,
	ssh.KeyAlgoRSA,
	"rsa-sha2-256",
	"rsa-sha2-512",
	ssh.KeyAlgoECDSA256,
	ssh.KeyAlgoECDSA384,
	ssh.KeyAlgoECDSA521,
}

// Decrypt implements Encryption.Decrypt.
func (e *SSHEncryption) Decrypt(ciphertext []byte) ([]byte, error) {
	if !bytes.HasPrefix(ciphertext, []byte(sshEncryptionHeaderMagic)) {
		return nil, errors.New("ssh: invalid header")
	}

	reader := bytes.NewReader(ciphertext[len(sshEncryptionHeaderMagic):])

	salt := make([]byte, sshEncryptionSaltSize)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return nil, err
	}

	challenge := make([]byte, sshEncryptionChallengeSize)
	if _, err := io.ReadFull(reader, challenge); err != nil {
		return nil, err
	}

	nonce := make([]byte, sshEncryptionNonceSize)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return nil, err
	}

	fingerprint := make([]byte, sshEncryptionFingerprintSize)
	if _, err := io.ReadFull(reader, fingerprint); err != nil {
		return nil, err
	}

	// Connect to agent
	a, err := e.connectAgent()
	if err != nil {
		return nil, err
	}

	// Parse configured public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(e.Identity)) //nolint:dogsled
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to parse identity: %w", err)
	}

	// Verify fingerprint matches configured key (fast fail)
	pubKeyFingerprint := sha256.Sum256(pubKey.Marshal())
	if !bytes.Equal(fingerprint, pubKeyFingerprint[:]) {
		return nil, errors.New("ssh: encrypted data size does not match configured identity")
	}

	// Derive key
	key, err := e.deriveKey(a, pubKey, challenge, salt)
	if err != nil {
		return nil, err
	}

	// Decrypt
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	encryptedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	plaintext, err := aead.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, fmt.Errorf("ssh: decryption failed: %w", err)
	}

	return plaintext, nil
}

// DecryptToFile implements Encryption.DecryptToFile.
func (e *SSHEncryption) DecryptToFile(plaintextAbsPath AbsPath, ciphertext []byte) error {
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return err
	}
	return os.WriteFile(plaintextAbsPath.String(), plaintext, 0o644)
}

// Encrypt implements Encryption.Encrypt.
func (e *SSHEncryption) Encrypt(plaintext []byte) ([]byte, error) {
	if e.Identity == "" {
		return nil, errors.New("ssh: missing identity configuration")
	}

	// Connect to agent
	a, err := e.connectAgent()
	if err != nil {
		return nil, err
	}

	// Parse configured public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(e.Identity)) //nolint:dogsled
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to parse identity: %w", err)
	}

	// Generate salt, challenge, nonce
	salt := make([]byte, sshEncryptionSaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	challenge := make([]byte, sshEncryptionChallengeSize)
	if _, err := rand.Read(challenge); err != nil {
		return nil, err
	}
	nonce := make([]byte, sshEncryptionNonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Derive key
	key, err := e.deriveKey(a, pubKey, challenge, salt)
	if err != nil {
		return nil, err
	}

	// Encrypt
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// Construct output
	var buf bytes.Buffer
	buf.WriteString(sshEncryptionHeaderMagic)
	buf.Write(salt)
	buf.Write(challenge)
	buf.Write(nonce)

	// Add fingerprint of the using key
	fingerprint := sha256.Sum256(pubKey.Marshal())
	buf.Write(fingerprint[:])

	buf.Write(ciphertext)

	return buf.Bytes(), nil
}

// EncryptFile implements Encryption.EncryptFile.
func (e *SSHEncryption) EncryptFile(plaintextAbsPath AbsPath) ([]byte, error) {
	plaintext, err := os.ReadFile(plaintextAbsPath.String())
	if err != nil {
		return nil, err
	}
	return e.Encrypt(plaintext)
}

// EncryptedSuffix implements Encryption.EncryptedSuffix.
func (e *SSHEncryption) EncryptedSuffix() string {
	return ".ssh"
}

func (e *SSHEncryption) connectAgent() (agent.Agent, error) {
	if e.agentFactory != nil {
		return e.agentFactory()
	}
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, errors.New("ssh: SSH_AUTH_SOCK not set")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to connect to agent: %w", err)
	}
	return agent.NewClient(conn), nil
}

func (e *SSHEncryption) deriveKey(a agent.Agent, pubKey ssh.PublicKey, challenge, salt []byte) ([]byte, error) {
	// Sign the challenge to prove possession and get input for KDF
	// We use the signature blob as the KDF input.
	// NOTE: We rely on the agent to support signing.

	// Construct data to sign.
	// We add a prefix to ensure we don't accidentally sign something meaningful in another protocol.
	dataToSign := append([]byte(sshEncryptionSignaturePrefix), challenge...)
	sig, err := a.Sign(pubKey, dataToSign)
	if err != nil {
		return nil, fmt.Errorf("ssh: failed to sign challenge: %w", err)
	}

	// Verify algorithm is allowed
	if !slices.Contains(sshEncryptionSupportedAlgorithms, sig.Format) {
		return nil, fmt.Errorf("ssh: unsupported signature algorithm: %s (supported: %s)",
			sig.Format,
			strings.Join(sshEncryptionSupportedAlgorithms, ", "),
		)
	}

	// We treat the signature blob as the secret.
	// Scrypt parameters: N=32768, r=8, p=1
	key, err := scrypt.Key(sig.Blob, salt, 32768, 8, 1, chacha20poly1305.KeySize)
	if err != nil {
		return nil, fmt.Errorf("ssh: kdf failed: %w", err)
	}
	return key, nil
}
