package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type removeCmdConfig struct {
	force bool
}

var removeCmd = &cobra.Command{
	Use:      "remove targets...",
	Aliases:  []string{"rm"},
	Args:     cobra.MinimumNArgs(1),
	Short:    "Remove a target from the source state and the destination directory",
	Long:     mustGetLongHelp("remove"),
	Example:  getExample("remove"),
	PreRunE:  config.ensureNoError,
	RunE:     config.runRemoveCmd,
	PostRunE: config.autoCommitAndAutoPush,
}

func init() {
	rootCmd.AddCommand(removeCmd)

	persistentFlags := removeCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.remove.force, "force", "f", false, "remove without prompting")

	markRemainingZshCompPositionalArgumentsAsFiles(removeCmd, 1)
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return nil
	}
	for _, entry := range entries {
		destDirPath := filepath.Join(c.DestDir, entry.TargetName())
		sourceDirPath := filepath.Join(c.SourceDir, entry.SourceName())
		if !c.remove.force {
			choice, err := c.prompt(fmt.Sprintf("Remove %s and %s", destDirPath, sourceDirPath), "ynqa")
			if err != nil {
				return err
			}
			switch choice {
			case 'y':
			case 'n':
				continue
			case 'q':
				return nil
			case 'a':
				c.remove.force = true
			}
		}
		if err := c.mutator.RemoveAll(destDirPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := c.mutator.RemoveAll(sourceDirPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
