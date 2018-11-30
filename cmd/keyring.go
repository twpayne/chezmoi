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

// A keyringCommandConfig is a configuration for the keyring command.
type keyringCommandConfig struct {
	service  string
	user     string
	password string
}

func init() {
	rootCommand.AddCommand(keyringCommand)

	persistentFlags := keyringCommand.PersistentFlags()

	persistentFlags.StringVar(&config.keyring.service, "service", "", "service")
	keyringCommand.MarkPersistentFlagRequired("service")

	persistentFlags.StringVar(&config.keyring.user, "user", "", "user")
	keyringCommand.MarkPersistentFlagRequired("user")

	config.addFunc("keyring", config.keyringFunc)
}

func (*Config) keyringFunc(service, user string) string {
	password, err := keyring.Get(service, user)
	if err != nil {
		return err.Error()
	}
	return password
}
