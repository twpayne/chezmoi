package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

func (c *Config) newRemoveCmd() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:     "remove target...",
		Aliases: []string{"rm"},
		Short:   "Remove a target from the source state and the destination directory",
		Long:    mustLongHelp("remove"),
		Example: example("remove"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runRemoveCmd),
		Annotations: map[string]string{
			modifiesDestinationDirectory: "true",
			modifiesSourceDirectory:      "true",
		},
	}

	return removeCmd
}

func (c *Config) runRemoveCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return err
	}

	for _, targetRelPath := range targetRelPaths {
		destAbsPath := c.destDirAbsPath.Join(targetRelPath)
		sourceAbsPath := c.sourceDirAbsPath.Join(sourceState.MustEntry(targetRelPath).SourceRelPath().RelPath())
		if !c.force {
			choice, err := c.promptValue(fmt.Sprintf("Remove %s and %s", destAbsPath, sourceAbsPath), yesNoAllQuit)
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
		if err := c.destSystem.RemoveAll(destAbsPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err := c.sourceSystem.RemoveAll(sourceAbsPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
