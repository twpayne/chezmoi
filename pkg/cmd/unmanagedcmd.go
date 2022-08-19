package cmd

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v4"
	"golang.org/x/exp/maps"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		Use:     "unmanaged [path]...",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
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
		managed := sourceState.Contains(targetRelPath)
		ignored := sourceState.Ignore(targetRelPath)
		if !managed && !ignored {
			unmanagedRelPaths[targetRelPath] = struct{}{}
		}
		if fileInfo.IsDir() && (!managed || ignored) {
			return vfs.SkipDir
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
		builder.WriteString(relPath.String())
		builder.WriteByte('\n')
	}
	return c.writeOutputString(builder.String())
}
