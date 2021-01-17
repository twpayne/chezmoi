package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

type secretKeyringCmdConfig struct {
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

	persistentFlags := keyringCmd.PersistentFlags()
	persistentFlags.StringVar(&c.secretKeyring.service, "service", "", "service")
	persistentFlags.StringVar(&c.secretKeyring.user, "user", "", "user")
	markPersistentFlagsRequired(keyringCmd, "service", "user")

	keyringGetCmd := &cobra.Command{
		Use:   "get",
		Args:  cobra.NoArgs,
		Short: "Get a value from keyring",
		RunE:  c.runKeyringGetCmdE,
	}
	keyringCmd.AddCommand(keyringGetCmd)

	keyringSetCmd := &cobra.Command{
		Use:   "set",
		Args:  cobra.NoArgs,
		Short: "Set a value in keyring",
		RunE:  c.runKeyringSetCmdE,
	}
	keyringCmd.AddCommand(keyringSetCmd)

	return keyringCmd
}

func (c *Config) runKeyringGetCmdE(cmd *cobra.Command, args []string) error {
	value, err := keyring.Get(c.secretKeyring.service, c.secretKeyring.user)
	if err != nil {
		return err
	}
	return c.writeOutputString(value)
}

func (c *Config) runKeyringSetCmdE(cmd *cobra.Command, args []string) error {
	valueStr := c.secretKeyring.value
	if valueStr == "" {
		fmt.Print("Value: ")
		value, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		valueStr = string(value)
	}
	return keyring.Set(c.secretKeyring.service, c.secretKeyring.user, valueStr)
}
