package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type verifyCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	init      bool
	recursive bool
}

func (c *Config) newVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:               "verify [target]...",
		Short:             "Exit with success if the destination state matches the target state, fail otherwise",
		Long:              mustLongHelp("verify"),
		Example:           example("verify"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runVerifyCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadMockWrite,
		},
	}

	flags := verifyCmd.Flags()
	flags.VarP(c.verify.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.verify.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.verify.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.verify.recursive, "recursive", "r", c.verify.recursive, "Recurse into subdirectories")

	return verifyCmd
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	errorOnWriteSystem := chezmoi.NewErrorOnWriteSystem(c.destSystem, chezmoi.ExitCodeError(1))
	return c.applyArgs(cmd.Context(), errorOnWriteSystem, c.DestDirAbsPath, args, applyArgsOptions{
		concurrency: c.Concurrency,
		include:     c.verify.include.Sub(c.verify.exclude),
		init:        c.verify.init,
		recursive:   c.verify.recursive,
		umask:       c.Umask,
	})
}
