package cmd

// FIXME add --force flag
// FIXME add --recursive flag
// FIXME add --prompt flag

import (
	"fmt"
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
	actuator := c.getDefaultActuator(fs)
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		entry, err := targetState.Get(targetPath)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("%s: not under management", arg)
		}
		if err := actuator.RemoveAll(targetPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := actuator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
