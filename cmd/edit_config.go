package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var editConfigCommand = &cobra.Command{
	Use:   "edit-config",
	Args:  cobra.NoArgs,
	Short: "Edit the configuration file",
	RunE:  makeRunE(config.runEditConfigCmd),
}

func init() {
	rootCmd.AddCommand(editConfigCommand)
}

func (c *Config) runEditConfigCmd(fs vfs.FS, args []string) error {
	return c.execEditor(c.configFile)
}
