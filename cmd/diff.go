package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	bolt "go.etcd.io/bbolt"
)

var diffCmd = &cobra.Command{
	Use:     "diff [targets...]",
	Short:   "Write the diff between the target state and the destination state to stdout",
	Long:    mustGetLongHelp("diff"),
	Example: getExample("diff"),
	PreRunE: config.ensureNoError,
	RunE:    config.runDiffCmd,
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	c.DryRun = true
	c.mutator = chezmoi.NewVerboseMutator(c.Stdout(), chezmoi.NullMutator{}, c.colored)

	persistentState, err := c.getPersistentState(&bolt.Options{
		ReadOnly: true,
	})
	if err != nil {
		return err
	}
	defer persistentState.Close()

	return c.applyArgs(args, persistentState)
}
