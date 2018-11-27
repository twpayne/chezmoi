package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
	"github.com/zalando/go-keyring"
)

var keyringGetCommand = &cobra.Command{
	Use:   "get",
	Args:  cobra.NoArgs,
	Short: "Get a password from keyring",
	RunE:  makeRunE(config.runKeyringGetCommand),
}

func init() {
	keyringCommand.AddCommand(keyringGetCommand)
}

func (c *Config) runKeyringGetCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	password, err := keyring.Get(c.Keyring.Service, c.Keyring.User)
	if err != nil {
		return err
	}
	fmt.Println(password)
	return nil
}
