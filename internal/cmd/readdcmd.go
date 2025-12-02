package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"runtime"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
)

type reAddCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	reEncrypt bool
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
		GroupID:           groupIDDaily,
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
	reAddCmd.Flags().BoolVar(&c.reAdd.reEncrypt, "re-encrypt", c.reAdd.reEncrypt, "Re-encrypt encrypted files")
	reAddCmd.Flags().BoolVarP(&c.reAdd.recursive, "recursive", "r", c.reAdd.recursive, "Recurse into subdirectories")

	return reAddCmd
}

func (c *Config) runReAddCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var targetRelPaths []chezmoi.RelPath
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
	slices.SortFunc(targetRelPaths, chezmoi.CompareRelPaths)

	// Track which files were already processed to avoid double-processing with exact directories
	processedFiles := make(map[chezmoi.RelPath]bool)

TARGET_REL_PATH:
	for _, targetRelPath := range targetRelPaths {
		sourceStateFile, ok := sourceStateEntries[targetRelPath].(*chezmoi.SourceStateFile)
		if !ok {
			continue
		}
		if sourceStateFile.Attr().Template {
			continue
		}
		if sourceStateFile.Attr().Type != chezmoi.SourceFileTypeFile {
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

		bytesEqual := bytes.Equal(actualContents, targetContents)
		// On Windows, ignore permission changes as they are not preserved
		// by the filesystem. On other systems, if there are no permission
		// changes, continue.
		//
		// See https://github.com/twpayne/chezmoi/issues/3891.
		permsEqual := runtime.GOOS == "windows" || actualStateFile.Perm() == targetStateFile.Perm(c.Umask)
		reEncrypt := sourceStateFile.Attr().Encrypted && c.reAdd.reEncrypt
		if bytesEqual && permsEqual && !reEncrypt {
			continue
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
					if err := c.diffFile(
						targetRelPath,
						c.SourceDirAbsPath.Join(sourceStateFile.SourceRelPath().RelPath()), targetContents, targetStateFile.Perm(c.Umask),
						destAbsPath, actualContents, actualStateFile.Perm(),
					); err != nil {
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
			Encrypt:         sourceStateFile.Attr().Encrypted,
			EncryptedSuffix: c.encryption.EncryptedSuffix(),
			Errorf:          c.errorf,
			Filter:          c.reAdd.filter,
			PreAddFunc:      c.defaultPreAddFunc,
		}); err != nil {
			return err
		}
		// Mark this file as processed
		processedFiles[targetRelPath] = true
	}

	// Process exact directories - add new files and remove deleted files
	return c.processExactDirs(sourceState, sourceStateEntries, processedFiles)
}

