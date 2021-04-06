package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type diffCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	recursive bool
	Pager     string `mapstructure:"pager"`
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
	flags.VarP(c.Diff.exclude, "exclude", "x", "exclude entry types")
	flags.VarP(c.Diff.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "recursive")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	gitDiffSystem := chezmoi.NewGitDiffSystem(dryRunSystem, &sb, c.destDirAbsPath, c.color)
	if err := c.applyArgs(gitDiffSystem, c.destDirAbsPath, args, applyArgsOptions{
		include:   c.Diff.include.Sub(c.Diff.exclude),
		recursive: c.Diff.recursive,
		umask:     c.Umask,
	}); err != nil {
		return err
	}
	return c.diffPager(sb.String())
}
