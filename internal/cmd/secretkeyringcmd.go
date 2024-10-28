//go:build !freebsd || (freebsd && cgo)

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type secretKeyringCmdConfig struct {
	delete secretKeyringDeleteCmdConfig
	get    secretKeyringGetCmdConfig
	set    secretKeyringSetCmdConfig
}

type secretKeyringDeleteCmdConfig struct {
	service string
	user    string
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
	secretKeyringCmd := &cobra.Command{
		Use:   "keyring",
		Args:  cobra.NoArgs,
		Short: "Interact with keyring",
		Annotations: newAnnotations(
			persistentStateModeNone,
		),
	}

	secretKeyringDeleteCmd := &cobra.Command{
		Use:   "delete",
		Args:  cobra.NoArgs,
		Short: "Delete a value from keyring",
		RunE:  c.runSecretKeyringDeleteCmdE,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	secretKeyringDeleteCmd.Flags().StringVar(&c.secret.keyring.delete.service, "service", "", "service")
	secretKeyringDeleteCmd.Flags().StringVar(&c.secret.keyring.delete.user, "user", "", "user")
	markFlagsRequired(secretKeyringDeleteCmd, "service", "user")
	secretKeyringCmd.AddCommand(secretKeyringDeleteCmd)

	secretKeyringGetCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.NoArgs,
		Short: "Get a value from keyring",
		RunE:  c.runSecretKeyringGetCmdE,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	secretKeyringGetCmd.Flags().StringVar(&c.secret.keyring.get.service, "service", "", "service")
	secretKeyringGetCmd.Flags().StringVar(&c.secret.keyring.get.user, "user", "", "user")
	markFlagsRequired(secretKeyringGetCmd, "service", "user")
	secretKeyringCmd.AddCommand(secretKeyringGetCmd)

	secretKeyringSetCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.NoArgs,
		Short: "Set a value in keyring",
		RunE:  c.runSecretKeyringSetCmdE,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}
	secretKeyringSetCmd.Flags().StringVar(&c.secret.keyring.set.service, "service", "", "service")
	secretKeyringSetCmd.Flags().StringVar(&c.secret.keyring.set.user, "user", "", "user")
	secretKeyringSetCmd.Flags().StringVar(&c.secret.keyring.set.value, "value", "", "value")
	markFlagsRequired(secretKeyringSetCmd, "service", "user")
	secretKeyringCmd.AddCommand(secretKeyringSetCmd)

	return secretKeyringCmd
}

func (c *Config) runSecretKeyringDeleteCmdE(cmd *cobra.Command, args []string) error {
	return keyring.Delete(c.secret.keyring.delete.service, c.secret.keyring.delete.user)
}

func (c *Config) runSecretKeyringGetCmdE(cmd *cobra.Command, args []string) error {
	value, err := keyring.Get(c.secret.keyring.get.service, c.secret.keyring.get.user)
	if err != nil {
		return err
	}
	return c.writeOutputString(value)
}

func (c *Config) runSecretKeyringSetCmdE(cmd *cobra.Command, args []string) error {
	value := c.secret.keyring.set.value
	if value == "" {
		var err error
		value, err = c.readPassword("Value: ")
		if err != nil {
			return err
		}
	}
	return keyring.Set(c.secret.keyring.set.service, c.secret.keyring.set.user, value)
}
