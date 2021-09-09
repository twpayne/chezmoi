package cmd

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newDecryptCommand() *cobra.Command {
	decryptCommand := &cobra.Command{
		Use:     "decrypt [file...]",
		Short:   "Decrypt file or standard input",
		Long:    mustLongHelp("decrypt"),
		Example: example("decrypt"),
		RunE:    c.runDecryptCmd,
	}

	return decryptCommand
}

func (c *Config) runDecryptCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		ciphertext, err := io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		plaintext, err := c.encryption.Decrypt(ciphertext)
		if err != nil {
			return err
		}
		return c.writeOutput(plaintext)
	}

	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return err
		}
		ciphertext, err := c.baseSystem.ReadFile(argAbsPath)
		if err != nil {
			return err
		}
		plaintext, err := c.encryption.Decrypt(ciphertext)
		if err != nil {
			return err
		}
		if err := c.writeOutput(plaintext); err != nil {
			return err
		}
	}

	return nil
}
