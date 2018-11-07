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

func runRemoveCommand(fs afero.Fs, command *cobra.Command, args []string) error {
	// FIXME support directories
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceFileNames, err := config.getSourceFileNames(targetState, args)
	if err != nil {
		return err
	}
	actuator := config.getDefaultActuator(fs)
	for i, targetFileName := range args {
		if err := actuator.RemoveAll(filepath.Join(config.TargetDir, targetFileName)); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := actuator.RemoveAll(filepath.Join(config.SourceDir, sourceFileNames[i])); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
