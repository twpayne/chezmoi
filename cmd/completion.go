package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var completionCmd = &cobra.Command{
	Use:       "completion shell",
	Short:     "Output shell completion code for the specified shell (bash or zsh)",
	ValidArgs: []string{"bash", "zsh"},
	Args:      cobra.ExactArgs(1),
	RunE:      makeRunE(config.runCompletion),
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func (c *Config) runCompletion(fs vfs.FS, args []string) error {
	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(c.Stdout())
	case "zsh":
		return rootCmd.GenZshCompletion(c.Stdout())
	default:
		return errors.New("unsupported shell")
	}
}
