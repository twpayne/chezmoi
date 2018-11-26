package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var forgetCommand = &cobra.Command{
	Use:   "forget",
	Args:  cobra.MinimumNArgs(1),
	Short: "Forget a file or directory",
	RunE:  makeRunE(config.runForgetCommandE),
}

func init() {
	rootCommand.AddCommand(forgetCommand)
}

func (c *Config) runForgetCommandE(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceNames, err := c.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	for _, sourceName := range sourceNames {
		if err := actuator.RemoveAll(filepath.Join(c.SourceDir, sourceName)); err != nil {
			return err
		}
	}
	return nil
}
