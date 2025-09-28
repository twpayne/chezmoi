package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type cdCmdConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
}

func (c *Config) newCDCmd() *cobra.Command {
	cdCmd := &cobra.Command{
		GroupID: groupIDAdvanced,
		Use:     "cd [path]",
		Short:   "Launch a shell in the source directory",
		Long:    mustLongHelp("cd"),
		Example: example("cd"),
		RunE:    c.runCDCmd,
		Args:    cobra.MaximumNArgs(1),
		Annotations: newAnnotations(
			createSourceDirectoryIfNeeded,
			doesNotRequireValidConfig,
			persistentStateModeReadWrite,
			requiresWorkingTree,
			runsCommands,
		),
	}

	return cdCmd
}

func (c *Config) runCDCmd(cmd *cobra.Command, args []string) error {
	os.Setenv("CHEZMOI_SUBSHELL", "1")

	cdCommand, cdArgs, err := c.cdCommand()
	if err != nil {
		return err
	}
	var dir chezmoi.AbsPath
	if len(args) == 0 {
		dir = c.WorkingTreeAbsPath
	} else {
		switch argAbsPath, err := chezmoi.NewAbsPathFromExtPath(args[0], c.homeDirAbsPath); {
		case err != nil:
			return err
		case argAbsPath == c.DestDirAbsPath:
			dir, err = c.getSourceDirAbsPath(nil)
			if err != nil {
				return err
			}
		default:
			sourceState, err := c.getSourceState(cmd.Context(), cmd)
			if err != nil {
				return err
			}
			sourceAbsPaths, err := c.sourceAbsPaths(sourceState, args)
			if err != nil {
				return err
			}
			dir = sourceAbsPaths[0]
		}
	}

	switch fileInfo, err := c.baseSystem.Stat(dir); {
	case err != nil:
		return err
	case !fileInfo.IsDir():
		return fmt.Errorf("%s: not a directory", dir)
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
