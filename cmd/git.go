package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"
)

var gitCmd = &cobra.Command{
	Use:     "git [args...]",
	Short:   "Run git in the source directory",
	Long:    mustGetLongHelp("git"),
	Example: getExample("git"),
	RunE:    config.runGitCmd,
}

func init() {
	rootCmd.AddCommand(gitCmd)
}

func (c *Config) runGitCmd(cmd *cobra.Command, args []string) error {
	name := "git"
	if trimExecutableSuffix(filepath.Base(c.SourceVCS.Command)) == "git" {
		name = c.SourceVCS.Command
	}
	return c.run(c.SourceDir, name, args...)
}
