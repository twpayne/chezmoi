package cmd

import (
	"cmp"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type managedCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	format    *choiceFlag
	pathStyle *choiceFlag
	tree      bool
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
		Annotations: newAnnotations(
			persistentStateModeReadWrite,
		),
	}

	managedCmd.Flags().VarP(c.managed.filter.Exclude, "exclude", "x", "Exclude entry types")
	managedCmd.Flags().VarP(c.managed.format, "format", "f", "Format")
	managedCmd.Flags().VarP(c.managed.filter.Include, "include", "i", "Include entry types")
	managedCmd.Flags().VarP(c.managed.pathStyle, "path-style", "p", "Path style")
	must(managedCmd.RegisterFlagCompletionFunc("path-style", c.managed.pathStyle.FlagCompletionFunc()))
	managedCmd.Flags().BoolVarP(&c.managed.tree, "tree", "t", c.managed.tree, "Print paths as a tree")

	return managedCmd
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	// Build queued relPaths. When there are no arguments, start from root,
	// otherwise start from arguments.
	var relPaths []chezmoi.RelPath
	for _, arg := range args {
		if absPath, err := chezmoi.NormalizePath(arg); err != nil {
			return err
		} else if relPath, err := absPath.TrimDirPrefix(c.DestDirAbsPath); err != nil {
			return err
		} else { //nolint:revive
			relPaths = append(relPaths, relPath)
		}
	}

	type entryPaths struct {
		targetRelPath  chezmoi.RelPath
		Absolute       chezmoi.AbsPath       `json:"absolute"       yaml:"absolute"`
		SourceAbsolute chezmoi.AbsPath       `json:"sourceAbsolute" yaml:"sourceAbsolute"`
		SourceRelative chezmoi.SourceRelPath `json:"sourceRelative" yaml:"sourceRelative"`
	}
	var allEntryPaths []*entryPaths
	_ = sourceState.ForEach(
		func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
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

			entryPaths := &entryPaths{
				targetRelPath:  targetRelPath,
				Absolute:       c.DestDirAbsPath.Join(targetRelPath),
				SourceAbsolute: c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath()),
				SourceRelative: sourceStateEntry.SourceRelPath(),
			}
			allEntryPaths = append(allEntryPaths, entryPaths)
			return nil
		},
	)

	switch pathStyle := c.managed.pathStyle.String(); pathStyle {
	case pathStyleAbsolute, pathStyleRelative, pathStyleSourceAbsolute, pathStyleSourceRelative:
		paths := make([]string, len(allEntryPaths))
		for i, structuredPath := range allEntryPaths {
			switch c.managed.pathStyle.String() {
			case pathStyleAbsolute:
				paths[i] = structuredPath.Absolute.String()
			case pathStyleRelative:
				paths[i] = structuredPath.targetRelPath.String()
			case pathStyleSourceAbsolute:
				paths[i] = structuredPath.SourceAbsolute.String()
			case pathStyleSourceRelative:
				paths[i] = structuredPath.SourceRelative.String()
			}
		}
		return c.writePaths(paths, writePathsOptions{
			tree: c.managed.tree,
		})
	case pathStyleAll:
		allEntryPathsMap := make(map[string]*entryPaths, len(allEntryPaths))
		for _, entryPaths := range allEntryPaths {
			allEntryPathsMap[entryPaths.targetRelPath.String()] = entryPaths
		}
		return c.marshal(cmp.Or(c.managed.format.String(), c.Format.String()), allEntryPathsMap)
	default:
		return fmt.Errorf("%s: invalid path style", pathStyle)
	}
}
