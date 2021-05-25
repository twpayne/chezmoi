package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type diffCmdConfig struct {
	Exclude   *chezmoi.EntryTypeSet `mapstructure:"exclude"`
	Pager     string                `mapstructure:"pager"`
	include   *chezmoi.EntryTypeSet
	recursive bool
}

func (c *Config) newDiffCmd() *cobra.Command {
	diffCmd := &cobra.Command{
		Use:     "diff [target]...",
		Short:   "Print the diff between the target state and the destination state",
		Long:    mustLongHelp("diff"),
		Example: example("diff"),
		RunE:    c.runDiffCmd,
		Annotations: map[string]string{
			persistentStateMode: persistentStateModeReadMockWrite,
		},
	}

	flags := diffCmd.Flags()
	flags.VarP(c.Diff.Exclude, "exclude", "x", "exclude entry types")
	flags.VarP(c.Diff.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "recursive")
	flags.StringVar(&c.Diff.Pager, "pager", c.Diff.Pager, "pager")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	color, err := c.Color.Value()
	if err != nil {
		return err
	}
	gitDiffSystem := chezmoi.NewGitDiffSystem(dryRunSystem, &sb, c.DestDirAbsPath, color)
	if err := c.applyArgs(gitDiffSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:   c.Diff.include.Sub(c.Diff.Exclude),
		recursive: c.Diff.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	return c.pageOutputString(sb.String(), c.Diff.Pager)
}
