//go:build noupgrade

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type upgradeCmdConfig struct {
	method string
	owner  string
	repo   string
}

func (c *Config) newUpgradeCmd() *cobra.Command {
	return nil
}

func getUpgradeMethod(vfs.FS, chezmoi.AbsPath) (string, error) {
	return "", nil
}
