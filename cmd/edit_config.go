package cmd

import (
	"os"
	"path/filepath"

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
	dir := filepath.Dir(c.configFile)
	if _, err := fs.Stat(dir); os.IsNotExist(err) {
		mutator := c.getDefaultMutator(fs)
		if err := vfs.MkdirAll(mutator, dir, 0777); err != nil {
			return err
		}
	}
	return c.execEditor(c.configFile)
}
