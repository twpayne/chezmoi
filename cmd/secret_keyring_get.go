package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
	keyring "github.com/zalando/go-keyring"
)

var keyringGetCmd = &cobra.Command{
	Use:     "get",
	Args:    cobra.NoArgs,
	Short:   "Get a password from keyring",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runKeyringGetCmd),
}

func init() {
	keyringCmd.AddCommand(keyringGetCmd)
}

func (c *Config) runKeyringGetCmd(fs vfs.FS, args []string) error {
	password, err := keyring.Get(c.keyring.service, c.keyring.user)
	if err != nil {
		return err
	}
	fmt.Println(password)
	return nil
}
