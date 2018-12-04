package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

var dumpCommand = &cobra.Command{
	Use:   "dump",
	Short: "Write a dump of the target state to stdout",
	RunE:  makeRunE(config.runDumpCommand),
}

func init() {
	rootCommand.AddCommand(dumpCommand)
}

func (c *Config) runDumpCommand(fs vfs.FS, command *cobra.Command, args []string) error {
	targetState, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		spew.Dump(targetState)
	} else {
		entries, err := c.getEntries(targetState, args)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			spew.Dump(entry)
		}
	}
	return nil
}
