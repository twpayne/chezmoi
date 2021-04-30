package cmd

import (
	"github.com/spf13/cobra"
)

type upgradeCmdConfig struct {
	method string
	owner  string
	repo   string
}

func (c *Config) newUpgradeCmd() *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade chezmoi to the latest released version",
		Long:    mustLongHelp("upgrade"),
		Example: example("upgrade"),
		Args:    cobra.NoArgs,
		RunE:    c.runUpgradeCmd,
		Annotations: map[string]string{
			runsCommands: "true",
		},
	}

	flags := upgradeCmd.Flags()
	flags.StringVar(&c.upgrade.method, "method", "", "set method")
	flags.StringVar(&c.upgrade.owner, "owner", defaultOwner, "set owner")
	flags.StringVar(&c.upgrade.repo, "repo", defaultRepo, "set repo")

	return upgradeCmd
}
