package cmd

import (
	"fmt"
	"strings"

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
	var pullArgs []string
	if c.SourceVCS.Pull != nil {
		switch v := c.SourceVCS.Pull.(type) {
		case string:
			pullArgs = strings.Split(v, " ")
		case []string:
			pullArgs = v
		default:
			return fmt.Errorf("sourceVCS.pull: cannot parse value")
		}
	} else {
		pullArgs = vcsInfo.pullArgs
	}
	if pullArgs == nil {
		return fmt.Errorf("%s: pull not supported", c.SourceVCS.Command)
	}

	if err := c.run(c.SourceDir, c.SourceVCS.Command, pullArgs...); err != nil {
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
