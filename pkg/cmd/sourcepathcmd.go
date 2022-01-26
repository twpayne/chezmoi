package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newSourcePathCmd() *cobra.Command {
	sourcePathCmd := &cobra.Command{
		Use:               "source-path [target]...",
		Short:             "Print the path of a target in the source state",
		Long:              mustLongHelp("source-path"),
		Example:           example("source-path"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.makeRunEWithSourceState(c.runSourcePathCmd),
	}

	return sourcePathCmd
}

func (c *Config) runSourcePathCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	if len(args) == 0 {
		return c.writeOutputString(c.SourceDirAbsPath.String() + "\n")
	}

	sourceAbsPaths, err := c.sourceAbsPaths(sourceState, args)
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	for _, sourceAbsPath := range sourceAbsPaths {
		fmt.Fprintln(&builder, sourceAbsPath)
	}
	return c.writeOutputString(builder.String())
}
