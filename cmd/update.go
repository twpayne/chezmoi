package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type updateCmdConfig struct {
	apply bool
}

var updateCmd = &cobra.Command{
	Use:     "update",
	Args:    cobra.NoArgs,
	Short:   "Pull changes from the source VCS and apply any changes",
	Long:    mustGetLongHelp("update"),
	Example: getExample("update"),
	PreRunE: config.ensureNoError,
	RunE:    config.runUpdateCmd,
}

func init() {
	rootCmd.AddCommand(updateCmd)

	persistentFlags := updateCmd.PersistentFlags()
	persistentFlags.BoolVarP(&config.update.apply, "apply", "a", true, "apply after pulling")
}

func (c *Config) runUpdateCmd(cmd *cobra.Command, args []string) error {
	vcs, err := c.getVCS()
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
		pullArgs = vcs.PullArgs()
	}
	if pullArgs == nil {
		return fmt.Errorf("%s: pull not supported", c.SourceVCS.Command)
	}

	if err := c.run(c.SourceDir, c.SourceVCS.Command, pullArgs...); err != nil {
		return err
	}

	if c.update.apply {
		persistentState, err := c.getPersistentState(nil)
		if err != nil {
			return err
		}
		defer persistentState.Close()
		if err := c.applyArgs(nil, persistentState); err != nil {
			return err
		}
	}

	return nil
}
