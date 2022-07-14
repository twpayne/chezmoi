package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

func (c *Config) newForgetCmd() *cobra.Command {
	forgetCmd := &cobra.Command{
		Use:               "forget target...",
		Aliases:           []string{"unmanage"},
		Short:             "Remove a target from the source state",
		Long:              mustLongHelp("forget"),
		Example:           example("forget"),
		ValidArgsFunction: c.targetValidArgs,
		Args:              cobra.MinimumNArgs(1),
		RunE:              c.makeRunEWithSourceState(c.runForgetCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
			persistentStateMode:     persistentStateModeReadWrite,
		},
	}

	return forgetCmd
}

func (c *Config) runForgetCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		// Filter out source state entries with an empty source relative path.
		// These are generated by externals, which we cannot handle.
		relPath := sourceState.MustEntry(targetRelPath).SourceRelPath().RelPath()
		if relPath.Empty() {
			c.errorf("warning: %s: cannot forget entry, is it an external?", targetRelPath)
			continue
		}

		sourceAbsPath := c.SourceDirAbsPath.Join(relPath)
		if !c.force {
			choice, err := c.promptChoice(fmt.Sprintf("Remove %s", sourceAbsPath), choicesYesNoAllQuit)
			if err != nil {
				return err
			}
			switch choice {
			case "yes":
			case "no":
				continue
			case "all":
				c.force = false
			case "quit":
				return nil
			}
		}
		if err := c.sourceSystem.RemoveAll(sourceAbsPath); err != nil {
			return err
		}

		targetAbsPath := c.DestDirAbsPath.Join(targetRelPath)
		if err := c.persistentState.Delete(chezmoi.EntryStateBucket, targetAbsPath.Bytes()); err != nil {
			return err
		}
	}

	return nil
}