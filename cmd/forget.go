package cmd

import (
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var forgetCommand = &cobra.Command{
	Use:   "forget",
	Args:  cobra.MinimumNArgs(1),
	Short: "Forget a file or directory",
	RunE:  makeRunE(runForgetCommandE),
}

func init() {
	rootCommand.AddCommand(forgetCommand)
}

func runForgetCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceNames, err := config.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	actuator := config.getDefaultActuator(fs)
	for _, sourceName := range sourceNames {
		if err := actuator.RemoveAll(filepath.Join(config.SourceDir, sourceName)); err != nil {
			return err
		}
	}
	return nil
}
