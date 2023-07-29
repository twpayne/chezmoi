package cmd

import (
	"github.com/spf13/cobra"
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
	return c.filterInput(args, c.encryption.Decrypt)
}
