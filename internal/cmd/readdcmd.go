package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
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
		Args:              cobra.ArbitraryArgs,
		RunE:              c.makeRunEWithSourceState(c.runReAddCmd),
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
			requiresSourceDirectory,
		),
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
	if len(args) == 0 {
		_ = sourceState.ForEach(
			func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
				targetRelPaths = append(targetRelPaths, targetRelPath)
				sourceStateEntries[targetRelPath] = sourceStateEntry
				return nil
			},
		)
	} else {
		for _, arg := range args {
			arg = filepath.Clean(arg)
			destAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
			if err != nil {
				return err
			}
			targetRelPath, err := c.targetRelPath(destAbsPath)
			if err != nil {
				return err
			}
			targetRelPaths = append(targetRelPaths, targetRelPath)
			sourceStateEntries[targetRelPath] = sourceState.Get(targetRelPath)
		}
	}
	sort.Sort(targetRelPaths)

TARGET_REL_PATH:
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

		if c.interactive {
			prompt := fmt.Sprintf("Re-add %s", targetRelPath)
			var choices []string
			if actualContents != nil || targetContents != nil {
				choices = append(choices, "diff")
			}
			choices = append(choices, choicesYesNoAllQuit...)
		FOR:
			for {
				switch choice, err := c.promptChoice(prompt, choices); {
				case err != nil:
					return err
				case choice == "diff":
					if err := c.diffFile(targetRelPath, targetContents, targetStateFile.Perm(c.Umask), actualContents, actualStateFile.Perm()); err != nil {
						return err
					}
				case choice == "yes":
					break FOR
				case choice == "no":
					continue TARGET_REL_PATH
				case choice == "all":
					c.interactive = false
					break FOR
				case choice == "quit":
					return chezmoi.ExitCodeError(0)
				default:
					panic(fmt.Sprintf("%s: unexpected choice", choice))
				}
			}
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
