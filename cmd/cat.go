package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var catCmd = &cobra.Command{
	Use:     "cat targets...",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Write the target state of a file or symlink to stdout",
	Long:    mustGetLongHelp("cat"),
	Example: getExample("cat"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runCatCmd),
}

func init() {
	rootCmd.AddCommand(catCmd)
}

func (c *Config) runCatCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(fs, ts, args)
	if err != nil {
		return err
	}
	for i, entry := range entries {
		switch entry := entry.(type) {
		case *chezmoi.File:
			contents, err := entry.Contents()
			if err != nil {
				return err
			}
			if _, err := c.Stdout().Write(contents); err != nil {
				return err
			}
		case *chezmoi.Symlink:
			linkname, err := entry.Linkname()
			if err != nil {
				return err
			}
			fmt.Println(linkname)
		default:
			return fmt.Errorf("%s: not a file or symlink", args[i])
		}
	}
	return nil
}
