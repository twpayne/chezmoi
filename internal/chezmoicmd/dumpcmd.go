package chezmoicmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type dumpCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	format    dataFormat
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
	flags.VarP(c.dump.exclude, "exclude", "x", "exclude entry types")
	flags.VarP(&c.dump.format, "format", "f", "format")
	flags.VarP(c.dump.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.dump.recursive, "recursive", "r", c.dump.recursive, "recursive")

	return dumpCmd
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(dumpSystem, "", args, applyArgsOptions{
		include:   c.dump.include.Sub(c.dump.exclude),
		recursive: c.dump.recursive,
	}); err != nil {
		return err
	}
	return c.marshal(c.dump.format, dumpSystem.Data())
}
