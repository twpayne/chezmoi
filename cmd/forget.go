package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var forgetCommand = &cobra.Command{
	Use:   "forget",
	Args:  cobra.MinimumNArgs(1),
	Short: "Forget a file or directory",
	RunE:  makeRunE(config.runForgetCommand),
}

func init() {
	rootCommand.AddCommand(forgetCommand)
}

func (c *Config) runForgetCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(targetState, args)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	for _, entry := range entries {
		if err := actuator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil {
			return err
		}
	}
	return nil
}
