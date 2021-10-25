package cmd

import (
	"bytes"
	"io/fs"
	"sort"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type reAddCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	recursive bool
}

func (c *Config) newReAddCmd() *cobra.Command {
	reAddCmd := &cobra.Command{
		Use:     "re-add [targets...]",
		Short:   "Re-add modified files",
		Long:    mustLongHelp("re-add"),
		Example: example("re-add"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runReAddCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
			requiresSourceDirectory: "true",
		},
	}

	flags := reAddCmd.Flags()
	flags.VarP(c.reAdd.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.reAdd.include, "include", "i", "Include entry types")
	flags.BoolVarP(&c.reAdd.recursive, "recursive", "r", c.reAdd.recursive, "Recurse into subdirectories")

	return reAddCmd
}

func (c *Config) runReAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sourceStateEntries := sourceState.Entries()
	destRelPaths := make(chezmoi.RelPaths, 0, len(sourceStateEntries))
	for destRelPath := range sourceStateEntries {
		destRelPaths = append(destRelPaths, destRelPath)
	}
	sort.Sort(destRelPaths)

	for _, destRelPath := range destRelPaths {
		sourceStateFile, ok := sourceStateEntries[destRelPath].(*chezmoi.SourceStateFile)
		if !ok {
			continue
		}
		if sourceStateFile.Attr.Template {
			continue
		}
		if sourceStateFile.Attr.Type != chezmoi.SourceFileTypeFile {
			continue
		}

		destAbsPath := c.DestDirAbsPath.Join(destRelPath)
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
			Empty:           sourceStateFile.Attr.Empty,
			Encrypt:         sourceStateFile.Attr.Encrypted,
			EncryptedSuffix: c.encryption.EncryptedSuffix(),
			Include:         c.reAdd.include.Sub(c.reAdd.exclude),
		}); err != nil {
			return err
		}
	}

	return nil
}
