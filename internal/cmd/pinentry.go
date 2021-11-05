package cmd

import (
	"github.com/twpayne/go-pinentry"
	"go.uber.org/multierr"
)

type pinEntryConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
	Options []string `mapstructure:"options"`
}

var pinEntryDefaultOptions = []string{
	pinentry.OptionAllowExternalPasswordCache,
}

func (c *Config) readPINEntry(prompt string) (pin string, err error) {
	var client *pinentry.Client
	client, err = pinentry.NewClient(
		pinentry.WithArgs(c.PINEntry.Args),
		pinentry.WithBinaryName(c.PINEntry.Command),
		pinentry.WithGPGTTY(),
		pinentry.WithOptions(c.PINEntry.Options),
		pinentry.WithPrompt(prompt),
		pinentry.WithTitle("chezmoi"),
	)
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, client.Close())
	}()

	pin, _, err = client.GetPIN()
	return
}
