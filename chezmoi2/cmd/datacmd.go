package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

func (c *Config) newDataCmd() *cobra.Command {
	dataCmd := &cobra.Command{
		Use:     "data",
		Short:   "Print the template data",
		Long:    mustLongHelp("data"),
		Example: example("data"),
		Args:    cobra.NoArgs,
		RunE:    c.makeRunEWithSourceState(c.runDataCmd),
	}

	return dataCmd
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	return c.marshal(sourceState.TemplateData())
}
