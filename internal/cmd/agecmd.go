package cmd

import (
	"bytes"
	"errors"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
	"github.com/spf13/cobra"
)

type ageDecryptCmdConfig struct {
	passphrase bool
}

type ageEncryptCmdConfig struct {
	passphrase bool
}

type ageCmdConfig struct {
	decrypt ageDecryptCmdConfig
	encrypt ageEncryptCmdConfig
}

func (c *Config) newAgeCmd() *cobra.Command {
	ageCmd := &cobra.Command{
		Use:   "age",
		Args:  cobra.NoArgs,
		Short: "Interact with age",
	}

	ageDecryptCmd := &cobra.Command{
		Use:   "decrypt [file...]",
		Short: "Decrypt file or standard input",
		RunE:  c.runAgeDecryptCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	ageDecryptCmd.Flags().
		BoolVarP(&c.age.decrypt.passphrase, "passphrase", "p", c.age.decrypt.passphrase, "Decrypt with a passphrase")
	ageCmd.AddCommand(ageDecryptCmd)

	ageEncryptCmd := &cobra.Command{
		Use:   "encrypt [file...]",
		Short: "Encrypt file or standard input",
		RunE:  c.runAgeEncryptCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}
	ageEncryptCmd.Flags().
		BoolVarP(&c.age.encrypt.passphrase, "passphrase", "p", c.age.encrypt.passphrase, "Encrypt with a passphrase")
	ageCmd.AddCommand(ageEncryptCmd)

	return ageCmd
}

func (c *Config) runAgeDecryptCmd(cmd *cobra.Command, args []string) error {
	if !c.age.decrypt.passphrase {
		return errors.New("only passphrase encryption is supported")
	}
	decrypt := func(ciphertext []byte) ([]byte, error) {
		var ciphertextReader io.Reader = bytes.NewReader(ciphertext)
		if bytes.HasPrefix(ciphertext, []byte(armor.Header)) {
			ciphertextReader = armor.NewReader(ciphertextReader)
		}
		identity := &LazyScryptIdentity{
			Passphrase: func() (string, error) {
				return c.readPassword("Enter passphrase: ")
			},
		}
		plaintextReader, err := age.Decrypt(ciphertextReader, identity)
		if err != nil {
			return nil, err
		}
		plaintextBuffer := &bytes.Buffer{}
		if _, err := io.Copy(plaintextBuffer, plaintextReader); err != nil {
			return nil, err
		}
		return plaintextBuffer.Bytes(), nil
	}
	return c.filterInput(args, decrypt)
}

func (c *Config) runAgeEncryptCmd(cmd *cobra.Command, args []string) error {
	if !c.age.encrypt.passphrase {
		return errors.New("only passphrase encryption is supported")
	}
	passphrase, err := c.readPassword("Enter passphrase: ")
	if err != nil {
		return err
	}
	confirmPassphrase, err := c.readPassword("Confirm passphrase: ")
	if err != nil {
		return err
	}
	if passphrase != confirmPassphrase {
		return errors.New("passphrases didn't match")
	}
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return err
	}
	encrypt := func(plaintext []byte) ([]byte, error) {
		ciphertextBuffer := &bytes.Buffer{}
		armoredCiphertextWriter := armor.NewWriter(ciphertextBuffer)
		ciphertextWriteCloser, err := age.Encrypt(armoredCiphertextWriter, recipient)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(ciphertextWriteCloser, bytes.NewReader(plaintext)); err != nil {
			return nil, err
		}
		if err := ciphertextWriteCloser.Close(); err != nil {
			return nil, err
		}
		if err := armoredCiphertextWriter.Close(); err != nil {
			return nil, err
		}
		return ciphertextBuffer.Bytes(), nil
	}
	return c.filterInput(args, encrypt)
}
