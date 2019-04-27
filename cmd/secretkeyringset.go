package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
)

var keyringSetCmd = &cobra.Command{
	Use:     "set",
	Args:    cobra.NoArgs,
	Short:   "Set a password in keyring",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runKeyringSetCmd),
}

func init() {
	keyringCmd.AddCommand(keyringSetCmd)

	persistentFlags := keyringSetCmd.PersistentFlags()
	persistentFlags.StringVar(&config.keyring.password, "password", "", "password")
}

func (c *Config) runKeyringSetCmd(fs vfs.FS, args []string) error {
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
