package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var removeCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a file or directory",
	RunE:    makeRunE(config.runRemoveCommand),
}

func init() {
	rootCommand.AddCommand(removeCommand)
}

func (c *Config) runRemoveCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceNames, err := c.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	for i, targetFileName := range args {
		if err := actuator.RemoveAll(filepath.Join(c.TargetDir, targetFileName)); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := actuator.RemoveAll(filepath.Join(c.SourceDir, sourceNames[i])); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
