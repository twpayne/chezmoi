package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	bolt "go.etcd.io/bbolt"
)

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Print the diff between the target state and the destination state",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDiffCmd,
}

func init() {
	rootCmd.AddCommand(diffCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(diffCmd, 1)
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	c.DryRun = true
	c.mutator = chezmoi.NullMutator{}
	if c.Debug {
		c.mutator = chezmoi.NewDebugMutator(c.mutator)
	}
	c.mutator = chezmoi.NewVerboseMutator(c.Stdout, c.mutator, c.colored, c.maxDiffDataSize)

	persistentState, err := c.getPersistentState(&bolt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}
	defer persistentState.Close()

	return c.applyArgs(args, persistentState)
}
