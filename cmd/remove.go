package cmd

// FIXME add --force flag
// FIXME add --recursive flag
// FIXME add --prompt flag

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var removeCommand = &cobra.Command{
	Use:     "remove",
	Aliases: []string{"rm"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a file or directory",
	RunE:    makeRunE(config.runRemoveCommand),
}

func init() {
	rootCommand.AddCommand(removeCommand)
}

func (c *Config) runRemoveCommand(fs vfs.FS, command *cobra.Command, args []string) error {
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
		if err := mutator.RemoveAll(filepath.Join(c.TargetDir, entry.TargetName())); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := mutator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
