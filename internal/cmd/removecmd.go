package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Deprecated: "use forget or destroy instead",
		Use:        "remove",
		Aliases:    []string{"rm"},
		RunE:       c.runRemoveCmd,
		Long:       mustLongHelp("remove"),
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
		),
	}

	return removeCmd
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string) error {
	return chezmoi.ExitCodeError(1)
}
