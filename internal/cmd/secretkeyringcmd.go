package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type secretKeyringCmdConfig struct {
	get secretKeyringGetCmdConfig
	set secretKeyringSetCmdConfig
}

type secretKeyringGetCmdConfig struct {
	service string
	user    string
}

type secretKeyringSetCmdConfig struct {
	service string
	user    string
	value   string
}

func (c *Config) newSecretKeyringCmd() *cobra.Command {
	keyringCmd := &cobra.Command{
		Use:   "keyring",
		Args:  cobra.NoArgs,
		Short: "Interact with keyring",
	}

	keyringGetCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.NoArgs,
		Short: "Get a value from keyring",
		RunE:  c.runKeyringGetCmdE,
	}
	secretKeyringGetPersistentFlags := keyringGetCmd.PersistentFlags()
	secretKeyringGetPersistentFlags.StringVar(&c.secretKeyring.get.service, "service", "", "service")
	secretKeyringGetPersistentFlags.StringVar(&c.secretKeyring.get.user, "user", "", "user")
	markPersistentFlagsRequired(keyringGetCmd, "service", "user")
	keyringCmd.AddCommand(keyringGetCmd)

	keyringSetCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.NoArgs,
		Short: "Set a value in keyring",
		RunE:  c.runKeyringSetCmdE,
	}
	secretKeyringSetPersistentFlags := keyringSetCmd.PersistentFlags()
	secretKeyringSetPersistentFlags.StringVar(&c.secretKeyring.set.service, "service", "", "service")
	secretKeyringSetPersistentFlags.StringVar(&c.secretKeyring.set.user, "user", "", "user")
	secretKeyringSetPersistentFlags.StringVar(&c.secretKeyring.set.value, "value", "", "value")
	markPersistentFlagsRequired(keyringSetCmd, "service", "user")
	keyringCmd.AddCommand(keyringSetCmd)

	return keyringCmd
}

func (c *Config) runKeyringGetCmdE(cmd *cobra.Command, args []string) error {
	value, err := keyring.Get(c.secretKeyring.get.service, c.secretKeyring.get.user)
	if err != nil {
		return err
	}
	return c.writeOutputString(value)
}

func (c *Config) runKeyringSetCmdE(cmd *cobra.Command, args []string) error {
	value := c.secretKeyring.set.value
	if value == "" {
		var err error
		value, err = c.readPassword("Value: ")
		if err != nil {
			return err
		}
	}
	return keyring.Set(c.secretKeyring.set.service, c.secretKeyring.set.user, value)
}
