package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newDecryptCommand() *cobra.Command {
	decryptCmd := &cobra.Command{
		Use:         "decrypt [file...]",
		Short:       "Decrypt file or standard input",
		Long:        mustLongHelp("decrypt"),
		Example:     example("decrypt"),
		RunE:        c.runDecryptCmd,
		Annotations: newAnnotations(),
	}

	return decryptCmd
}

func (c *Config) runDecryptCmd(cmd *cobra.Command, args []string) error {
	return c.filterInput(args, c.encryption.Decrypt)
}
