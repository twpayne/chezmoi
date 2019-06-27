package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var forgetCmd = &cobra.Command{
	Use:     "forget targets...",
	Aliases: []string{"unmanage"},
	Args:    cobra.MinimumNArgs(1),
	Short:   "Remove a target from the source state",
	Long:    mustGetLongHelp("forget"),
	Example: getExample("forget"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runForgetCmd),
}

func init() {
	rootCmd.AddCommand(forgetCmd)
}

func (c *Config) runForgetCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs, nil)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(fs, ts, args)
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
