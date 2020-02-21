package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:     "help [command]",
	Short:   "Print help about a command",
	Long:    mustGetLongHelp("help"),
	Example: getExample("help"),
	RunE:    config.runHelpCmd,
}

func init() {
	rootCmd.SetHelpCommand(helpCmd)
}

func (c *Config) runHelpCmd(cmd *cobra.Command, args []string) error {
	subCmd, _, err := rootCmd.Find(args)
	if err != nil {
		return err
	}
	if subCmd == nil {
		return fmt.Errorf("unknown command: %q", strings.Join(args, " "))
	}
	return subCmd.Help()
}
