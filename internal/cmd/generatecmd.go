package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/assets/templates"
	"github.com/twpayne/chezmoi/v2/internal/git"
)

func (c *Config) newGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:       "generate file",
		Short:     "Generate a file for use with chezmoi",
		Long:      mustLongHelp("generate"),
		Example:   example("generate"),
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"install.sh"},
		RunE:      c.runGenerateCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
		),
	}

	return generateCmd
}

func (c *Config) runGenerateCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
	builder.Grow(16384)
	switch args[0] {
	case "git-commit-message":
		output, err := c.cmdOutput(c.WorkingTreeAbsPath, c.Git.Command, []string{"status", "--porcelain=v2"})
		if err != nil {
			return err
		}
		status, err := git.ParseStatusPorcelainV2(output)
		if err != nil {
			return err
		}
		data, err := c.gitCommitMessage(cmd, status)
		if err != nil {
			return err
		}
		if _, err := builder.Write(data); err != nil {
			return err
		}
	case "install.sh":
		if _, err := builder.Write(templates.InstallSH); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unsupported file", args[0])
	}
	return c.writeOutputString(builder.String())
}
