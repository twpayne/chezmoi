package cmd

import (
	"os"

	"github.com/absfs/afero"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Args:  cobra.NoArgs, // FIXME should accept list of targets
	Short: "Print the diff between the actual state and the target state",
	RunE:  makeRunE(config.runDiffCommandE),
}

func init() {
	rootCommand.AddCommand(diffCommand)
}

func (c *Config) runDiffCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	actuator := chezmoi.NewLoggingActuator(os.Stdout, chezmoi.NewNullActuator())
	return targetState.Apply(fs, actuator)
}
