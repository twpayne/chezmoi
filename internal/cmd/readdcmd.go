package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"runtime"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type reAddCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	recursive bool
}

// A fileInfo is a simple struct that implements the io/fs.FileInfo interface
// for the purpose of overriding the mode on Windows.
type fileInfo struct {
	name    string
	size    int64
	mode    fs.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *fileInfo) Name() string       { return fi.name }
func (fi *fileInfo) Size() int64        { return fi.size }
func (fi *fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi *fileInfo) ModTime() time.Time { return fi.modTime }
func (fi *fileInfo) IsDir() bool        { return fi.isDir }
func (fi *fileInfo) Sys() any           { return nil } // Sys always returns nil to avoid any inconsistency.

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

	reAddCmd.Flags().VarP(c.reAdd.filter.Exclude, "exclude", "x", "Exclude entry types")
	reAddCmd.Flags().VarP(c.reAdd.filter.Include, "include", "i", "Include entry types")
	reAddCmd.Flags().BoolVarP(&c.reAdd.recursive, "recursive", "r", c.reAdd.recursive, "Recurse into subdirectories")

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
		var err error
		targetRelPaths, err = c.targetRelPaths(sourceState, args, targetRelPathsOptions{
			recursive: c.reAdd.recursive,
		})
		if err != nil {
			return err
		}
		for _, targetRelPath := range targetRelPaths {
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
		if bytes.Equal(actualContents, targetContents) {
			// On Windows, ignore permission changes as they are not preserved
			// by the filesystem. On other systems, if there are no permission
			// changes, continue.
			//
			// See https://github.com/twpayne/chezmoi/issues/3891.
			if runtime.GOOS == "windows" || actualStateFile.Perm() == targetStateFile.Perm(c.Umask) {
				continue
			}
		}

		if c.Interactive {
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
					c.Interactive = false
					break FOR
				case choice == "quit":
					return chezmoi.ExitCodeError(0)
				default:
					panic(choice + ": unexpected choice")
				}
			}
		}

		// On Windows, as the file mode is not preserved by the filesystem, copy
		// the existing mode from the target file. Hack this in by replacing the
		// io/fs.FileInfo of the destination file with a new io/fs.FileInfo with
		// the mode of the target file.
		//
		// See https://github.com/twpayne/chezmoi/issues/3891.
		if runtime.GOOS == "windows" {
			destAbsPathInfo = &fileInfo{
				name:    destAbsPathInfo.Name(),
				size:    destAbsPathInfo.Size(),
				mode:    targetStateFile.Perm(0), // Use the mode from the target.
				modTime: destAbsPathInfo.ModTime(),
			}
		}

		destAbsPathInfos := map[chezmoi.AbsPath]fs.FileInfo{
			destAbsPath: destAbsPathInfo,
		}
		if err := sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
			Encrypt:         sourceStateFile.Attr.Encrypted,
			EncryptedSuffix: c.encryption.EncryptedSuffix(),
			Errorf:          c.errorf,
			Filter:          c.reAdd.filter,
		}); err != nil {
			return err
		}
	}

	return nil
}
