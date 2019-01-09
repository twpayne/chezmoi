package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var editConfigCommand = &cobra.Command{
	Use:   "edit-config",
	Args:  cobra.NoArgs,
	Short: "Edit the configuration file",
	RunE:  makeRunE(config.runEditConfigCommand),
}

func init() {
	rootCommand.AddCommand(editConfigCommand)
}

func (c *Config) runEditConfigCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	return c.execEditor(configFile)
}
