package cmd

import (
	"github.com/spf13/cobra"
)

var sourceCmd = &cobra.Command{
	Use:     "source [args...]",
	Short:   "Run the source version control system command in the source directory",
	Long:    mustGetLongHelp("source"),
	Example: getExample("source"),
	PreRunE: config.ensureNoError,
	RunE:    config.runSourceCmd,
}

func init() {
	rootCmd.AddCommand(sourceCmd)
}

func (c *Config) runSourceCmd(cmd *cobra.Command, args []string) error {
	return c.run(c.SourceDir, c.SourceVCS.Command, args...)
}
