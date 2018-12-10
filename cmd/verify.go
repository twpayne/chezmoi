package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var verifyCommand = &cobra.Command{
	Use:   "verify",
	Short: "Exit with success if the actual state matches the target state, fail otherwise",
	RunE:  makeRunE(config.runVerifyCommand),
}

func init() {
	rootCommand.AddCommand(verifyCommand)
}

func (c *Config) runVerifyCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	mutator := chezmoi.NewAnyMutator(chezmoi.NullMutator)
	if err := c.applyArgs(fs, args, mutator); err != nil {
		return err
	}
	if mutator.Mutated() {
		os.Exit(1)
	}
	return nil
}
