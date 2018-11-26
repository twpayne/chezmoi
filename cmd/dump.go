package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Dump the target state",
	RunE:  makeRunE(config.runDumpCommandE),
}

func init() {
	rootCommand.AddCommand(dumpCommand)
}

func (c *Config) runDumpCommandE(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		spew.Dump(targetState)
	} else {
		for _, arg := range args {
			path, err := filepath.Abs(arg)
			if err != nil {
				return err
			}
			state, err := targetState.Get(path)
			if err != nil {
				return err
			}
			if state == nil {
				return fmt.Errorf("%s: not found", arg)
			}
			spew.Dump(state)
		}
	}
	return nil
}
