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
		Annotations: newAnnotations(
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}

	return forgetCmd
}

func (c *Config) runForgetCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeManaged: true,
	})
	if err != nil {
		return err
	}

TARGET_REL_PATH:
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntry := sourceState.MustEntry(targetRelPath)

		// Skip source state entries that are not regular entries. These are
		// removes or externals, which we cannot handle.
		switch sourceStateOrigin := sourceStateEntry.Origin(); sourceStateOrigin.(type) {
		case chezmoi.SourceStateOriginAbsPath:
			// OK, keep going.
		case chezmoi.SourceStateOriginRemove:
			c.errorf("warning: %s: cannot forget entry from remove\n", targetRelPath)
			continue TARGET_REL_PATH
		case *chezmoi.External:
			c.errorf("warning: %s: cannot forget entry from external %s\n", targetRelPath, sourceStateOrigin.OriginString())
			continue TARGET_REL_PATH
		default:
			panic(fmt.Sprintf("%s: %T: unknown source state origin type", targetRelPath, sourceStateOrigin))
		}

		sourceAbsPath := c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath())
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
