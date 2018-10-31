package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "chezmoi",
	Short: "chezmoi manages your home directory",
	Run:   makeRun(runRootCommand),
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runRootCommand(cmd *cobra.Command, args []string) error {
	return nil
}
