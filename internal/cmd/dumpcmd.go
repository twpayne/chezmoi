package cmd

import (
	"cmp"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type dumpCmdConfig struct {
	filter     *chezmoi.EntryTypeFilter
	format     *choiceFlag
	init       bool
	parentDirs bool
	recursive  bool
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
	dumpCmd.Flags().VarP(c.dump.format, "format", "f", "Output format")
	must(dumpCmd.RegisterFlagCompletionFunc("format", c.dump.format.FlagCompletionFunc()))
	dumpCmd.Flags().VarP(c.dump.filter.Include, "include", "i", "Include entry types")
	dumpCmd.Flags().BoolVar(&c.dump.init, "init", c.dump.init, "Recreate config file from template")
	dumpCmd.Flags().BoolVarP(&c.dump.parentDirs, "parent-dirs", "P", c.dump.parentDirs, "Dump all parent directories")
	dumpCmd.Flags().BoolVarP(&c.dump.recursive, "recursive", "r", c.dump.recursive, "Recurse into subdirectories")

	return dumpCmd
}

func (c *Config) runDumpCmd(cmd *cobra.Command, args []string) error {
	dumpSystem := chezmoi.NewDumpSystem()
	if err := c.applyArgs(cmd.Context(), dumpSystem, chezmoi.EmptyAbsPath, args, applyArgsOptions{
		cmd:        cmd,
		filter:     c.dump.filter,
		init:       c.dump.init,
		parentDirs: c.dump.parentDirs,
		recursive:  c.dump.recursive,
		umask:      c.Umask,
	}); err != nil {
		return err
	}
	return c.marshal(cmp.Or(c.dump.format.String(), c.Format.String()), dumpSystem.Data())
}
