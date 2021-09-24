package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
)

type cdCmdConfig struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

func (c *Config) newCDCmd() *cobra.Command {
	cdCmd := &cobra.Command{
		Use:     "cd",
		Short:   "Launch a shell in the source directory",
		Long:    mustLongHelp("cd"),
		Example: example("cd"),
		RunE:    c.runCDCmd,
		Args:    cobra.NoArgs,
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
			requiresSourceDirectory:   "true",
			requiresWorkingTree:       "true",
			runsCommands:              "true",
		},
	}

	return cdCmd
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	shellCommand := c.CD.Command
	if shellCommand == "" {
		shellCommand, _ = shell.CurrentUserShell()
	}
	return c.run(c.WorkingTreeAbsPath, shellCommand, c.CD.Args)
}
