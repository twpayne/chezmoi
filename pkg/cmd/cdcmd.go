package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/shell"
)

type cdCmdConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args" mapstructure:"args" yaml:"args"`
}

func (c *Config) newCDCmd() *cobra.Command {
	cdCmd := &cobra.Command{
		Use:     "cd [path]",
		Short:   "Launch a shell in the source directory",
		Long:    mustLongHelp("cd"),
		Example: example("cd"),
		RunE:    c.runCDCmd,
		Args:    cobra.MaximumNArgs(1),
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			requiresWorkingTree,
			runsCommands,
			runsWithInvalidConfig,
		),
	}

	return cdCmd
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	cdCommand, cdArgs, err := c.cdCommand()
	if err != nil {
		return err
	}
	var dir chezmoi.AbsPath
	if len(args) == 0 {
		dir = c.WorkingTreeAbsPath
	} else if argAbsPath, err := chezmoi.NewAbsPathFromExtPath(args[0], c.homeDirAbsPath); err != nil {
		return err
	} else if argAbsPath == c.DestDirAbsPath {
		dir, err = c.getSourceDirAbsPath(nil)
		if err != nil {
			return err
		}
	} else {
		sourceState, err := c.getSourceState(cmd.Context())
		if err != nil {
			return err
		}
		sourceAbsPaths, err := c.sourceAbsPaths(sourceState, args)
		if err != nil {
			return err
		}
		dir = sourceAbsPaths[0]
	}
	return c.run(dir, cdCommand, cdArgs)
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
