package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type verifyCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	init      bool
	recursive bool
}

func (c *Config) newVerifyCmd() *cobra.Command {
	verifyCmd := &cobra.Command{
		Use:     "verify [target]...",
		Short:   "Exit with success if the destination state matches the target state, fail otherwise",
		Long:    mustLongHelp("verify"),
		Example: example("verify"),
		RunE:    c.runVerifyCmd,
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
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	if err := c.applyArgs(cmd.Context(), dryRunSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:   c.verify.include.Sub(c.verify.exclude),
		init:      c.verify.init,
		recursive: c.verify.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	if dryRunSystem.Modified() {
		return ExitCodeError(1)
	}
	return nil
}
