package chezmoicmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

func (c *Config) newSourcePathCmd() *cobra.Command {
	sourcePathCmd := &cobra.Command{
		Use:     "source-path [target]...",
		Short:   "Print the path of a target in the source state",
		Long:    mustLongHelp("source-path"),
		Example: example("source-path"),
		RunE:    c.makeRunEWithSourceState(c.runSourcePathCmd),
	}

	return sourcePathCmd
}

func (c *Config) runSourcePathCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	if len(args) == 0 {
		return c.writeOutputString(string(c.SourceDirAbsPath) + "\n")
	}

	sourceAbsPaths, err := c.sourceAbsPaths(sourceState, args)
	if err != nil {
		return err
	}

	sb := strings.Builder{}
	for _, sourceAbsPath := range sourceAbsPaths {
		fmt.Fprintln(&sb, sourceAbsPath)
	}
	return c.writeOutputString(sb.String())
}
