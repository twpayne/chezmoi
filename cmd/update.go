package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

type updateCommandConfig struct {
	apply bool
}

var updateCommand = &cobra.Command{
	Use:   "update",
	Args:  cobra.NoArgs,
	Short: "Pull changes from the source VCS and apply any changes",
	RunE:  makeRunE(config.runUpdateCommand),
}

func init() {
	rootCommand.AddCommand(updateCommand)

	persistentFlags := updateCommand.PersistentFlags()
	persistentFlags.BoolVarP(&config.update.apply, "apply", "a", true, "apply after pulling")
}

func (c *Config) runUpdateCommand(fs vfs.FS, args []string) error {
	vcsInfo, err := c.getVCSInfo()
	if err != nil {
		return err
	}
	if vcsInfo.pullArgs == nil {
		return fmt.Errorf("%s: pull not supported", c.SourceVCSCommand)
	}

	if err := c.run(c.SourceDir, c.SourceVCSCommand, vcsInfo.pullArgs...); err != nil {
		return err
	}

	if c.update.apply {
		mutator := c.getDefaultMutator(fs)
		if err := c.applyArgs(fs, nil, mutator); err != nil {
			return err
		}
	}

	return nil
}
