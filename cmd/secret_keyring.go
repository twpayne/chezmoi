package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	keyring "github.com/zalando/go-keyring"
)

var keyringCommand = &cobra.Command{
	Use:   "keyring",
	Args:  cobra.NoArgs,
	Short: "Interact with keyring",
}

type keyringCommandConfig struct {
	service  string
	user     string
	password string
}

type keyringKey struct {
	service string
	user    string
}

var keyringCache = make(map[keyringKey]string)

func init() {
	secretCommand.AddCommand(keyringCommand)

	persistentFlags := keyringCommand.PersistentFlags()

	persistentFlags.StringVar(&config.keyring.service, "service", "", "service")
	keyringCommand.MarkPersistentFlagRequired("service")

	persistentFlags.StringVar(&config.keyring.user, "user", "", "user")
	keyringCommand.MarkPersistentFlagRequired("user")

	config.addFunc("keyring", config.keyringFunc)
}

func (*Config) keyringFunc(service, user string) string {
	key := keyringKey{
		service: service,
		user:    user,
	}
	if password, ok := keyringCache[key]; ok {
		return password
	}
	password, err := keyring.Get(service, user)
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("keyring %q %q: %v", service, user, err))
	}
	keyringCache[key] = password
	return password
}
