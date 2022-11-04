package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type verifyCmdConfig struct {
	Exclude   *chezmoi.EntryTypeSet `json:"exclude" mapstructure:"exclude" yaml:"exclude"`
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
			persistentStateMode:     persistentStateModeReadMockWrite,
			requiresSourceDirectory: "true",
		},
	}

	flags := verifyCmd.Flags()
	flags.VarP(c.Verify.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Verify.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Verify.init, "init", c.Verify.init, "Recreate config file from template")
	flags.BoolVarP(&c.Verify.recursive, "recursive", "r", c.Verify.recursive, "Recurse into subdirectories")

	registerExcludeIncludeFlagCompletionFuncs(verifyCmd)

	return verifyCmd
}

func (c *Config) runVerifyCmd(cmd *cobra.Command, args []string) error {
	errorOnWriteSystem := chezmoi.NewErrorOnWriteSystem(c.destSystem, chezmoi.ExitCodeError(1))
	return c.applyArgs(cmd.Context(), errorOnWriteSystem, c.DestDirAbsPath, args, applyArgsOptions{
		filter:    chezmoi.NewEntryTypeFilter(c.Verify.include.Bits(), c.Verify.Exclude.Bits()),
		init:      c.Verify.init,
		recursive: c.Verify.recursive,
		umask:     c.Umask,
	})
}
