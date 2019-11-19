package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
)

var cdCmd = &cobra.Command{
	Use:     "cd",
	Args:    cobra.NoArgs,
	Short:   "Launch a shell in the source directory",
	Long:    mustGetLongHelp("cd"),
	Example: getExample("cd"),
	PreRunE: config.ensureNoError,
	RunE:    config.runCDCmd,
}

func init() {
	rootCmd.AddCommand(cdCmd)
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	if err := c.ensureSourceDirectory(); err != nil {
		return err
	}

	shell, _ := shell.CurrentUserShell()
	return c.run(c.SourceDir, shell)
}
