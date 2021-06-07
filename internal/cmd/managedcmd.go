package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type managedCmdConfig struct {
	exclude *chezmoi.EntryTypeSet
	include *chezmoi.EntryTypeSet
}

func (c *Config) newManagedCmd() *cobra.Command {
	managedCmd := &cobra.Command{
		Use:     "managed",
		Short:   "List the managed entries in the destination directory",
		Long:    mustLongHelp("managed"),
		Example: example("managed"),
		Args:    cobra.NoArgs,
		RunE:    c.makeRunEWithSourceState(c.runManagedCmd),
	}

	flags := managedCmd.Flags()
	flags.VarP(c.managed.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.managed.include, "include", "i", "Include entry types")

	return managedCmd
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	include := c.managed.include.Sub(c.managed.exclude)
	entries := sourceState.Entries()
	targetRelPaths := make([]chezmoi.RelPath, 0, len(entries))
	for targetRelPath, sourceStateEntry := range entries {
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
		if err != nil {
			return err
		}
		if !include.IncludeTargetStateEntry(targetStateEntry) {
			continue
		}
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}

	sort.Slice(targetRelPaths, func(i, j int) bool {
		return targetRelPaths[i] < targetRelPaths[j]
	})
	sb := strings.Builder{}
	for _, targetRelPath := range targetRelPaths {
		fmt.Fprintln(&sb, targetRelPath)
	}
	return c.writeOutputString(sb.String())
}
