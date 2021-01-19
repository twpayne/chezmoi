package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type applyCmdConfig struct {
	ignoreEncrypted bool
	include         *chezmoi.IncludeSet
	recursive       bool
	sourcePath      bool
}

func (c *Config) newApplyCmd() *cobra.Command {
	applyCmd := &cobra.Command{
		Use:     "apply [target]...",
		Short:   "Update the destination directory to match the target state",
		Long:    mustLongHelp("apply"),
		Example: example("apply"),
		RunE:    c.runApplyCmd,
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadWrite,
		},
	}

	flags := applyCmd.Flags()
	flags.BoolVar(&c.apply.ignoreEncrypted, "ignore-encrypted", c.apply.ignoreEncrypted, "ignore encrypted files")
	flags.VarP(c.apply.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.apply.recursive, "recursive", "r", c.apply.recursive, "recursive")
	flags.BoolVar(&c.apply.sourcePath, "source-path", c.apply.sourcePath, "specify targets by source path")

	return applyCmd
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	return c.applyArgs(c.destSystem, c.destDirAbsPath, args, applyArgsOptions{
		ignoreEncrypted: c.apply.ignoreEncrypted,
		include:         c.apply.include,
		recursive:       c.apply.recursive,
		sourcePath:      c.apply.sourcePath,
		umask:           c.Umask.FileMode(),
		preApplyFunc:    c.defaultPreApplyFunc,
	})
}
