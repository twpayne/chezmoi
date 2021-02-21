package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type dumpCmdConfig struct {
	format    string
	include   *chezmoi.EntryTypeSet
	recursive bool
}

func (c *Config) newDumpCmd() *cobra.Command {
	dumpCmd := &cobra.Command{
		Use:     "dump [target]...",
		Short:   "Generate a dump of the target state",
		Long:    mustLongHelp("dump"),
		Example: example("dump"),
		RunE:    c.runDumpCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeEmpty,
		},
	}

	flags := dumpCmd.Flags()
	flags.StringVarP(&c.dump.format, "format", "f", c.dump.format, "format (json or yaml)")
	flags.VarP(c.dump.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.dump.recursive, "recursive", "r", c.dump.recursive, "recursive")

	return dumpCmd
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(dumpSystem, "", args, applyArgsOptions{
		include:   c.dump.include,
		recursive: c.dump.recursive,
	}); err != nil {
		return err
	}
	return c.marshal(c.dump.format, dumpSystem.Data())
}
