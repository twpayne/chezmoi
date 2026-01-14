package cmd

import (
	"fmt"
	"io/fs"
	"slices"

	"github.com/spf13/cobra"

	"chezmoi.io/chezmoi/internal/chezmoi"
	"chezmoi.io/chezmoi/internal/chezmoiset"
)

type unmanagedCmdConfig struct {
	filter           *chezmoi.EntryTypeFilter
	nulPathSeparator bool
	pathStyle        *choiceFlag
	tree             bool
}

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		GroupID: groupIDAdvanced,
		Use:     "unmanaged [path]...",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}

	unmanagedCmd.Flags().VarP(c.unmanaged.filter.Exclude, "exclude", "x", "Exclude entry types")
	unmanagedCmd.Flags().VarP(c.unmanaged.filter.Include, "include", "i", "Include entry types")
	unmanagedCmd.Flags().
		BoolVarP(&c.unmanaged.nulPathSeparator, "nul-path-separator", "0", c.unmanaged.nulPathSeparator, "Use the NUL character as a path separator")
	unmanagedCmd.Flags().VarP(c.unmanaged.pathStyle, "path-style", "p", "Path style")
	must(unmanagedCmd.RegisterFlagCompletionFunc("path-style", c.unmanaged.pathStyle.FlagCompletionFunc()))
	unmanagedCmd.Flags().BoolVarP(&c.unmanaged.tree, "tree", "t", c.unmanaged.tree, "Print paths as a tree")

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var absPaths []chezmoi.AbsPath
	if len(args) == 0 {
		absPaths = append(absPaths, c.DestDirAbsPath)
	} else {
		argsAbsPaths := chezmoiset.New[chezmoi.AbsPath]()
		for _, arg := range args {
			argAbsPath, err := chezmoi.NormalizePath(arg)
			if err != nil {
				return err
			}
			argsAbsPaths.Add(argAbsPath)
		}
		absPaths = argsAbsPaths.Elements()
		slices.Sort(absPaths)
	}

	unmanagedRelPaths := chezmoiset.New[chezmoi.RelPath]()
	walkFunc := func(destAbsPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			c.errorf("%s: %v\n", destAbsPath, err)
			if fileInfo == nil || fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if destAbsPath == c.DestDirAbsPath {
			return nil
		}
		targetRelPath, err := destAbsPath.TrimDirPrefix(c.DestDirAbsPath)
		if err != nil {
			return err
		}
		sourceStateEntry := sourceState.Get(targetRelPath)
		managed := sourceStateEntry != nil
		ignored := sourceState.Ignore(targetRelPath)
		included := c.unmanaged.filter.IncludeFileInfo(fileInfo)
		if !managed && !ignored && included {
			unmanagedRelPaths.Add(targetRelPath)
		}
		if fileInfo.IsDir() {
			switch {
			case !managed:
				return fs.SkipDir
			case ignored:
				return fs.SkipDir
			case sourceStateEntry != nil:
				if external, ok := sourceStateEntry.Origin().(*chezmoi.External); ok {
					if external.Type == chezmoi.ExternalTypeGitRepo {
						return fs.SkipDir
					}
				}
			}
		}
		return nil
	}
	for _, absPath := range absPaths {
		if err := chezmoi.Walk(c.destSystem, absPath, walkFunc); err != nil {
			return err
		}
	}

	paths := make([]fmt.Stringer, 0, len(unmanagedRelPaths.Elements()))
	for relPath := range unmanagedRelPaths {
		var path fmt.Stringer
		switch pathStyle := c.unmanaged.pathStyle.String(); pathStyle {
		case pathStyleAbsolute:
			path = c.DestDirAbsPath.Join(relPath)
		case pathStyleRelative:
			path = relPath
		default:
			return fmt.Errorf("%s: invalid path style", pathStyle)
		}
		paths = append(paths, path)
	}

	return c.writePaths(stringersToStrings(paths), writePathsOptions{
		nulPathSeparator: c.unmanaged.nulPathSeparator,
		tree:             c.unmanaged.tree,
	})
}
