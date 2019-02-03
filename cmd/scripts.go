package cmd

import (
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var scriptsCmd = &cobra.Command{
	Use:   "scripts [targets...]",
	Short: "Run scripts that need to run",
	RunE:  makeRunE(config.runScriptsCmd),
}

func init() {
	rootCmd.AddCommand(scriptsCmd)
}

func (c *Config) runScriptsCmd(fs vfs.FS, args []string) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	return ts.ApplyScripts()
}
