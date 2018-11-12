package cmd

import (
	"os"
	"path/filepath"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var addCommand = &cobra.Command{
	Use:   "add",
	Args:  cobra.MinimumNArgs(1),
	Short: "Add an existing file or directory",
	RunE:  makeRunE(runAddCommandE),
}

func init() {
	rootCommand.AddCommand(addCommand)

	persistentFlags := addCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.Add.Recursive, "recursive", "r", false, "recurse in to subdirectories")
	persistentFlags.BoolVarP(&config.Add.Template, "template", "T", false, "add files as templates")
}

func runAddCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := config.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := config.getDefaultActuator(fs)
	for _, arg := range args {
		if config.Add.Recursive {
			if err := afero.Walk(fs, filepath.Join(targetState.TargetDir, arg), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				targetName, err := filepath.Rel(targetState.TargetDir, path)
				if err != nil {
					return err
				}
				return targetState.Add(fs, targetName, info, config.Add.Template, actuator)
			}); err != nil {
				return err
			}
		} else {
			targetName, err := filepath.Rel(targetState.TargetDir, arg)
			if err != nil {
				return err
			}
			if err := targetState.Add(fs, targetName, nil, config.Add.Template, actuator); err != nil {
				return err
			}
		}
	}
	return nil
}
