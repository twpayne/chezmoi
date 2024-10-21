package cmd

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type destroyCmdConfig struct {
	recursive bool
}

func (c *Config) newDestroyCmd() *cobra.Command {
	destroyCmd := &cobra.Command{
		Use:               "destroy target...",
		Short:             "Permanently delete an entry from the source state, the destination directory, and the state",
		Long:              mustLongHelp("destroy"),
		Example:           example("destroy"),
		ValidArgsFunction: c.targetValidArgs,
		Args:              cobra.MinimumNArgs(1),
		RunE:              c.makeRunEWithSourceState(c.runDestroyCmd),
		Annotations: newAnnotations(
			modifiesDestinationDirectory,
			modifiesSourceDirectory,
			persistentStateModeReadWrite,
		),
	}

	destroyCmd.Flags().BoolVarP(&c.destroy.recursive, "recursive", "r", c.destroy.recursive, "Recurse into subdirectories")

	return destroyCmd
}

func (c *Config) runDestroyCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		recursive: c.destroy.recursive,
	})
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		destAbsPath := c.DestDirAbsPath.Join(targetRelPath)
		// Find the path of the entry in the source state, if any.
		//
		// chezmoi destroy might be called on an entry in an exact_ directory
		// that is not present in the source directory. The entry is still
		// managed by chezmoi because chezmoi apply will remove it. Therefore,
		// chezmoi destroy should remove such entries from the target state,
		// even if they are not present in the source state. So, when calling
		// chezmoi destroy on entries like this, we should only remove the entry
		// from the target state, not the source state.
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
				prompt = fmt.Sprintf("Destroy %s", destAbsPath)
			} else {
				prompt = fmt.Sprintf("Destroy %s and %s", destAbsPath, sourceAbsPath)
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
