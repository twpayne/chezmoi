package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var forgetCmd = &cobra.Command{
	Use:      "forget targets...",
	Aliases:  []string{"unmanage"},
	Args:     cobra.MinimumNArgs(1),
	Short:    "Remove a target from the source state",
	Long:     mustGetLongHelp("forget"),
	Example:  getExample("forget"),
	PreRunE:  config.ensureNoError,
	RunE:     config.runForgetCmd,
	PostRunE: config.autoCommitAndAutoPush,
}

func init() {
	rootCmd.AddCommand(forgetCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(forgetCmd, 1)
}

func (c *Config) runForgetCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := c.mutator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil {
			return err
		}
	}
	return nil
}
