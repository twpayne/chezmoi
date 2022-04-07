package cmd

import (
	"io/fs"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newUnmanagedCmd() *cobra.Command {
	unmanagedCmd := &cobra.Command{
		Use:     "unmanaged [paths]...",
		Short:   "List the unmanaged files in the destination directory",
		Long:    mustLongHelp("unmanaged"),
		Example: example("unmanaged"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runUnmanagedCmd),
	}

	return unmanagedCmd
}

func (c *Config) runUnmanagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// the set of discovered, unmanaged items
	unmanaged := map[string]bool{}
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
			unmanaged[targetRelPath.String()] = true
		}
		if fileInfo.IsDir() && (!managed || ignored) {
			return vfs.SkipDir
		}
		return nil
	}

	// Build queued paths. When no arguments, start from root; otherwise start
	// from arguments.	The paths are deduplicated and sorted.
	paths := make([]chezmoi.AbsPath, 0, len(args)) // (lsttype, size, capacity)
	if len(args) == 0 {
		paths = append(paths, c.DestDirAbsPath)
	} else {
		qPaths := make(map[chezmoi.AbsPath]bool, len(args)) // (map, capacity)
		for _, arg := range args {
			p, err := chezmoi.NormalizePath(arg)
			if err != nil {
				return err
			}
			qPaths[p] = true
		}
		for path := range qPaths {
			paths = append(paths, path)
		}
		sort.Slice(paths,
			func(i, j int) bool { return paths[i].Less(paths[j]) })
	}

	for _, path := range paths {
		if err := chezmoi.Walk(c.destSystem, path, walkFunc); err != nil {
			return err
		}
	}

	// collect the keys and sort
	builder := strings.Builder{}
	unmPaths := make([]string, 0, len(unmanaged))
	for path := range unmanaged {
		unmPaths = append(unmPaths, path)
	}
	sort.Strings(unmPaths)

	for _, path := range unmPaths {
		builder.WriteString(path)
		builder.WriteByte('\n')
	}
	return c.writeOutputString(builder.String())
}
