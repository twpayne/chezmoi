package cmd

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v4"
	"golang.org/x/exp/maps"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type unmanagedCmdConfig struct {
	pathStyle pathStyle
}

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		Use:     "unmanaged [path]...",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
	}

	flags := unmanagedCmd.Flags()
	flags.VarP(&c.unmanaged.pathStyle, "path-style", "p", "Path style")

	if err := unmanagedCmd.RegisterFlagCompletionFunc("path-style", pathStyleFlagCompletionFunc); err != nil {
		panic(err)
	}

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	var absPaths chezmoi.AbsPaths
	if len(args) == 0 {
		absPaths = append(absPaths, c.DestDirAbsPath)
	} else {
		argsAbsPaths := make(map[chezmoi.AbsPath]struct{})
		for _, arg := range args {
			argAbsPath, err := chezmoi.NormalizePath(arg)
			if err != nil {
				return err
			}
			argsAbsPaths[argAbsPath] = struct{}{}
		}
		for argAbsPath := range argsAbsPaths {
			absPaths = append(absPaths, argAbsPath)
		}
		sort.Sort(absPaths)
	}

	unmanagedRelPaths := make(map[chezmoi.RelPath]struct{})
	walkFunc := func(destAbsPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
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
			unmanagedRelPaths[targetRelPath] = struct{}{}
		}
		if fileInfo.IsDir() {
			switch {
			case !managed:
				return vfs.SkipDir
			case ignored:
				return vfs.SkipDir
			case sourceStateEntry != nil:
				if origin, ok := sourceStateEntry.Origin().(*chezmoi.External); ok {
					if origin.Type == chezmoi.ExternalTypeGitRepo {
						return vfs.SkipDir
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

	builder := strings.Builder{}
	sortedRelPaths := chezmoi.RelPaths(maps.Keys(unmanagedRelPaths))
	sort.Sort(sortedRelPaths)
	for _, relPath := range sortedRelPaths {
		switch c.unmanaged.pathStyle {
		case pathStyleAbsolute:
			fmt.Fprintln(&builder, c.DestDirAbsPath.Join(relPath))
		case pathStyleRelative:
			fmt.Fprintln(&builder, relPath)
		}
	}
	return c.writeOutputString(builder.String())
}
