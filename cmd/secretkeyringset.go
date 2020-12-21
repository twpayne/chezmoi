package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
	"golang.org/x/term"
)

var keyringSetCmd = &cobra.Command{
	Use:     "set",
	Args:    cobra.NoArgs,
	Short:   "Set a value in keyring",
	PreRunE: config.ensureNoError,
	RunE:    config.runKeyringSetCmd,
}

func init() {
	keyringCmd.AddCommand(keyringSetCmd)

	persistentFlags := keyringSetCmd.PersistentFlags()
	persistentFlags.StringVar(&config.keyring.value, "value", "", "value")
}

func (c *Config) runKeyringSetCmd(cmd *cobra.Command, args []string) error {
	valueStr := c.keyring.value
	if valueStr == "" {
		fmt.Print("Value: ")
		value, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		valueStr = string(value)
	}
	return keyring.Set(c.keyring.service, c.keyring.user, valueStr)
}
