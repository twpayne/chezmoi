package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var sourcePathCmd = &cobra.Command{
	Use:     "source-path [targets...]",
	Short:   "Print the path of a target in the source state",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runSourcePathCmd),
}

func init() {
	rootCmd.AddCommand(sourcePathCmd)
}

func (c *Config) runSourcePathCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		_, err := fmt.Println(ts.SourceDir)
		return err
	}
	entries, err := c.getEntries(fs, ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err := fmt.Println(filepath.Join(ts.SourceDir, entry.SourceName())); err != nil {
			return err
		}
	}
	return nil
}
