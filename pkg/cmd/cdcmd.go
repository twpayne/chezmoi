package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/shell"
)

type cdCmdConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args" mapstructure:"args" yaml:"args"`
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
			createSourceDirectoryIfNeeded: "true",
			doesNotRequireValidConfig:     "true",
			requiresWorkingTree:           "true",
			runsCommands:                  "true",
		},
	}

	return cdCmd
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	cdCommand, cdArgs, err := c.cdCommand()
	if err != nil {
		return err
	}
	return c.run(c.WorkingTreeAbsPath, cdCommand, cdArgs)
}

func (c *Config) cdCommand() (string, []string, error) {
	cdCommand := c.CD.Command
	cdArgs := c.CD.Args

	if cdCommand != "" {
		return cdCommand, cdArgs, nil
	}

	cdCommand, _ = shell.CurrentUserShell()
	return parseCommand(cdCommand, cdArgs)
}
