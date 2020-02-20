package cmd

import (
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:     "apply [targets...]",
	Short:   "Update the destination directory to match the target state",
	Long:    mustGetLongHelp("apply"),
	Example: getExample("apply"),
	PreRunE: config.ensureNoError,
	RunE:    config.runApplyCmd,
}

func init() {
	rootCmd.AddCommand(applyCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(applyCmd, 1)
}

func (c *Config) runApplyCmd(cmd *cobra.Command, args []string) error {
	persistentState, err := c.getPersistentState(nil)
	if err != nil {
		return err
	}
	defer persistentState.Close()

	return c.applyArgs(args, persistentState)
}
