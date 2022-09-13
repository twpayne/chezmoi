package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func (c *Config) newSourcePathCmd() *cobra.Command {
	sourcePathCmd := &cobra.Command{
		Use:               "source-path [target]...",
		Short:             "Print the source path of a target",
		Long:              mustLongHelp("source-path"),
		Example:           example("source-path"),
		ValidArgsFunction: c.targetValidArgs,
		RunE:              c.runSourcePathCmd,
	}

	return sourcePathCmd
}

func (c *Config) runSourcePathCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return c.writeOutputString(c.SourceDirAbsPath.String() + "\n")
	}

	sourceState, err := c.newSourceState(cmd.Context())
	if err != nil {
		return err
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
