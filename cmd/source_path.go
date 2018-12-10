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

func (c *Config) runSourcePathCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		_, err := fmt.Println(targetState.SourceDir)
		return err
	}
	entries, err := c.getEntries(targetState, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if _, err := fmt.Println(filepath.Join(targetState.SourceDir, entry.SourceName())); err != nil {
			return err
		}
	}
	return nil
}
