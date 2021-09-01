package cmd

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type statusCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
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
	flags.VarP(c.status.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.status.include, "include", "i", "Include entry types")
	flags.BoolVarP(&c.status.recursive, "recursive", "r", c.status.recursive, "Recurse into subdirectories")

	return statusCmd
}

func (c *Config) runStatusCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sb := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	statusCmdPreApplyFunc := func(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
		log.Debug().
			Stringer("targetRelPath", targetRelPath).
			Object("targetEntryState", targetEntryState).
			Object("lastWrittenEntryState", lastWrittenEntryState).
			Object("actualEntryState", actualEntryState).
			Msg("statusPreApplyFunc")

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
	if err := c.applyArgs(cmd.Context(), dryRunSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:      c.status.include.Sub(c.status.exclude),
		recursive:    c.status.recursive,
		umask:        c.Umask,
		preApplyFunc: statusCmdPreApplyFunc,
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
	case chezmoi.EntryStateTypeRemove:
		return 'D'
	case chezmoi.EntryStateTypeDir, chezmoi.EntryStateTypeFile, chezmoi.EntryStateTypeSymlink:
		//nolint:exhaustive
		switch fromState.Type {
		case chezmoi.EntryStateTypeRemove:
			return 'A'
		default:
			return 'M'
		}
	case chezmoi.EntryStateTypeScript:
		return 'R'
	default:
		return '?'
	}
}
