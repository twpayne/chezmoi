package chezmoi

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/alecthomas/assert/v2"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func TestSSHEncryption(t *testing.T) {
	// Create keys
	pubEd, privEd, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	privRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	// Setup mock agent
	keyring := agent.NewKeyring()

	// Add Ed25519 key
	err = keyring.Add(agent.AddedKey{
		PrivateKey: privEd,
	})
	assert.NoError(t, err)

	// Add RSA key
	err = keyring.Add(agent.AddedKey{
		PrivateKey: privRSA,
	})
	assert.NoError(t, err)

	// Helper to get formatted public key string
	getIdentity := func(k any) string {
		sshKey, err := ssh.NewPublicKey(k)
		if err != nil {
			panic(err)
		}
		return string(ssh.MarshalAuthorizedKey(sshKey))
	}

	tests := []struct {
		name     string
		identity string
		payload  string
	}{
		{
			name:     "ed25519",
			identity: getIdentity(pubEd),
			payload:  "secret data ed25519",
		},
		{
			name:     "rsa",
			identity: getIdentity(&privRSA.PublicKey),
			payload:  "secret data rsa",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Re-instantiate struct for each test to be clean
			// Since we use a factory, we can assume the agent is "connected"
			enc := &SSHEncryption{
				Identity: tc.identity,
				agentFactory: func() (agent.Agent, error) {
					return keyring, nil
				},
			}

			// Encrypt
			plaintext := []byte(tc.payload)
			ciphertext, err := enc.Encrypt(plaintext)
			assert.NoError(t, err)

			// Decrypt
			decrypted, err := enc.Decrypt(ciphertext)
			assert.NoError(t, err)

			assert.Equal(t, plaintext, decrypted)
		})
	}
}

func TestSSHEncryption_Failures(t *testing.T) {
	// Generate two keys
	pub1, priv1, _ := ed25519.GenerateKey(rand.Reader)
	pub2, _, _ := ed25519.GenerateKey(rand.Reader)

	keyring := agent.NewKeyring()
	_ = keyring.Add(agent.AddedKey{PrivateKey: priv1})

	sshPub1, _ := ssh.NewPublicKey(pub1)
	identity1 := string(ssh.MarshalAuthorizedKey(sshPub1))

	sshPub2, _ := ssh.NewPublicKey(pub2)
	identity2 := string(ssh.MarshalAuthorizedKey(sshPub2))

	t.Run("IdentityNotInAgent", func(t *testing.T) {
		enc := &SSHEncryption{
			Identity:     identity2, // Key 2 not in agent
			agentFactory: func() (agent.Agent, error) { return keyring, nil },
		}
		_, err := enc.Encrypt([]byte("data"))
		assert.Error(t, err) // Should fail during Sign
	})

	t.Run("DecryptWithWrongIdentity", func(t *testing.T) {
		// Encrypt with Key 1
		enc1 := &SSHEncryption{
			Identity:     identity1,
			agentFactory: func() (agent.Agent, error) { return keyring, nil },
		}
		ciphertext, _ := enc1.Encrypt([]byte("data"))

		// Try to decrypt with Key 2 config (simulating user changing config)
		enc2 := &SSHEncryption{
			Identity:     identity2,
			agentFactory: func() (agent.Agent, error) { return keyring, nil },
		}
		_, err := enc2.Decrypt(ciphertext)
		assert.Error(t, err)
		// Should be specific error about fingerprint mismatch
		assert.Contains(t, err.Error(), "does not match configured identity")
	})
}
