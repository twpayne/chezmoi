package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type managedCmdConfig struct {
	exclude *chezmoi.EntryTypeSet
	include *chezmoi.EntryTypeSet
}

func (c *Config) newManagedCmd() *cobra.Command {
	managedCmd := &cobra.Command{
		Use:     "managed [paths]...",
		Aliases: []string{"list"},
		Short:   "List the managed entries in the destination directory",
		Long:    mustLongHelp("managed"),
		Example: example("managed"),
		Args:    cobra.ArbitraryArgs,
		RunE:    c.makeRunEWithSourceState(c.runManagedCmd),
	}

	flags := managedCmd.Flags()
	flags.VarP(c.managed.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.managed.include, "include", "i", "Include entry types")

	return managedCmd
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	include := c.managed.include.Sub(c.managed.exclude)

	// Build queued paths. When no arguments, start from root; otherwise start
	// from arguments.
	paths := []chezmoi.RelPath{}
	if len(args) != 0 {
		for _, arg := range args {
			if p, err := chezmoi.NormalizePath(arg); err != nil {
				return err
			} else if p, err := p.TrimDirPrefix(c.DestDirAbsPath); err != nil {
				return err
			} else {
				paths = append(paths, p)
			}
		}
	}

	var targetRelPaths chezmoi.RelPaths
	_ = sourceState.ForEach(func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
		if err != nil {
			return err
		}
		if !include.IncludeTargetStateEntry(targetStateEntry) {
			return nil
		}

		// when specified arguments, only include paths under these arguments
		if len(paths) != 0 {
			included := false
			for _, path := range paths {
				if targetRelPath.HasDirPrefix(path) || targetRelPath.String() == path.String() {
					included = true
					break
				}
			}
			if !included {
				return nil
			}
		}

		targetRelPaths = append(targetRelPaths, targetRelPath)
		return nil
	})

	sort.Sort(targetRelPaths)
	builder := strings.Builder{}
	for _, targetRelPath := range targetRelPaths {
		fmt.Fprintln(&builder, targetRelPath)
	}
	return c.writeOutputString(builder.String())
}
