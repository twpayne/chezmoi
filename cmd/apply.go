package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var applyCommand = &cobra.Command{
	Use:   "apply",
	Short: "Update the actual state to match the target state",
	RunE:  makeRunE(config.runApplyCommand),
}

func init() {
	rootCommand.AddCommand(applyCommand)
}

func (c *Config) runApplyCommand(fs vfs.FS, args []string) error {
	mutator := c.getDefaultMutator(fs)
	return c.applyArgs(fs, args, mutator)
}
