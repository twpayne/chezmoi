package cmd

import (
	"github.com/absfs/afero"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Dump the target state",
	RunE:  makeRunE(config.runDumpCommandE),
}

func init() {
	rootCommand.AddCommand(dumpCommand)
}

func (c *Config) runDumpCommandE(fs afero.Fs, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		spew.Dump(targetState)
	} else {
		for _, arg := range args {
			state := targetState.Get(arg)
			if state == nil {
				return errors.Errorf("%s: not found", arg)
			}
			spew.Dump(state)
		}
	}
	return nil
}
