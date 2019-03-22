package cmd

import (
	"github.com/spf13/cobra"

	"errors"
	"os"
)

// bashCompletionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:       "completion [shell]",
	Short:     "Output shell completion code for the specified shell (bash or zsh)",
	Long:      "Output shell completion code for the specified shell (bash or zsh)",
	ValidArgs: []string{"bash"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		}
		return errors.New("Unsupported shell")
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
