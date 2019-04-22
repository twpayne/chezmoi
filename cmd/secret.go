package cmd

import "github.com/spf13/cobra"

var secretCmd = &cobra.Command{
	Use:     "secret",
	Args:    cobra.NoArgs,
	Short:   "Interact with a secret manager",
	Long:    mustGetLongHelp("secret"),
	Example: getExample("secret"),
}

func init() {
	rootCmd.AddCommand(secretCmd)
}
