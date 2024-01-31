package cmd

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoimaps"
)

type unmanagedCmdConfig struct {
	pathStyle chezmoi.PathStyle
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

	flags := unmanagedCmd.Flags()
	flags.VarP(&c.unmanaged.pathStyle, "path-style", "p", "Path style")

	if err := unmanagedCmd.RegisterFlagCompletionFunc("path-style", chezmoi.PathStyleFlagCompletionFunc); err != nil {
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

	builder := strings.Builder{}
	sortedRelPaths := chezmoi.RelPaths(chezmoimaps.Keys(unmanagedRelPaths))
	sort.Sort(sortedRelPaths)
	for _, relPath := range sortedRelPaths {
		switch c.unmanaged.pathStyle {
		case chezmoi.PathStyleAbsolute:
			fmt.Fprintln(&builder, c.DestDirAbsPath.Join(relPath))
		case chezmoi.PathStyleRelative:
			fmt.Fprintln(&builder, relPath)
		}
	}
	return c.writeOutputString(builder.String())
}
