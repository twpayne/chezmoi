package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:       "completion shell",
	Args:      cobra.ExactArgs(1),
	Short:     "Output shell completion code for the specified shell (bash, fish, or zsh)",
	Long:      mustGetLongHelp("completion"),
	Example:   getExample("completion"),
	ValidArgs: []string{"bash", "fish", "zsh"},
	RunE:      config.runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

func (c *Config) runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return rootCmd.GenBashCompletion(c.Stdout)
	case "zsh":
		return rootCmd.GenZshCompletion(c.Stdout)
	case "fish":
		return rootCmd.GenFishCompletion(c.Stdout)
	default:
		return errors.New("unsupported shell")
	}
}
