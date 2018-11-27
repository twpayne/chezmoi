package cmd

import (
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

var keyringCommand = &cobra.Command{
	Use:   "keyring",
	Args:  cobra.NoArgs,
	Short: "Interact with keyring",
}

func init() {
	rootCommand.AddCommand(keyringCommand)

	persistentFlags := keyringCommand.PersistentFlags()

	persistentFlags.StringVar(&config.Keyring.Service, "service", "", "service")
	keyringCommand.MarkPersistentFlagRequired("service")

	persistentFlags.StringVar(&config.Keyring.User, "user", "", "user")
	keyringCommand.MarkPersistentFlagRequired("user")

	config.addFunc("keyring", func(service, user string) string {
		password, err := keyring.Get(service, user)
		if err != nil {
			return err.Error()
		}
		return password
	})
}
