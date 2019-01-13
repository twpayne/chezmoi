package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var diffCommand = &cobra.Command{
	Use:   "diff",
	Short: "Write the diff between the target state and the destination state to stdout",
	RunE:  makeRunE(config.runDiffCommand),
}

func init() {
	rootCommand.AddCommand(diffCommand)
}

func (c *Config) runDiffCommand(fs vfs.FS, args []string) error {
	mutator := chezmoi.NewLoggingMutator(os.Stdout, chezmoi.NullMutator)
	return c.applyArgs(fs, args, mutator)
}
