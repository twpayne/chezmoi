package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
)

func (c *Config) newForgetCmd() *cobra.Command {
	forgetCmd := &cobra.Command{
		Use:     "forget target...",
		Aliases: []string{"unmanage"},
		Short:   "Remove a target from the source state",
		Long:    mustLongHelp("forget"),
		Example: example("forget"),
		Args:    cobra.MinimumNArgs(1),
		RunE:    c.makeRunEWithSourceState(c.runForgetCmd),
		Annotations: map[string]string{
			modifiesSourceDirectory: "true",
		},
	}

	return forgetCmd
}

func (c *Config) runForgetCmd(cmd *cobra.Command, args []string, sourceState *chezmoi.SourceState) error {
	sourceAbsPaths, err := c.sourceAbsPaths(sourceState, args)
	if err != nil {
		return err
	}

	for _, sourceAbsPath := range sourceAbsPaths {
		if !c.force {
			choice, err := c.promptValue(fmt.Sprintf("Remove %s", sourceAbsPath), yesNoAllQuit)
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
	}

	return nil
}
