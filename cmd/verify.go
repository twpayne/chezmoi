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
	Run:   makeRun(runVerifyCommand),
}

func init() {
	rootCommand.AddCommand(verifyCommand)
}

func runVerifyCommand(command *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	targetState, err := getTargetState(fs)
	if err != nil {
		return err
	}
	anyActuator := chezmoi.NewAnyActuator(chezmoi.NewNullActuator())
	if err := targetState.Ensure(fs, targetDir, getUmask(), anyActuator); err != nil {
		return err
	}
	if anyActuator.Actuated() {
		os.Exit(1)
	}
	return nil
}
