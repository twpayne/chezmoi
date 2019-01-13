package cmd

import (
	"os"

	"github.com/spf13/cobra"
	shell "github.com/twpayne/go-shell"
	vfs "github.com/twpayne/go-vfs"
)

var cdCommand = &cobra.Command{
	Use:   "cd",
	Args:  cobra.NoArgs,
	Short: "Launch a shell in the source directory",
	RunE:  makeRunE(config.runCDCommand),
}

func init() {
	rootCommand.AddCommand(cdCommand)
}

func (c *Config) runCDCommand(fs vfs.FS, args []string) error {
	if err := os.Chdir(c.SourceDir); err != nil {
		return err
	}
	shell, err := shell.CurrentUserShell()
	if err != nil {
		return err
	}
	return c.exec([]string{shell})
}
