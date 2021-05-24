package chezmoicmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type verifyCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
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
	flags.VarP(c.verify.exclude, "exclude", "x", "exclude entry types")
	flags.VarP(c.verify.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.verify.recursive, "recursive", "r", c.verify.recursive, "recursive")

	return verifyCmd
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	if err := c.applyArgs(dryRunSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:   c.verify.include.Sub(c.verify.exclude),
		recursive: c.verify.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	if dryRunSystem.Modified() {
		return ErrExitCode(1)
	}
	return nil
}
