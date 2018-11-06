package cmd

import (
	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var ensureCommand = &cobra.Command{
	Use:   "ensure",
	Short: "Ensure that the actual state matches the target state",
	Run:   makeRun(runEnsureCommand),
}

func init() {
	rootCommand.AddCommand(ensureCommand)
}

func runEnsureCommand(command *cobra.Command, args []string) error {
	fs := afero.NewOsFs()
	targetState, err := getTargetState(fs)
	if err != nil {
		return err
	}
	return targetState.Ensure(fs, targetDir, getUmask(), getDefaultActuator(fs))
}
