package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type dumpCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
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
		Annotations: newAnnotations(
			persistentStateModeReadMockWrite,
			requiresSourceDirectory,
		),
	}

	dumpCmd.Flags().VarP(c.dump.filter.Exclude, "exclude", "x", "Exclude entry types")
	dumpCmd.Flags().VarP(&c.Format, "format", "f", "Output format")
	dumpCmd.Flags().VarP(c.dump.filter.Include, "include", "i", "Include entry types")
	dumpCmd.Flags().BoolVar(&c.dump.init, "init", c.dump.init, "Recreate config file from template")
	dumpCmd.Flags().BoolVarP(&c.dump.recursive, "recursive", "r", c.dump.recursive, "Recurse into subdirectories")
	if err := dumpCmd.RegisterFlagCompletionFunc("format", writeDataFormatFlagCompletionFunc); err != nil {
		panic(err)
	}

	registerExcludeIncludeFlagCompletionFuncs(dumpCmd)

	return dumpCmd
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(cmd.Context(), dumpSystem, chezmoi.EmptyAbsPath, args, applyArgsOptions{
		cmd:       cmd,
		filter:    c.dump.filter,
		init:      c.dump.init,
		recursive: c.dump.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	return c.marshal(c.Format, dumpSystem.Data())
}
