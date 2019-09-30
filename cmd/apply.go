package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var applyCmd = &cobra.Command{
	Use:     "apply [targets...]",
	Short:   "Update the destination directory to match the target state",
	Long:    mustGetLongHelp("apply"),
	Example: getExample("apply"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runApplyCmd),
}

func init() {
	rootCmd.AddCommand(applyCmd)
}

func (c *Config) runApplyCmd(fs vfs.FS, args []string) error {
	mutator := c.getDefaultMutator(fs)

	persistentState, err := c.getPersistentState(fs, nil)
	if err != nil {
		return err
	}
	defer persistentState.Close()

	return c.applyArgs(fs, args, mutator, persistentState)
}
