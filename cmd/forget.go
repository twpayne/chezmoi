package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var forgetCommand = &cobra.Command{
	Use:   "forget",
	Args:  cobra.MinimumNArgs(1),
	Short: "Forget a file or directory",
	RunE:  makeRunE(config.runForgetCommand),
}

func init() {
	rootCommand.AddCommand(forgetCommand)
}

func (c *Config) runForgetCommand(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	mutator := c.getDefaultMutator(fs)
	for _, entry := range entries {
		if err := mutator.RemoveAll(filepath.Join(c.SourceDir, entry.SourceName())); err != nil {
			return err
		}
	}
	return nil
}
