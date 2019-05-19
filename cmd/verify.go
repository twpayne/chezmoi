package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var verifyCmd = &cobra.Command{
	Use:     "verify [targets...]",
	Short:   "Exit with success if the destination state matches the target state, fail otherwise",
	Long:    mustGetLongHelp("verify"),
	Example: getExample("verify"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runVerifyCmd),
}

func init() {
	rootCmd.AddCommand(verifyCmd)
}

func (c *Config) runVerifyCmd(fs vfs.FS, args []string) error {
	mutator := chezmoi.NewAnyMutator(chezmoi.NullMutator{})
	if err := c.applyArgs(fs, args, mutator); err != nil {
		return err
	}
	if mutator.Mutated() {
		os.Exit(1)
	}
	return nil
}
