package cmd

import (
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

func (c *Config) newRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Deprecated: "use forget or destroy instead",
		Use:        "remove",
		Aliases:    []string{"rm"},
		RunE:       c.runRemoveCmd,
		Long:       mustLongHelp("remove"),
		Hidden:     true,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
		),
	}

	return removeCmd
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string) error {
	return chezmoi.ExitCodeError(1)
}
