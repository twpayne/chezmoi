package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type statusCmdConfig struct {
	Exclude    *chezmoi.EntryTypeSet `json:"exclude"   mapstructure:"exclude"   yaml:"exclude"`
	PathStyle  *chezmoi.PathStyle    `json:"pathStyle" mapstructure:"pathStyle" yaml:"pathStyle"`
	include    *chezmoi.EntryTypeSet
	init       bool
	parentDirs bool
	recursive  bool
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
			persistentStateModeReadMockWrite,
			requiresSourceDirectory,
		),
	}

	statusCmd.Flags().VarP(c.Status.Exclude, "exclude", "x", "Exclude entry types")
	statusCmd.Flags().VarP(c.Status.PathStyle, "path-style", "p", "Path style")
	statusCmd.Flags().VarP(c.Status.include, "include", "i", "Include entry types")
	statusCmd.Flags().BoolVar(&c.Status.init, "init", c.Status.init, "Recreate config file from template")
	statusCmd.Flags().BoolVarP(&c.Status.parentDirs, "parent-dirs", "P", c.Status.parentDirs, "Show status of all parent directories")
	statusCmd.Flags().BoolVarP(&c.Status.recursive, "recursive", "r", c.Status.recursive, "Recurse into subdirectories")

	return statusCmd
}

func (c *Config) runStatusCmd(cmd *cobra.Command, args []string) error {
	builder := strings.Builder{}
	preApplyFunc := func(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
		c.logger.Info("statusPreApplyFunc",
			chezmoilog.Stringer("targetRelPath", targetRelPath),
			slog.Any("targetEntryState", targetEntryState),
			slog.Any("lastWrittenEntryState", lastWrittenEntryState),
			slog.Any("actualEntryState", actualEntryState),
		)

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
				return errors.New("source-absolute not supported for status")
			case chezmoi.PathStyleSourceRelative:
				return errors.New("source-relative not supported for status")
			}

			fmt.Fprintf(&builder, "%c%c %s\n", x, y, path)
		}
		return fs.SkipDir
	}
	if err := c.applyArgs(cmd.Context(), c.destSystem, c.DestDirAbsPath, args, applyArgsOptions{
		cmd:          cmd,
		filter:       chezmoi.NewEntryTypeFilter(c.Status.include.Bits(), c.Status.Exclude.Bits()),
		init:         c.Status.init,
		parentDirs:   c.Status.parentDirs,
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
