package chezmoicmd

import "github.com/spf13/cobra"

func (c *Config) newSecretCmd() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Args:    cobra.NoArgs,
		Short:   "Interact with a secret manager",
		Long:    mustLongHelp("secret"),
		Example: example("secret"),
	}

	secretCmd.AddCommand(c.newSecretKeyringCmd())

	return secretCmd
}
