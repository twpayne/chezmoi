package cmd

import (
	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
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
	targetState, err := chezmoi.ReadTargetDirState(fs, sourceDir, nil)
	if err != nil {
		return err
	}
	return targetState.Ensure(fs, targetDir)
}
