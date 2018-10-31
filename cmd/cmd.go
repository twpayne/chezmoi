package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func makeRun(runCommand func(*cobra.Command, []string) error) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := runCommand(cmd, args); err != nil {
			log.Fatal(err)
		}
	}
}
