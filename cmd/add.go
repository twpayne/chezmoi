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
	RunE:  makeRunE(config.runAddCommandE),
}

func init() {
	rootCommand.AddCommand(addCommand)

	persistentFlags := addCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.Add.Empty, "empty", "e", false, "add empty files")
	persistentFlags.BoolVarP(&config.Add.Recursive, "recursive", "r", false, "recurse in to subdirectories")
	persistentFlags.BoolVarP(&config.Add.Template, "template", "T", false, "add files as templates")
}

func (c *Config) runAddCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	for _, arg := range args {
		if c.Add.Recursive {
			if err := afero.Walk(fs, filepath.Join(targetState.TargetDir, arg), func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				return targetState.Add(fs, path, info, c.Add.Empty, c.Add.Template, actuator)
			}); err != nil {
				return err
			}
		} else {
			path, err := filepath.Abs(arg)
			if err != nil {
				return err
			}
			if err := targetState.Add(fs, path, nil, c.Add.Empty, c.Add.Template, actuator); err != nil {
				return err
			}
		}
	}
	return nil
}
