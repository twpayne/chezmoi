// +build noupgrade windows

package chezmoicmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func (c *Config) runUpgradeCmd(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("%s: unsupported GOOS", runtime.GOOS)
}