// processExactDirs handles exact directory synchronization during re-add.
// It adds new files found in exact target directories and removes files from
// source that no longer exist in exact target directories.
// processedFiles contains files that were already re-added in the file loop,
// outside of the exact directory logic, to avoid double-processing them.
func (c *Config) processExactDirs(
	sourceState *chezmoi.SourceState,
	sourceStateEntries map[chezmoi.RelPath]chezmoi.SourceStateEntry,
	processedFiles map[chezmoi.RelPath]bool,
) error {
	// Collect all exact directories from source state that are in scope
	// An exact directory is in scope if:
	// 1. It's directly in sourceStateEntries (was explicitly or implicitly targeted)
	// 2. Any of its children are in sourceStateEntries (parent dir needs sync)
	exactDirs := make(map[chezmoi.RelPath]*chezmoi.SourceStateDir)

	// First, collect exact directories that are directly in the entries
	for targetRelPath, entry := range sourceStateEntries {
		if sourceStateDir, ok := entry.(*chezmoi.SourceStateDir); ok && sourceStateDir.Attr().Exact {
			exactDirs[targetRelPath] = sourceStateDir
		}
	}

	// Then, collect parent exact directories for any entries
	for targetRelPath := range sourceStateEntries {
		// Walk up the path to find parent exact directories
		for parentPath := targetRelPath.Dir(); parentPath != chezmoi.DotRelPath; parentPath = parentPath.Dir() {
			// Check if this parent is an exact directory in source state
			parentEntry := sourceState.Get(parentPath)
			if sourceStateDir, ok := parentEntry.(*chezmoi.SourceStateDir); ok && sourceStateDir.Attr().Exact {
				// Only add if not already present
				if _, exists := exactDirs[parentPath]; !exists {
					exactDirs[parentPath] = sourceStateDir
				}
			}
		}
	}

	// Process each exact directory
	for targetRelPath := range exactDirs {
		targetDirAbsPath := c.DestDirAbsPath.Join(targetRelPath)

		// Read the target directory contents
		dirEntries, err := c.destSystem.ReadDir(targetDirAbsPath)
		if err != nil {
			// If directory doesn't exist in target, skip it
			continue
		}

		// Build sets of files in source and target
		sourceFiles := make(map[string]chezmoi.SourceStateEntry)
		targetFiles := make(map[string]fs.DirEntry)

		// Collect files from source state that are in this exact directory
		// Only consider actual files, not removes or other entry types
		for entryRelPath, entry := range sourceStateEntries {
			// Check if this entry is a direct child of the exact directory
			if entryRelPath.Dir() == targetRelPath {
				// Only count actual files in source, not remove entries
				if _, ok := entry.(*chezmoi.SourceStateFile); ok {
					sourceFiles[entryRelPath.Base()] = entry
				}
			}
		}

		// Collect files from target directory
		for _, dirEntry := range dirEntries {
			name := dirEntry.Name()
			if name == "." || name == ".." {
				continue
			}
			targetFiles[name] = dirEntry
		}

		// Find new files (in target but not in source) and add them
		destAbsPathInfos := make(map[chezmoi.AbsPath]fs.FileInfo)
		for name, dirEntry := range targetFiles {
			entryRelPath := targetRelPath.JoinString(name)

			// Skip files that were already processed in the file re-add loop
			if processedFiles[entryRelPath] {
				continue
			}

			if _, inSource := sourceFiles[name]; !inSource {
				// Check if ignored
				if sourceState.Ignore(entryRelPath) {
					continue
				}

				// Skip subdirectories that are themselves exact directories
				// They will be processed separately in their own exact directory processing
				if dirEntry.IsDir() {
					childEntry := sourceState.Get(entryRelPath)
					if childDir, ok := childEntry.(*chezmoi.SourceStateDir); ok && childDir.Attr().Exact {
						continue
					}
				}

				// This is a new file - add it
				// Get full FileInfo (DirEntry only has partial info)
				destAbsPath := targetDirAbsPath.JoinString(name)
				fileInfo, err := dirEntry.Info()
				if err != nil {
					return err
				}
				destAbsPathInfos[destAbsPath] = fileInfo
			}
		}

		// Add new files if any were found
		if len(destAbsPathInfos) > 0 {
			if err := sourceState.Add(c.sourceSystem, c.persistentState, c.destSystem, destAbsPathInfos, &chezmoi.AddOptions{
				Filter:     c.reAdd.filter,
				PreAddFunc: c.defaultPreAddFunc,
			}); err != nil {
				return err
			}
		}

		// Find deleted files (in source but not in target) and remove them
		for name, sourceEntry := range sourceFiles {
			if _, inTarget := targetFiles[name]; !inTarget {
				// Check if ignored
				entryRelPath := targetRelPath.JoinString(name)
				if sourceState.Ignore(entryRelPath) {
					continue
				}

				// This file was deleted from target - remove it from source
				sourceAbsPath := c.SourceDirAbsPath.Join(sourceEntry.SourceRelPath().RelPath())
				if err := c.sourceSystem.RemoveAll(sourceAbsPath); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
