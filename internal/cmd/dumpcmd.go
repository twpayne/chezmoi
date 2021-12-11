package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type dumpCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	format    writeDataFormat
	include   *chezmoi.EntryTypeSet
	init      bool
	recursive bool
}

func (c *Config) newDumpCmd() *cobra.Command {
	dumpCmd := &cobra.Command{
		Use:               "dump [target]...",
		Short:             "Generate a dump of the target state",
		Long:              mustLongHelp("dump"),
		Example:           example("dump"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runDumpCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeEmpty,
		},
	}

	flags := dumpCmd.Flags()
	flags.VarP(c.dump.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(&c.dump.format, "format", "f", "Set output format")
	flags.VarP(c.dump.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.dump.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.dump.recursive, "recursive", "r", c.dump.recursive, "Recurse into subdirectories")

	return dumpCmd
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(cmd.Context(), dumpSystem, chezmoi.EmptyAbsPath, args, applyArgsOptions{
		include:   c.dump.include.Sub(c.dump.exclude),
		init:      c.dump.init,
		recursive: c.dump.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	return c.marshal(c.dump.format, dumpSystem.Data())
}
