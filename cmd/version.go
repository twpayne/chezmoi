package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Args:  cobra.NoArgs,
	Short: "Write the version of chezmoi to stdout",
	RunE:  makeRunE(config.runVersionCommand),
}

func init() {
	rootCommand.AddCommand(versionCommand)
}

func (c *Config) runVersionCommand(fs vfs.FS, cmd *cobra.Command, args []string) error {
	fmt.Printf("Version: %s Commit: %s Date: %s\n", c.version.Version, c.version.Commit, c.version.Date)
	return nil
}
