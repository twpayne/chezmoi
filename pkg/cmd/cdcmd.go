package cmd

import (
	"os/exec"

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
			createSourceDirectoryIfNeeded: "true",
			doesNotRequireValidConfig:     "true",
			requiresWorkingTree:           "true",
			runsCommands:                  "true",
		},
	}

	return cdCmd
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	cdCommand, cdArgs := c.cdCommand()
	return c.run(c.WorkingTreeAbsPath, cdCommand, cdArgs)
}

func (c *Config) cdCommand() (string, []string) {
	cdCommand := c.CD.Command
	cdArgs := c.CD.Args

	// If the user has set a cd command then use it.
	if cdCommand != "" {
		return cdCommand, cdArgs
	}

	// Determine the user's shell.
	cdCommand, _ = shell.CurrentUserShell()

	// If the shell is found, return it.
	if path, err := exec.LookPath(cdCommand); err == nil {
		return path, cdArgs
	}

	// Otherwise, if the shell contains spaces, then assume that the first word
	// is the editor and the rest are arguments.
	components := whitespaceRx.Split(cdCommand, -1)
	if len(components) > 1 {
		if path, err := exec.LookPath(components[0]); err == nil {
			return path, append(components[1:], cdArgs...)
		}
	}

	// Fallback to shell command only.
	return cdCommand, cdArgs
}
