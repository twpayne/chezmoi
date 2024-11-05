package cmd

import (
	"cmp"

	"github.com/spf13/cobra"
)

type dumpConfigCmdConfig struct {
	format *choiceFlag
}

func (c *Config) newDumpConfigCmd() *cobra.Command {
	dumpConfigCmd := &cobra.Command{
		Use:     "dump-config",
		Short:   "Dump the configuration values",
		Long:    mustLongHelp("dump-config"),
		Example: example("dump-config"),
		Args:    cobra.NoArgs,
		RunE:    c.runDumpConfigCmd,
		Annotations: newAnnotations(
			persistentStateModeReadOnly,
		),
	}

	dumpConfigCmd.Flags().VarP(c.dumpConfig.format, "format", "f", "Output format")

	must(dumpConfigCmd.RegisterFlagCompletionFunc("format", c.dumpConfig.format.FlagCompletionFunc()))

	return dumpConfigCmd
}

func (c *Config) runDumpConfigCmd(cmd *cobra.Command, args []string) error {
	return c.marshal(cmp.Or(c.dumpConfig.format.String(), c.Format), c)
}
