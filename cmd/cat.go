package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var catCmd = &cobra.Command{
	Use:     "cat targets...",
	Args:    cobra.MinimumNArgs(1),
	Short:   "Write the target state of a file or symlink to stdout",
	Long:    mustGetLongHelp("cat"),
	Example: getExample("cat"),
	PreRunE: config.ensureNoError,
	RunE:    config.runCatCmd,
}

func init() {
	rootCmd.AddCommand(catCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(catCmd, 1)
}

func (c *Config) runCatCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	entries, err := c.getEntries(ts, args)
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
			if _, err := c.Stdout.Write(contents); err != nil {
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
