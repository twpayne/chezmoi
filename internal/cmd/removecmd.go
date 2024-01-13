package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
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
		mustBeManaged: true,
		recursive:     c.remove.recursive,
	})
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		destAbsPath := c.DestDirAbsPath.Join(targetRelPath)
		// Find the path of the entry in the source state, if any.
		//
		// chezmoi remove might be called on an entry in an exact_ directory
		// that is not present in the source directory. The entry is still
		// managed by chezmoi because chezmoi apply will remove it. Therefore,
		// chezmoi remove should remove such entries from the target state, even
		// if they are not present in the source state. So, when calling chezmoi
		// remove on entries like this, we should only remove the entry from the
		// target state, not the source state.
		//
		// For entries in exact_ directories in the target state that are not
		// present in the source state, we generate SourceStateRemove entries.
		// So, if the source state entry is a SourceStateRemove then we know
		// that there is no actual source state entry to remove.
		var sourceAbsPath chezmoi.AbsPath
		sourceStateEntry := sourceState.MustEntry(targetRelPath)
		if _, ok := sourceStateEntry.(*chezmoi.SourceStateRemove); !ok {
			sourceAbsPath = c.SourceDirAbsPath.Join(sourceStateEntry.SourceRelPath().RelPath())
		}
		if !c.force {
			var prompt string
			if sourceAbsPath.Empty() {
				prompt = fmt.Sprintf("Remove %s", destAbsPath)
			} else {
				prompt = fmt.Sprintf("Remove %s and %s", destAbsPath, sourceAbsPath)
			}
			choice, err := c.promptChoice(prompt, choicesYesNoAllQuit)
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
		if !sourceAbsPath.Empty() {
			if err := c.sourceSystem.RemoveAll(sourceAbsPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		}
		if err := c.persistentState.Delete(chezmoi.EntryStateBucket, destAbsPath.Bytes()); err != nil {
			return err
		}
	}
	return nil
}
