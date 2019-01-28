package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var removeCmd = &cobra.Command{
	Use:     "remove targets...",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state and the destination directory",
	RunE:    makeRunE(config.runRemoveCmd),
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func (c *Config) runRemoveCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return nil
	}
	mutator := c.getDefaultMutator(fs)
	for _, entry := range entries {
		if err := mutator.RemoveAll(filepath.Join(c.DestDir, entry.TargetName())); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := mutator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
