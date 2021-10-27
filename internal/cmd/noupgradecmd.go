//go:build noupgrade || windows
// +build noupgrade windows

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
	return nil
}
