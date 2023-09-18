package cmd

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type statusCmdConfig struct {
	Exclude   *chezmoi.EntryTypeSet `json:"exclude"   mapstructure:"exclude"   yaml:"exclude"`
	PathStyle *chezmoi.PathStyle    `json:"pathStyle" mapstructure:"pathStyle" yaml:"pathStyle"`
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
		RunE:              c.runStatusCmd,
		Annotations: newAnnotations(
			dryRun,
			modifiesDestinationDirectory,
			persistentStateModeReadMockWrite,
			requiresSourceDirectory,
		),
	}

	flags := statusCmd.Flags()
	flags.VarP(c.Status.Exclude, "exclude", "x", "Exclude entry types")
	flags.VarP(c.Status.PathStyle, "path-style", "p", "Path style")
	flags.VarP(c.Status.include, "include", "i", "Include entry types")
	flags.BoolVar(&c.Status.init, "init", c.Status.init, "Recreate config file from template")
	flags.BoolVarP(
		&c.Status.recursive,
		"recursive",
		"r",
		c.Status.recursive,
		"Recurse into subdirectories",
	)

	registerExcludeIncludeFlagCompletionFuncs(statusCmd)

	return statusCmd
}

func (c *Config) runStatusCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
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
			var path string
			switch *c.Status.PathStyle {
			case chezmoi.PathStyleAbsolute:
				path = c.DestDirAbsPath.Join(targetRelPath).String()
			case chezmoi.PathStyleRelative:
				path = targetRelPath.String()
			case chezmoi.PathStyleSourceAbsolute:
				return fmt.Errorf("source-absolute not supported for status")
			case chezmoi.PathStyleSourceRelative:
				return fmt.Errorf("source-relative not supported for status")
			}

			fmt.Fprintf(&builder, "%c%c %s\n", x, y, path)
		}
		return fs.SkipDir
	}
	if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:          cmd,
		filter:       chezmoi.NewEntryTypeFilter(c.Status.include.Bits(), c.Status.Exclude.Bits()),
		init:         c.Status.init,
		recursive:    c.Status.recursive,
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
