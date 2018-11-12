package cmd

import (
	"github.com/absfs/afero"
	"github.com/davecgh/go-spew/spew"
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
	spew.Dump(targetState)
	return nil
}
