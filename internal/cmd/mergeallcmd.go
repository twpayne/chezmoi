package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type mergeAllCmdConfig struct {
	init      bool
	recursive bool

	Merge mergeCmdConfig `json:"merge" mapstructure:"merge" yaml:"merge"`
}

func (c *Config) newMergeAllCmd() *cobra.Command {
	mergeAllCmd := &cobra.Command{
		Use:     "merge-all",
		Short:   "Perform a three-way merge for each modified file",
		Long:    mustLongHelp("merge-all"),
		Example: example("merge-all"),
		RunE:    c.runMergeAllCmd,
		Annotations: newAnnotations(
			dryRun,
			modifiesSourceDirectory,
			requiresSourceDirectory,
		),
	}

	flags := mergeAllCmd.Flags()
	flags.BoolVar(&c.MergeAll.init, "init", c.MergeAll.init, "Recreate config file from template")
	flags.BoolVarP(
		&c.MergeAll.recursive,
		"recursive",
		"r",
		c.MergeAll.recursive,
		"Recurse into subdirectories",
	)

	return mergeAllCmd
}

func (c *Config) runMergeAllCmd(cmd *cobra.Command, args []string) error {
	var targetRelPaths []chezmoi.RelPath
	preApplyFunc := func(
		targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState,
	) error {
		if targetEntryState.Type == chezmoi.EntryStateTypeFile &&
			!targetEntryState.Equivalent(actualEntryState) {
			targetRelPaths = append(targetRelPaths, targetRelPath)
		}
		return fs.SkipDir
	}
	if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:          cmd,
		filter:       chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
		init:         c.MergeAll.init,
		recursive:    c.MergeAll.recursive,
		umask:        c.Umask,
		preApplyFunc: preApplyFunc,
	}); err != nil {
		return err
	}

	sourceState, err := c.getSourceState(cmd.Context(), cmd)
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		if err := c.doMerge(targetRelPath, sourceStateEntry, c.MergeAll.Merge); err != nil {
			return err
		}
	}

	return nil
}
