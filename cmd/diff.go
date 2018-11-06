package cmd

import (
	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Print the diff between the actual state and the target state",
	Run:   makeRun(runDiffCommand),
}

func init() {
	rootCommand.AddCommand(diffCommand)
}

func runDiffCommand(command *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	targetState, err := getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := chezmoi.NewLoggingActuator(chezmoi.NewNullActuator())
	return targetState.Ensure(fs, targetDir, getUmask(), actuator)
}
