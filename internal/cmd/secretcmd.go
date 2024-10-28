package cmd

import "github.com/spf13/cobra"

type secretCmdConfig struct {
	keyring secretKeyringCmdConfig
}

func (c *Config) newSecretCmd() *cobra.Command {
	secretCmd := &cobra.Command{
		Use:     "secret",
		Args:    cobra.NoArgs,
		Short:   "Interact with a secret manager",
		Long:    mustLongHelp("secret"),
		Example: example("secret"),
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}

	if secretKeyringCmd := c.newSecretKeyringCmd(); secretKeyringCmd != nil {
		secretCmd.AddCommand(secretKeyringCmd)
	}

	return secretCmd
}
