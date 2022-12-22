package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type managedCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	pathStyle pathStyle
}

func (c *Config) newManagedCmd() *cobra.Command {
	managedCmd := &cobra.Command{
		Use:     "managed [path]...",
		Aliases: []string{"list"},
		Short:   "List the managed entries in the destination directory",
		Long:    mustLongHelp("managed"),
		Example: example("managed"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runManagedCmd),
	}

	flags := managedCmd.Flags()
	flags.VarP(c.managed.filter.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.managed.filter.Include, "include", "i", "Include entry types")
	flags.VarP(&c.managed.pathStyle, "path-style", "p", "Path style")

	registerExcludeIncludeFlagCompletionFuncs(managedCmd)
	if err := managedCmd.RegisterFlagCompletionFunc("path-style", pathStyleFlagCompletionFunc); err != nil {
		panic(err)
	}

	return managedCmd
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// Build queued relPaths. When there are no arguments, start from root,
	// otherwise start from arguments.
	var relPaths chezmoi.RelPaths
	for _, arg := range args {
		if absPath, err := chezmoi.NormalizePath(arg); err != nil {
			return err
		} else if relPath, err := absPath.TrimDirPrefix(c.DestDirAbsPath); err != nil {
			return err
		} else {
			relPaths = append(relPaths, relPath)
		}
	}

	var paths []string
	_ = sourceState.ForEach(func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
		if !c.managed.filter.IncludeSourceStateEntry(sourceStateEntry) {
			return nil
		}

		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
		if err != nil {
			return err
		}
		if !c.managed.filter.IncludeTargetStateEntry(targetStateEntry) {
			return nil
		}

		// When arguments are given, only include paths under these arguments.
		if len(relPaths) != 0 {
			included := false
			for _, path := range relPaths {
				if targetRelPath.HasDirPrefix(path) || targetRelPath.String() == path.String() {
					included = true
					break
				}
			}
			if !included {
				return nil
			}
		}

		var path string
		switch c.managed.pathStyle {
		case pathStyleAbsolute:
			path = c.DestDirAbsPath.Join(targetRelPath).String()
		case pathStyleRelative:
			path = targetRelPath.String()
		case pathStyleSourceAbsolute:
			path = c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath()).String()
		case pathStyleSourceRelative:
			path = sourceStateEntry.SourceRelPath().RelPath().String()
		}
		paths = append(paths, path)
		return nil
	})

	sort.Strings(paths)
	builder := strings.Builder{}
	for _, path := range paths {
		fmt.Fprintln(&builder, path)
	}
	return c.writeOutputString(builder.String())
}
