package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type removeCmdConfig struct {
	recursive bool
}

func (c *Config) newRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:               "remove target...",
		Aliases:           []string{"rm"},
		Short:             "Remove a target from the source state and the destination directory",
		Long:              mustLongHelp("remove"),
		Example:           example("remove"),
		ValidArgsFunction: c.targetValidArgs,
		Args:              cobra.MinimumNArgs(1),
		RunE:              c.makeRunEWithSourceState(c.runRemoveCmd),
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}

	flags := removeCmd.Flags()
	flags.BoolVarP(&c.remove.recursive, "recursive", "r", c.remove.recursive, "Recurse into subdirectories")

	return removeCmd
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
		recursive:           c.remove.recursive,
	})
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		destAbsPath := c.DestDirAbsPath.Join(targetRelPath)
		sourceAbsPath := c.SourceDirAbsPath.Join(sourceState.MustEntry(targetRelPath).SourceRelPath().RelPath())
		if !c.force {
			choice, err := c.promptChoice(fmt.Sprintf("Remove %s and %s", destAbsPath, sourceAbsPath), choicesYesNoAllQuit)
			if err != nil {
				return err
			}
			switch choice {
			case "yes":
			case "no":
				continue
			case "all":
				c.force = true
			case "quit":
				return nil
			}
		}
		if err := c.destSystem.RemoveAll(destAbsPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := c.sourceSystem.RemoveAll(sourceAbsPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		if err := c.persistentState.Delete(chezmoi.EntryStateBucket, destAbsPath.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
