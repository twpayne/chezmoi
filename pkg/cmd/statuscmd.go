package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type statusCmdConfig struct {
	exclude   *chezmoi.EntryTypeSet
	include   *chezmoi.EntryTypeSet
	init      bool
	recursive bool
}

func (c *Config) newStatusCmd() *cobra.Command {
	statusCmd := &cobra.Command{
		Use:               "status [target]...",
		Short:             "Show the status of targets",
		Long:              mustLongHelp("status"),
		Example:           example("status"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.makeRunEWithSourceState(c.runStatusCmd),
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			persistentStateMode:          persistentStateModeReadMockWrite,
			requiresSourceDirectory:      "true",
		},
	}

	flags := statusCmd.Flags()
	flags.VarP(c.status.exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.status.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.status.init, "init", c.update.init, "Recreate config file from template")
	flags.BoolVarP(&c.status.recursive, "recursive", "r", c.status.recursive, "Recurse into subdirectories")

	return statusCmd
}

func (c *Config) runStatusCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	builder := strings.Builder{}
	dryRunSystem := chezmoi.NewDryRunSystem(c.destSystem)
	preApplyFunc := func(
		targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState,
	) error {
		c.logger.Info().
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
			fmt.Fprintf(&builder, "%c%c %s\n", x, y, targetRelPath)
		}
		return chezmoi.Skip
	}
	if err := c.applyArgs(cmd.Context(), dryRunSystem, c.DestDirAbsPath, args, applyArgsOptions{
		include:      c.status.include.Sub(c.status.exclude),
		init:         c.status.init,
		recursive:    c.status.recursive,
		umask:        c.Umask,
		preApplyFunc: preApplyFunc,
	}); err != nil {
		return err
	}
	return c.writeOutputString(builder.String())
}

func statusRune(fromState, toState *chezmoi.EntryState) rune {
	if fromState == nil || fromState.Equivalent(toState) {
		return ' '
	}
	switch toState.Type {
	case chezmoi.EntryStateTypeRemove:
		return 'D'
	case chezmoi.EntryStateTypeDir, chezmoi.EntryStateTypeFile, chezmoi.EntryStateTypeSymlink:
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
