package cmd

import "github.com/spf13/cobra"

var cdCmd = &cobra.Command{
	Use:     "cd",
	Args:    cobra.NoArgs,
	Short:   "Launch a shell in the source directory",
	Long:    mustGetLongHelp("cd"),
	Example: getExample("cd"),
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runCDCmd),
}

func init() {
	rootCmd.AddCommand(cdCmd)
}
