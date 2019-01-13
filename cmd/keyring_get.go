package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
	keyring "github.com/zalando/go-keyring"
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

func (c *Config) runKeyringGetCommand(fs vfs.FS, args []string) error {
	password, err := keyring.Get(c.keyring.service, c.keyring.user)
	if err != nil {
		return err
	}
	fmt.Println(password)
	return nil
}
