package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs"
)

var catCommand = &cobra.Command{
	Use:   "cat",
	Args:  cobra.MinimumNArgs(1),
	Short: "Write the contents of a file to stdout",
	RunE:  makeRunE(config.runCatCommand),
}

func init() {
	rootCommand.AddCommand(catCommand)
}

func (c *Config) runCatCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		if err != nil {
			return err
		}
		entry, err := targetState.Get(path)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("%s: not found", arg)
		}
		f, ok := entry.(*chezmoi.File)
		if !ok {
			return fmt.Errorf("%s: not a regular file", arg)
		}
		contents, err := f.Contents()
		if err != nil {
			return err
		}
		if _, err := os.Stdout.Write(contents); err != nil {
			return err
		}
	}
	return nil
}
