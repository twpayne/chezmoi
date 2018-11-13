package cmd

import (
	"os"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
)

var catCommand = &cobra.Command{
	Use:   "cat",
	Args:  cobra.MinimumNArgs(1),
	Short: "Print the contents of a file",
	RunE:  makeRunE(config.runCatCommand),
}

func init() {
	rootCommand.AddCommand(catCommand)
}

func (c *Config) runCatCommand(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	for _, arg := range args {
		state := targetState.Get(arg)
		if state == nil {
			return errors.Errorf("%s: not found", arg)
		}
		fileState, ok := state.(*chezmoi.FileState)
		if !ok {
			return errors.Errorf("%s: not a regular file", arg)
		}
		if _, err := os.Stdout.Write(fileState.Contents); err != nil {
			return err
		}
	}
	return nil
}
