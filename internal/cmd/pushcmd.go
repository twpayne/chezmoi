package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newPushCmd() *cobra.Command {
	pushCmd := &cobra.Command{
		GroupID: groupIDDaily,
		Use:     "push",
		Args:    cobra.NoArgs,
		Short:   "Add, commit, and push all changes in the source state",
		Long:    mustLongHelp("push"),
		Example: example("push"),
		RunE:    c.runPushCmd,
		Annotations: newAnnotations(
			persistentStateModeEmpty,
			requiresSourceDirectory,
			requiresWorkingTree,
		),
	}

	return pushCmd
}

func (c *Config) runPushCmd(cmd *cobra.Command, args []string) error {
	status, err := c.gitAdd()
	if err != nil {
		return err
	}
	if err := c.gitCommit(cmd, status); err != nil {
		return err
	}
	return c.gitPush(status)
}
