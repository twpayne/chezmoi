package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var editCommand = &cobra.Command{
	Use:   "edit",
	Args:  cobra.MinimumNArgs(1),
	Short: "Edit a file",
	RunE:  makeRunE(config.runEditCommandE),
}

func init() {
	rootCommand.AddCommand(editCommand)
}

func (c *Config) runEditCommandE(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	sourceFileNames, err := c.getSourceNames(targetState, args)
	if err != nil {
		return err
	}
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	argv := []string{editor}
	for _, sourceFileName := range sourceFileNames {
		argv = append(argv, filepath.Join(c.SourceDir, sourceFileName))
	}
	return c.exec(argv)
}
