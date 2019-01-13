package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var catCommand = &cobra.Command{
	Use:   "cat",
	Args:  cobra.MinimumNArgs(1),
	Short: "Write the target state of a file to stdout",
	RunE:  makeRunE(config.runCatCommand),
}

func init() {
	rootCommand.AddCommand(catCommand)
}

func (c *Config) runCatCommand(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	for i, entry := range entries {
		f, ok := entry.(*chezmoi.File)
		if !ok {
			return fmt.Errorf("%s: not a regular file", args[i])
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
