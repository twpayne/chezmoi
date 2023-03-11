//go:build freebsd && !cgo

package cmd

import "github.com/spf13/cobra"

type secretKeyringCmdConfig struct{}

func (c *Config) newSecretKeyringCmd() *cobra.Command {
	return nil
}
