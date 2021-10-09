package cmd

import (
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
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
		RunE:    c.makeRunEWithSourceState(c.runMergeAllCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			requiresSourceDirectory: "true",
		},
	}

	flags := mergeAllCmd.Flags()
	flags.BoolVar(&c.mergeAll.init, "init", c.mergeAll.init, "Recreate config file from template")
	flags.BoolVarP(&c.mergeAll.recursive, "recursive", "r", c.mergeAll.recursive, "Recurse into subdirectories")

	return mergeAllCmd
}

func (c *Config) runMergeAllCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var targetRelPaths []chezmoi.RelPath
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	preApplyFunc := func(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
		if !targetEntryState.Equivalent(actualEntryState) {
			targetRelPaths = append(targetRelPaths, targetRelPath)
		}
		return chezmoi.Skip
	}
	if err := c.applyArgs(cmd.Context(), dryRunSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:      chezmoi.NewEntryTypeSet(chezmoi.EntryTypeFiles),
		init:         c.mergeAll.init,
		recursive:    c.mergeAll.recursive,
		umask:        c.Umask,
		preApplyFunc: preApplyFunc,
	}); err != nil {
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
