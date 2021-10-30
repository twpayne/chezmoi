package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/multierr"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type diffCmdConfig struct {
	Command        string                `mapstructure:"command"`
	Args           []string              `mapstructure:"args"`
	Exclude        *chezmoi.EntryTypeSet `mapstructure:"exclude"`
	Pager          string                `mapstructure:"pager"`
	include        *chezmoi.EntryTypeSet
	init           bool
	recursive      bool
	useBuiltinDiff bool
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
	flags.VarP(c.Diff.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Diff.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Diff.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "Recurse into subdirectories")
	flags.StringVar(&c.Diff.Pager, "pager", c.Diff.Pager, "Set pager")
	flags.BoolVarP(&c.Diff.useBuiltinDiff, "use-builtin-diff", "", c.Diff.useBuiltinDiff, "Use the builtin diff")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) (err error) {
	builder := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	if c.Diff.useBuiltinDiff || c.Diff.Command == "" {
		color := c.Color.Value(c.colorAutoFunc)
		gitDiffSystem := chezmoi.NewGitDiffSystem(dryRunSystem, &builder, c.DestDirAbsPath, color)
		if err = c.applyArgs(cmd.Context(), gitDiffSystem, c.DestDirAbsPath, args, applyArgsOptions{
			include:   c.Diff.include.Sub(c.Diff.Exclude),
			init:      c.Diff.init,
			recursive: c.Diff.recursive,
			umask:     c.Umask,
		}); err != nil {
			return
		}
		err = c.pageOutputString(builder.String(), c.Diff.Pager)
		return
	}
	diffSystem := chezmoi.NewExternalDiffSystem(dryRunSystem, c.Diff.Command, c.Diff.Args, c.DestDirAbsPath)
	defer func() {
		err = multierr.Append(err, diffSystem.Close())
	}()
	err = c.applyArgs(cmd.Context(), diffSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:   c.Diff.include.Sub(c.Diff.Exclude),
		init:      c.Diff.init,
		recursive: c.Diff.recursive,
		umask:     c.Umask,
	})
	return
}
