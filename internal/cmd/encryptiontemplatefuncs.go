package cmd

import "chezmoi.io/chezmoi/v2/internal/chezmoi"

func (c *Config) decryptTemplateFunc(ciphertext string) string {
	chezmoi.SkipTemplateIf(c.skipSecrets)

	return string(mustValue(c.encryption.Decrypt([]byte(ciphertext))))
}

func (c *Config) encryptTemplateFunc(plaintext string) string {
	chezmoi.SkipTemplateIf(c.skipSecrets)

	return string(mustValue(c.encryption.Encrypt([]byte(plaintext))))
}
