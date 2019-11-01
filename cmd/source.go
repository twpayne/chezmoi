package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var sourceCmd = &cobra.Command{
	Use:     "source [args...]",
	Short:   "Run the source version control system command in the source directory",
	Long:    mustGetLongHelp("source"),
	Example: getExample("source"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runSourceCmd),
}

func init() {
	rootCmd.AddCommand(sourceCmd)
}

func (c *Config) runSourceCmd(fs vfs.FS, args []string) error {
	return c.run(fs, c.SourceDir, c.SourceVCS.Command, args...)
}
