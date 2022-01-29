package cmd

func (c *Config) decryptTemplateFunc(ciphertext string) string {
	plaintextBytes, err := c.encryption.Decrypt([]byte(ciphertext))
	if err != nil {
		raiseTemplateError(err)
		return ""
	}
	return string(plaintextBytes)
}

func (c *Config) encryptTemplateFunc(plaintext string) string {
	ciphertextBytes, err := c.encryption.Encrypt([]byte(plaintext))
	if err != nil {
		raiseTemplateError(err)
		return ""
	}
	return string(ciphertextBytes)
}
