package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	keyring "github.com/zalando/go-keyring"
)

var keyringCmd = &cobra.Command{
	Use:   "keyring",
	Args:  cobra.NoArgs,
	Short: "Interact with keyring",
}

type keyringCmdConfig struct {
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
	secretCmd.AddCommand(keyringCmd)

	persistentFlags := keyringCmd.PersistentFlags()

	persistentFlags.StringVar(&config.keyring.service, "service", "", "service")
	panicOnError(keyringCmd.MarkPersistentFlagRequired("service"))

	persistentFlags.StringVar(&config.keyring.user, "user", "", "user")
	panicOnError(keyringCmd.MarkPersistentFlagRequired("user"))

	config.addTemplateFunc("keyring", config.keyringFunc)
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
		panic(fmt.Errorf("%q %q: %w", service, user, err))
	}
	keyringCache[key] = password
	return password
}
