package cmd

import (
	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type applyCmdConfig struct {
	filter     *chezmoi.EntryTypeFilter
	init       bool
	parentDirs bool
	recursive  bool
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

	applyCmd.Flags().VarP(c.apply.filter.Exclude, "exclude", "x", "Exclude entry types")
	applyCmd.Flags().VarP(c.apply.filter.Include, "include", "i", "Include entry types")
	applyCmd.Flags().BoolVar(&c.apply.init, "init", c.apply.init, "Recreate config file from template")
	applyCmd.Flags().BoolVarP(&c.apply.parentDirs, "parent-dirs", "P", c.apply.parentDirs, "Apply all parent directories")
	applyCmd.Flags().BoolVarP(&c.apply.recursive, "recursive", "r", c.apply.recursive, "Recurse into subdirectories")

	return applyCmd
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:          cmd,
		filter:       c.apply.filter,
		init:         c.apply.init,
		parentDirs:   c.apply.parentDirs,
		recursive:    c.apply.recursive,
		umask:        c.Umask,
		preApplyFunc: c.defaultPreApplyFunc,
	})
}
