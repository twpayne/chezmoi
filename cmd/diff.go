package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Write the diff between the actual state and the target state to stdout",
	RunE:  makeRunE(config.runDiffCommand),
}

func init() {
	rootCommand.AddCommand(diffCommand)
}

func (c *Config) runDiffCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	actuator := chezmoi.NewLoggingActuator(os.Stdout, chezmoi.NewNullActuator())
	return c.applyArgs(fs, args, actuator)
}
