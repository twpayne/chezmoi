package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

type removeCmdConfig struct {
	force bool
}

var removeCmd = &cobra.Command{
	Use:     "remove targets...",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state and the destination directory",
	Long:    mustGetLongHelp("remove"),
	Example: getExample("remove"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runRemoveCmd),
}

func init() {
	rootCmd.AddCommand(removeCmd)

	persistentFlags := removeCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.remove.force, "force", "f", false, "remove without prompting")
}

func (c *Config) runRemoveCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs, nil)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(fs, ts, args)
	if err != nil {
		return nil
	}
	mutator := c.getDefaultMutator(fs)
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
		if err := mutator.RemoveAll(destDirPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := mutator.RemoveAll(sourceDirPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
