package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var applyCommand = &cobra.Command{
	Use:   "apply",
	Args:  cobra.NoArgs,
	Short: "Update the actual state to match the target state",
	RunE:  makeRunE(config.runApplyCommandE),
}

func init() {
	rootCommand.AddCommand(applyCommand)
}

func (c *Config) runApplyCommandE(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	return targetState.Apply(fs, actuator)
}
