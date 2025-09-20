package cmd

import (
	"github.com/twpayne/go-pinentry/v4"

	"chezmoi.io/chezmoi/internal/chezmoierrors"
)

type pinEntryConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
	Options []string `json:"options" mapstructure:"options" yaml:"options"`
}

var pinEntryDefaultOptions = []string{
	pinentry.OptionAllowExternalPasswordCache,
}

func (c *Config) readPINEntry(prompt string) (string, error) {
	var client *pinentry.Client
	client, err := pinentry.NewClient(
		pinentry.WithArgs(c.PINEntry.Args),
		pinentry.WithBinaryName(c.PINEntry.Command),
		pinentry.WithGPGTTY(),
		pinentry.WithOptions(c.PINEntry.Options),
		pinentry.WithPrompt(prompt),
		pinentry.WithTitle("chezmoi"),
	)
	if err != nil {
		return "", err
	}
	defer chezmoierrors.CombineFunc(&err, client.Close)

	result, err := client.GetPIN()
	if err != nil {
		return "", err
	}

	return result.PIN, nil
}
