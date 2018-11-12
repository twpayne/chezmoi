package cmd

import (
	"os"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var verifyCommand = &cobra.Command{
	Use:   "verify",
	Short: "Exit with success if the actual state matches the target state, fail otherwise",
	RunE:  makeRunE(config.runVerifyCommandE),
}

func init() {
	rootCommand.AddCommand(verifyCommand)
}

func (c *Config) runVerifyCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	anyActuator := chezmoi.NewAnyActuator(chezmoi.NewNullActuator())
	if err := targetState.Apply(fs, anyActuator); err != nil {
		return err
	}
	if anyActuator.Actuated() {
		os.Exit(1)
	}
	return nil
}
