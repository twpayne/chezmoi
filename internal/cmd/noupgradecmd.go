// +build noupgrade windows

package cmd

import (
	"github.com/spf13/cobra"
)

type upgradeCmdConfig struct{}

func (c *Config) newUpgradeCmd() *cobra.Command {
	return nil
}
