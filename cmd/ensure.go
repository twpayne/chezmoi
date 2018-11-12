package cmd

import (
	"github.com/absfs/afero"
	"github.com/spf13/cobra"
)

var ensureCommand = &cobra.Command{
	Use:   "ensure",
	Short: "Ensure that the actual state matches the target state",
	RunE:  makeRunE(config.runEnsureCommandE),
}

func init() {
	rootCommand.AddCommand(ensureCommand)
}

func (c *Config) runEnsureCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := c.getDefaultActuator(fs)
	return targetState.Ensure(fs, actuator)
}
