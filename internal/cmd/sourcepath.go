package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var sourcePathCmd = &cobra.Command{
	Use:     "source-path [targets...]",
	Short:   "Print the path of a target in the source state",
	Long:    mustGetLongHelp("source-path"),
	Example: getExample("source-path"),
	PreRunE: config.ensureNoError,
	RunE:    config.runSourcePathCmd,
}

func init() {
	rootCmd.AddCommand(sourcePathCmd)

	markRemainingZshCompPositionalArgumentsAsFiles(sourcePathCmd, 1)
}

func (c *Config) runSourcePathCmd(cmd *cobra.Command, args []string) error {
	ts, err := c.getTargetState(nil)
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
