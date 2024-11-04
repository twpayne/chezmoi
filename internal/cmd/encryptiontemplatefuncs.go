package cmd

func (c *Config) decryptTemplateFunc(ciphertext string) string {
	return string(mustValue(c.encryption.Decrypt([]byte(ciphertext))))
}

func (c *Config) encryptTemplateFunc(plaintext string) string {
	return string(mustValue(c.encryption.Encrypt([]byte(plaintext))))
}
