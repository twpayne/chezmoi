package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type applyCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	init      bool
	recursive bool
}

func (c *Config) newApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:               "apply [target]...",
		Short:             "Update the destination directory to match the target state",
		Long:              mustLongHelp("apply"),
		Example:           example("apply"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runApplyCmd,
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
		),
	}

	flags := applyCmd.Flags()
	flags.VarP(c.apply.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.apply.filter.Include, "include", "i", "Include entry types")
	flags.BoolVar(&c.apply.init, "init", c.apply.init, "Recreate config file from template")
	flags.BoolVarP(&c.apply.recursive, "recursive", "r", c.apply.recursive, "Recurse into subdirectories")

	registerExcludeIncludeFlagCompletionFuncs(applyCmd)

	return applyCmd
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		filter:       c.apply.filter,
		init:         c.apply.init,
		recursive:    c.apply.recursive,
		umask:        c.Umask,
		preApplyFunc: c.defaultPreApplyFunc,
	})
}
