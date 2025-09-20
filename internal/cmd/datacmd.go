package cmd

import (
	"cmp"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type dataCmdConfig struct {
	format *choiceFlag
}

func (c *Config) newDataCmd() *cobra.Command {
	dataCmd := &cobra.Command{
		Use:               "data",
		Short:             "Print the template data",
		Long:              mustLongHelp("data"),
		Example:           example("data"),
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE:              c.runDataCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}

	dataCmd.Flags().VarP(c.data.format, "format", "f", "Output format")
	must(dataCmd.RegisterFlagCompletionFunc("format", c.data.format.FlagCompletionFunc()))

	return dataCmd
}

func (c *Config) runDataCmd(cmd *cobra.Command, args []string) error {
	sourceState, err := c.newSourceState(cmd.Context(), cmd,
		chezmoi.WithTemplateDataOnly(true),
	)
	if err != nil {
		return err
	}
	return c.marshal(cmp.Or(c.data.format.String(), c.Format.String()), sourceState.TemplateData())
}
