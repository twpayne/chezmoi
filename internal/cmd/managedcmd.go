package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type managedCmdConfig struct {
	filter    *chezmoi.EntryTypeFilter
	pathStyle chezmoi.PathStyle
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
			persistentStateModeReadOnly,
		),
	}

	managedCmd.Flags().VarP(c.managed.filter.Exclude, "exclude", "x", "Exclude entry types")
	managedCmd.Flags().VarP(c.managed.filter.Include, "include", "i", "Include entry types")
	managedCmd.Flags().VarP(&c.managed.pathStyle, "path-style", "p", "Path style")
	managedCmd.Flags().BoolVarP(&c.managed.tree, "tree", "t", c.managed.tree, "Print paths as a tree")

	if err := managedCmd.RegisterFlagCompletionFunc("path-style", chezmoi.PathStyleFlagCompletionFunc); err != nil {
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
		} else { //nolint:revive
			relPaths = append(relPaths, relPath)
		}
	}

	var paths []fmt.Stringer
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

			var path fmt.Stringer
			switch c.managed.pathStyle {
			case chezmoi.PathStyleAbsolute:
				path = c.DestDirAbsPath.Join(targetRelPath)
			case chezmoi.PathStyleRelative:
				path = targetRelPath
			case chezmoi.PathStyleSourceAbsolute:
				path = c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath())
			case chezmoi.PathStyleSourceRelative:
				path = sourceStateEntry.SourceRelPath().RelPath()
			}
			paths = append(paths, path)
			return nil
		},
	)

	return c.writePaths(stringersToStrings(paths), writePathsOptions{
		tree: c.managed.tree,
	})
}
