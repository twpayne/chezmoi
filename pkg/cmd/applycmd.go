package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type applyCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	init      bool
	include   *chezmoi.EntryTypeSet
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
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadWrite,
			requiresSourceDirectory:      "true",
		},
	}

	flags := applyCmd.Flags()
	flags.VarP(c.apply.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.apply.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.apply.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.apply.recursive, "recursive", "r", c.apply.recursive, "Recurse into subdirectories")

	return applyCmd
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:      c.apply.include.Sub(c.apply.exclude),
		init:         c.apply.init,
		recursive:    c.apply.recursive,
		umask:        c.Umask,
		preApplyFunc: c.defaultPreApplyFunc,
	})
}
