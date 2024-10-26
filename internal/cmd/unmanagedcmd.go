package cmd

import (
	"fmt"
	"io/fs"
	"sort"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

type unmanagedCmdConfig struct {
	pathStyle chezmoi.PathStyleSimple
	tree      bool
}

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		Use:         "unmanaged [path]...",
		Short:       "List the unmanaged files in the destination directory",
		Long:        mustLongHelp("unmanaged"),
		Example:     example("unmanaged"),
		Args:        cobra.ArbitraryArgs,
		RunE:        c.makeRunEWithSourceState(c.runUnmanagedCmd),
		Annotations: newAnnotations(),
	}

	unmanagedCmd.Flags().VarP(&c.unmanaged.pathStyle, "path-style", "p", "Path style")
	unmanagedCmd.Flags().BoolVarP(&c.unmanaged.tree, "tree", "t", c.unmanaged.tree, "Print paths as a tree")

	if err := unmanagedCmd.RegisterFlagCompletionFunc("path-style", chezmoi.PathStyleSimpleFlagCompletionFunc); err != nil {
		panic(err)
	}

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var absPaths chezmoi.AbsPaths
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
		absPaths = chezmoi.AbsPaths(argsAbsPaths.Elements())
		sort.Sort(absPaths)
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
		if !managed && !ignored {
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
		if c.unmanaged.pathStyle.ToPathStyle() == chezmoi.PathStyleAbsolute {
			path = c.DestDirAbsPath.Join(relPath)
		} else {
			path = relPath
		}
		paths = append(paths, path)
	}

	return c.writePaths(stringersToStrings(paths), writePathsOptions{
		tree: c.unmanaged.tree,
	})
}
