package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type diffCmdConfig struct {
	Command        string                `json:"command"        mapstructure:"command"        yaml:"command"`
	Args           []string              `json:"args"           mapstructure:"args"           yaml:"args"`
	Exclude        *chezmoi.EntryTypeSet `json:"exclude"        mapstructure:"exclude"        yaml:"exclude"`
	Pager          string                `json:"pager"          mapstructure:"pager"          yaml:"pager"`
	PagerArgs      []string              `json:"pagerArgs"      mapstructure:"pagerArgs"      yaml:"pagerArgs"`
	Reverse        bool                  `json:"reverse"        mapstructure:"reverse"        yaml:"reverse"`
	ScriptContents bool                  `json:"scriptContents" mapstructure:"scriptContents" yaml:"scriptContents"`
	include        *chezmoi.EntryTypeSet
	init           bool
	parentDirs     bool
	recursive      bool
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
			dryRun,
			outputsDiff,
			persistentStateModeReadMockWrite,
			requiresSourceDirectory,
		),
	}

	diffCmd.Flags().VarP(c.Diff.Exclude, "exclude", "x", "Exclude entry types")
	diffCmd.Flags().VarP(c.Diff.include, "include", "i", "Include entry types")
	diffCmd.Flags().BoolVar(&c.Diff.init, "init", c.Diff.init, "Recreate config file from template")
	diffCmd.Flags().StringVar(&c.Diff.Pager, "pager", c.Diff.Pager, "Set pager")
	diffCmd.Flags().
		BoolVarP(&c.Diff.parentDirs, "parent-dirs", "P", c.apply.parentDirs, "Print the diff of all parent directories")
	diffCmd.Flags().BoolVarP(&c.Diff.recursive, "recursive", "r", c.Diff.recursive, "Recurse into subdirectories")
	diffCmd.Flags().BoolVar(&c.Diff.Reverse, "reverse", c.Diff.Reverse, "Reverse the direction of the diff")
	diffCmd.Flags().BoolVar(&c.Diff.ScriptContents, "script-contents", c.Diff.ScriptContents, "Show script contents")

	return diffCmd
}

func (c *Config) runDiffCmd(cmd *cobra.Command, args []string) (err error) {
	return c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:        cmd,
		filter:     chezmoi.NewEntryTypeFilter(c.Diff.include.Bits(), c.Diff.Exclude.Bits()),
		init:       c.Diff.init,
		parentDirs: c.Diff.parentDirs,
		recursive:  c.Diff.recursive,
		umask:      c.Umask,
	})
}
