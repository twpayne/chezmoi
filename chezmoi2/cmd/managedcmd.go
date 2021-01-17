package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type managedCmdConfig struct {
	include *chezmoi.IncludeSet
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
	flags.VarP(c.managed.include, "include", "i", "include entry types")

	return managedCmd
}

func (c *Config) runManagedCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	entries := sourceState.Entries()
	targetRelPaths := make(chezmoi.RelPaths, 0, len(entries))
	for targetRelPath, sourceStateEntry := range entries {
		targetStateEntry, err := sourceStateEntry.TargetStateEntry()
		if err != nil {
			return err
		}
		if !c.managed.include.IncludeTargetStateEntry(targetStateEntry) {
			continue
		}
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}

	sort.Sort(targetRelPaths)
	sb := strings.Builder{}
	for _, targetRelPath := range targetRelPaths {
		fmt.Fprintln(&sb, targetRelPath)
	}
	return c.writeOutputString(sb.String())
}
