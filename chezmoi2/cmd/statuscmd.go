package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

type statusCmdConfig struct {
	include   *chezmoi.IncludeSet
	recursive bool
}

func (c *Config) newStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:     "status [target]...",
		Short:   "Show the status of targets",
		Long:    mustLongHelp("status"),
		Example: example("status"),
		RunE:    c.makeRunEWithSourceState(c.runStatusCmd),
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadMockWrite,
		},
	}

	flags := statusCmd.Flags()
	flags.VarP(c.status.include, "include", "i", "include entry types")
	flags.BoolVarP(&c.status.recursive, "recursive", "r", c.status.recursive, "recursive")

	return statusCmd
}

func (c *Config) runStatusCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	preApplyFunc := func(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
		var (
			x = ' '
			y = ' '
		)
		switch {
		case targetEntryState.Type == chezmoi.EntryStateTypeScript:
			y = 'R'
		case !targetEntryState.Equivalent(actualEntryState):
			x = statusRune(lastWrittenEntryState, actualEntryState)
			y = statusRune(actualEntryState, targetEntryState)
		}
		if x != ' ' || y != ' ' {
			fmt.Fprintf(&sb, "%c%c %s\n", x, y, targetRelPath)
		}
		return chezmoi.Skip
	}
	if err := c.applyArgs(dryRunSystem, c.destDirAbsPath, args, applyArgsOptions{
		include:      c.status.include,
		recursive:    c.status.recursive,
		umask:        c.Umask,
		preApplyFunc: preApplyFunc,
	}); err != nil {
		return err
	}
	return c.writeOutputString(sb.String())
}

func statusRune(fromState, toState *chezmoi.EntryState) rune {
	if fromState == nil || fromState.Equivalent(toState) {
		return ' '
	}
	switch toState.Type {
	case chezmoi.EntryStateTypeAbsent:
		return 'D'
	case chezmoi.EntryStateTypeDir, chezmoi.EntryStateTypeFile, chezmoi.EntryStateTypeSymlink:
		//nolint:exhaustive
		switch fromState.Type {
		case chezmoi.EntryStateTypeAbsent:
			return 'A'
		default:
			return 'M'
		}
	case chezmoi.EntryStateTypePresent:
		return 'A'
	case chezmoi.EntryStateTypeScript:
		return 'R'
	default:
		return '?'
	}
}
