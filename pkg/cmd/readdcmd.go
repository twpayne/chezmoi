package cmd

import (
	"bytes"
	"io/fs"
	"sort"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type reAddCmdConfig struct {
	filter *chezmoi.EntryTypeFilter
}

func (c *Config) newReAddCmd() *cobra.Command {
	reAddCmd := &cobra.Command{
		Use:               "re-add",
		Short:             "Re-add modified files",
		Long:              mustLongHelp("re-add"),
		Example:           example("re-add"),
		ValidArgsFunction: c.targetValidArgs,
		Args:              cobra.NoArgs,
		RunE:              c.makeRunEWithSourceState(c.runReAddCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
			requiresSourceDirectory: "true",
		},
	}

	flags := reAddCmd.Flags()
	flags.VarP(c.reAdd.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.reAdd.filter.Include, "include", "i", "Include entry types")

	registerExcludeIncludeFlagCompletionFuncs(reAddCmd)

	return reAddCmd
}

func (c *Config) runReAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var targetRelPaths chezmoi.RelPaths
	sourceStateEntries := make(map[chezmoi.RelPath]chezmoi.SourceStateEntry)
	_ = sourceState.ForEach(func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
		targetRelPaths = append(targetRelPaths, targetRelPath)
		sourceStateEntries[targetRelPath] = sourceStateEntry
		return nil
	})
	sort.Sort(targetRelPaths)

	for _, targetRelPath := range targetRelPaths {
		sourceStateFile, ok := sourceStateEntries[targetRelPath].(*chezmoi.SourceStateFile)
		if !ok {
			continue
		}
		if sourceStateFile.Attr.Template {
			continue
		}
		if sourceStateFile.Attr.Type != chezmoi.SourceFileTypeFile {
			continue
		}

		destAbsPath := c.DestDirAbsPath.Join(targetRelPath)
		destAbsPathInfo, err := c.destSystem.Stat(destAbsPath)
		actualState, err := chezmoi.NewActualStateEntry(c.destSystem, destAbsPath, destAbsPathInfo, err)
		if err != nil {
			return err
		}
		actualStateFile, ok := actualState.(*chezmoi.ActualStateFile)
		if !ok {
			continue
		}

		targetState, err := sourceStateFile.TargetStateEntry(c.destSystem, c.DestDirAbsPath)
		if err != nil {
			return err
		}
		targetStateFile, ok := targetState.(*chezmoi.TargetStateFile)
		if !ok {
			continue
		}

		actualContents, err := actualStateFile.Contents()
		if err != nil {
			return err
		}
		targetContents, err := targetStateFile.Contents()
		if err != nil {
			return err
		}
		if bytes.Equal(actualContents, targetContents) && actualStateFile.Perm() == targetStateFile.Perm(c.Umask) {
			continue
		}

		destAbsPathInfos := map[chezmoi.AbsPath]fs.FileInfo{
			destAbsPath: destAbsPathInfo,
		}
		if err := sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
			Encrypt:         sourceStateFile.Attr.Encrypted,
			EncryptedSuffix: c.encryption.EncryptedSuffix(),
			Filter:          c.reAdd.filter,
		}); err != nil {
			return err
		}
	}

	return nil
}
