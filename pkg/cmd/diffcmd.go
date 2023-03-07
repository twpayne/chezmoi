package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type diffCmdConfig struct {
	Command        string                `json:"command" mapstructure:"command" yaml:"command"`
	Args           []string              `json:"args" mapstructure:"args" yaml:"args"`
	Exclude        *chezmoi.EntryTypeSet `json:"exclude" mapstructure:"exclude" yaml:"exclude"`
	Pager          string                `json:"pager" mapstructure:"pager" yaml:"pager"`
	Reverse        bool                  `json:"reverse" mapstructure:"reverse" yaml:"reverse"`
	ScriptContents bool                  `json:"scriptContents" mapstructure:"scriptContents" yaml:"scriptContents"`
	include        *chezmoi.EntryTypeSet
	init           bool
	recursive      bool
	useBuiltinDiff bool
}

func (c *Config) newDiffCmd() *cobra.Command {
	diffCmd := &cobra.Command{
		Use:               "diff [target]...",
		Short:             "Print the diff between the target state and the destination state",
		Long:              mustLongHelp("diff"),
		Example:           example("diff"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runDiffCmd,
		Annotations: newAnnotations(
			persistentStateModeReadMockWrite,
			requiresSourceDirectory,
		),
	}

	flags := diffCmd.Flags()
	flags.VarP(c.Diff.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Diff.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Diff.init, "init", c.Diff.init, "Recreate config file from template")
	flags.StringVar(&c.Diff.Pager, "pager", c.Diff.Pager, "Set pager")
	flags.BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "Recurse into subdirectories")
	flags.BoolVar(&c.Diff.Reverse, "reverse", c.Diff.Reverse, "Reverse the direction of the diff")
	flags.BoolVar(&c.Diff.ScriptContents, "script-contents", c.Diff.ScriptContents, "Show script contents")
	flags.BoolVarP(&c.Diff.useBuiltinDiff, "use-builtin-diff", "", c.Diff.useBuiltinDiff, "Use the builtin diff")

	registerExcludeIncludeFlagCompletionFuncs(diffCmd)

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) (err error) {
	builder := &strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	diffSystem := c.newDiffSystem(dryRunSystem, builder, c.DestDirAbsPath)
	if err = c.applyArgs(cmd.Context(), diffSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:       cmd,
		filter:    chezmoi.NewEntryTypeFilter(c.Diff.include.Bits(), c.Diff.Exclude.Bits()),
		init:      c.Diff.init,
		recursive: c.Diff.recursive,
		umask:     c.Umask,
	}); err != nil {
		return
	}
	if err = c.pageOutputString(builder.String(), c.Diff.Pager); err != nil {
		return
	}
	if closer, ok := diffSystem.(interface {
		Close() error
	}); ok {
		if err = closer.Close(); err != nil {
			return
		}
	}
	return
}
