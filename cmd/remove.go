package cmd

import (
	"os"
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var removeCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a file",
	Run:     makeRun(runRemoveCommand),
}

func init() {
	rootCommand.AddCommand(removeCommand)
}

func runRemoveCommand(command *cobra.Command, args []string) error {
	// FIXME support directories
	fs := afero.NewOsFs()
	targetState, err := getTargetState(fs)
	if err != nil {
		return err
	}
	sourceFileNames, err := getSourceFileNames(targetState, args)
	if err != nil {
		return err
	}
	actuator := getDefaultActuator(fs)
	for i, targetFileName := range args {
		if err := actuator.RemoveAll(filepath.Join(targetDir, targetFileName)); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := actuator.RemoveAll(filepath.Join(sourceDir, sourceFileNames[i])); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
