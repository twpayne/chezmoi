package cmd

import (
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type mergeAllCmdConfig struct {
	init      bool
	recursive bool
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
			persistentStateModeReadWrite,
			requiresSourceDirectory,
		),
	}

	mergeAllCmd.Flags().BoolVar(&c.mergeAll.init, "init", c.mergeAll.init, "Recreate config file from template")
	mergeAllCmd.Flags().BoolVarP(&c.mergeAll.recursive, "recursive", "r", c.mergeAll.recursive, "Recurse into subdirectories")

	return mergeAllCmd
}

func (c *Config) runMergeAllCmd(cmd *cobra.Command, args []string) error {
	var targetRelPaths []chezmoi.RelPath
	preApplyFunc := func(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
		if targetEntryState.Type == chezmoi.EntryStateTypeFile && !targetEntryState.Equivalent(actualEntryState) {
			targetRelPaths = append(targetRelPaths, targetRelPath)
		}
		return fs.SkipDir
	}
	if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:          cmd,
		filter:       chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
		init:         c.mergeAll.init,
		recursive:    c.mergeAll.recursive,
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
		if err := c.doMerge(targetRelPath, sourceStateEntry); err != nil {
			return err
		}
	}

	return nil
}
