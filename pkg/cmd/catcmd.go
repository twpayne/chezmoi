package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newCatCmd() *cobra.Command {
	catCmd := &cobra.Command{
		Use:               "cat target...",
		Short:             "Print the target contents of a file, script, or symlink",
		Long:              mustLongHelp("cat"),
		Example:           example("cat"),
		ValidArgsFunction: c.targetValidArgs,
		Args:              cobra.MinimumNArgs(1),
		RunE:              c.makeRunEWithSourceState(c.runCatCmd),
		Annotations: newAnnotations(
			requiresSourceDirectory,
		),
	}

	return catCmd
}

func (c *Config) runCatCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeManaged: true,
	})
	if err != nil {
		return err
	}

	builder := strings.Builder{}
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
		if err != nil {
			return fmt.Errorf("%s: %w", targetRelPath, err)
		}
		switch targetStateEntry := targetStateEntry.(type) {
		case *chezmoi.TargetStateFile:
			contents, err := targetStateEntry.Contents()
			if err != nil {
				return fmt.Errorf("%s: %w", targetRelPath, err)
			}
			builder.Write(contents)
		case *chezmoi.TargetStateScript:
			contents, err := targetStateEntry.Contents()
			if err != nil {
				return fmt.Errorf("%s: %w", targetRelPath, err)
			}
			builder.Write(contents)
		case *chezmoi.TargetStateSymlink:
			linkname, err := targetStateEntry.Linkname()
			if err != nil {
				return fmt.Errorf("%s: %w", targetRelPath, err)
			}
			builder.WriteString(linkname)
			builder.WriteByte('\n')
		default:
			return fmt.Errorf("%s: not a file, script, or symlink", targetRelPath)
		}
	}
	return c.writeOutputString(builder.String())
}
