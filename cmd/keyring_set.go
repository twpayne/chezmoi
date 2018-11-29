package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var keyringSetCommand = &cobra.Command{
	Use:   "set",
	Args:  cobra.NoArgs,
	Short: "Set a password in keyring",
	RunE:  makeRunE(config.runKeyringSetCommand),
}

func init() {
	keyringCommand.AddCommand(keyringSetCommand)

	persistentFlags := keyringSetCommand.PersistentFlags()
	persistentFlags.StringVar(&config.keyring.password, "password", "", "password")
}

func (c *Config) runKeyringSetCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	passwordString := c.keyring.password
	if passwordString == "" {
		fmt.Print("Password: ")
		password, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		passwordString = string(password)
	}
	return keyring.Set(c.keyring.service, c.keyring.user, passwordString)
}
