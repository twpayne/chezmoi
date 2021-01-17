package cmd

import (
	"github.com/spf13/cobra"
)

type purgeCmdConfig struct {
	binary bool
}

func (c *Config) newPurgeCmd() *cobra.Command {
	purgeCmd := &cobra.Command{
		Use:     "purge",
		Short:   "Purge chezmoi's configuration and data",
		Long:    mustLongHelp("purge"),
		Example: example("purge"),
		Args:    cobra.NoArgs,
		RunE:    c.runPurgeCmd,
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
		},
	}

	flags := purgeCmd.Flags()
	flags.BoolVarP(&c.purge.binary, "binary", "P", c.purge.binary, "purge chezmoi executable")

	return purgeCmd
}

func (c *Config) runPurgeCmd(cmd *cobra.Command, args []string) error {
	return c.doPurge(&purgeOptions{
		binary: c.purge.binary,
	})
}
