package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var hgCmd = &cobra.Command{
	Use:     "hg [args...]",
	Short:   "Run mercurial in the source directory",
	Long:    mustGetLongHelp("hg"),
	Example: getExample("hg"),
	RunE:    config.runHgCmd,
}

func init() {
	rootCmd.AddCommand(hgCmd)
}

func (c *Config) runHgCmd(cmd *cobra.Command, args []string) error {
	name := "hg"
	if trimExecutableSuffix(filepath.Base(c.SourceVCS.Command)) == "hg" {
		name = c.SourceVCS.Command
	}
	return c.run(c.SourceDir, name, args...)
}
