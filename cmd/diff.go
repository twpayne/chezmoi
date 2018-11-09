package cmd

import (
	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Print the diff between the actual state and the target state",
	RunE:  makeRunE(runDiffCommandE),
}

func init() {
	rootCommand.AddCommand(diffCommand)
}

func runDiffCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := chezmoi.NewLoggingActuator(chezmoi.NewNullActuator())
	return targetState.Ensure(fs, actuator)
}
