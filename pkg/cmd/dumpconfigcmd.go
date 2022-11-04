package cmd

import "github.com/spf13/cobra"

func (c *Config) newDumpConfigCmd() *cobra.Command {
	dumpConfigCmd := &cobra.Command{
		Use:     "dump-config",
		Short:   "Dump the configuration values",
		Long:    mustLongHelp("dump-config"),
		Example: example("dump-config"),
		Args:    cobra.NoArgs,
		RunE:    c.runDumpConfigCmd,
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
		},
	}

	flags := dumpConfigCmd.Flags()
	flags.VarP(&c.Format, "format", "f", "Output format")

	return dumpConfigCmd
}

func (c *Config) runDumpConfigCmd(cmd *cobra.Command, args []string) error {
	return c.marshal(c.Format, c)
}
