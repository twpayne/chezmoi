package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newEncryptCommand() *cobra.Command {
	encryptCmd := &cobra.Command{
		Use:         "encrypt [file...]",
		Short:       "Encrypt file or standard input",
		Long:        mustLongHelp("encrypt"),
		Example:     example("encrypt"),
		RunE:        c.runEncryptCmd,
		Annotations: newAnnotations(),
	}

	return encryptCmd
}

func (c *Config) runEncryptCmd(cmd *cobra.Command, args []string) error {
	return c.filterInput(args, c.encryption.Encrypt)
}
