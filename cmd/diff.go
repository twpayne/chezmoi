package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Write the diff between the target state and the destination state to stdout",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runDiffCmd),
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func (c *Config) runDiffCmd(fs vfs.FS, args []string) error {
	mutator := chezmoi.NewLoggingMutator(c.Stdout(), chezmoi.NullMutator)
	return c.applyArgs(fs, args, mutator)
}
