package cmd

import (
	"os"

	"github.com/spf13/cobra"
	shell "github.com/twpayne/go-shell"
	vfs "github.com/twpayne/go-vfs"
)

var cdCmd = &cobra.Command{
	Use:     "cd",
	Args:    cobra.NoArgs,
	Short:   "Launch a shell in the source directory",
	Long:    mustGetLongHelp("cd"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runCDCmd),
}

func init() {
	rootCmd.AddCommand(cdCmd)
}

func (c *Config) runCDCmd(fs vfs.FS, args []string) error {
	mutator := c.getDefaultMutator(fs)
	if err := c.ensureSourceDirectory(fs, mutator); err != nil {
		return err
	}

	if err := os.Chdir(c.SourceDir); err != nil {
		return err
	}
	shell, err := shell.CurrentUserShell()
	if err != nil {
		return err
	}
	return c.exec([]string{shell})
}
