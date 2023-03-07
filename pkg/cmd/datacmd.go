package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newDataCmd() *cobra.Command {
	dataCmd := &cobra.Command{
		Use:     "data",
		Short:   "Print the template data",
		Long:    mustLongHelp("data"),
		Example: example("data"),
		Args:    cobra.NoArgs,
		RunE:    c.runDataCmd,
	}

	persistentFlags := dataCmd.PersistentFlags()
	persistentFlags.VarP(&c.Format, "format", "f", "Output format")
	if err := dataCmd.RegisterFlagCompletionFunc("format", writeDataFormatFlagCompletionFunc); err != nil {
		panic(err)
	}

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
