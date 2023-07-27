//go:build !darwin

package cmd

import (
	"github.com/spf13/cobra"
)

func (c *Config) newMackupCmd() *cobra.Command {
	return nil
}
