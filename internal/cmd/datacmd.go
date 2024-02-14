package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newDataCmd() *cobra.Command {
	dataCmd := &cobra.Command{
		Use:         "data",
		Short:       "Print the template data",
		Long:        mustLongHelp("data"),
		Example:     example("data"),
		Args:        cobra.NoArgs,
		RunE:        c.runDataCmd,
		Annotations: newAnnotations(),
	}

	dataCmd.Flags().VarP(&c.Format, "format", "f", "Output format")

	return dataCmd
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string) error {
	sourceState, err := c.newSourceState(cmd.Context(), cmd,
		chezmoi.WithTemplateDataOnly(true),
	)
	if err != nil {
		return err
	}
	return c.marshal(c.Format, sourceState.TemplateData())
}
