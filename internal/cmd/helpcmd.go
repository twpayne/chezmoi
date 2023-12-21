package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func (c *Config) newHelpCmd() *cobra.Command {
	helpCmd := &cobra.Command{
		Use:     "help [command]",
		Short:   "Print help about a command",
		Long:    mustLongHelp("help"),
		Example: example("help"),
		RunE:    c.runHelpCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
		),
	}

	return helpCmd
}

func (c *Config) runHelpCmd(cmd *cobra.Command, args []string) error {
	subCmd, _, err := cmd.Root().Find(args)
	if err != nil {
		return err
	}
	if subCmd == nil {
		return fmt.Errorf("unknown command: %s", strings.Join(args, " "))
	}
	return subCmd.Help()
}
