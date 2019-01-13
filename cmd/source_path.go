package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var sourcePathCommand = &cobra.Command{
	Use:   "source-path",
	Short: "Print the source path of the specified targets",
	RunE:  makeRunE(config.runSourcePathCommand),
}

func init() {
	rootCommand.AddCommand(sourcePathCommand)
}

func (c *Config) runSourcePathCommand(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		_, err := fmt.Println(ts.SourceDir)
		return err
	}
	entries, err := c.getEntries(ts, args)
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
