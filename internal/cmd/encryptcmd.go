package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newEncryptCommand() *cobra.Command {
	decryptCommand := &cobra.Command{
		Use:     "encrypt [file...]",
		Short:   "Encrypt file or standard input",
		Long:    mustLongHelp("encrypt"),
		Example: example("encrypt"),
		RunE:    c.runEncryptCmd,
	}

	return decryptCommand
}

func (c *Config) runEncryptCmd(cmd *cobra.Command, args []string) error {
	return c.filterInput(args, c.encryption.Encrypt)
}
