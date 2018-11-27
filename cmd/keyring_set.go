package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
	"github.com/zalando/go-keyring"
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
	persistentFlags.StringVar(&config.Keyring.Password, "password", "", "password")
	keyringSetCommand.MarkPersistentFlagRequired("password")
}

func (c *Config) runKeyringSetCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	return keyring.Set(c.Keyring.Service, c.Keyring.User, c.Keyring.Password)
}
