package cmd

import "github.com/spf13/cobra"

var secretCommand = &cobra.Command{
	Use:   "secret",
	Args:  cobra.NoArgs,
	Short: "Interact with a secret manager",
}

func init() {
	rootCommand.AddCommand(secretCommand)
}
