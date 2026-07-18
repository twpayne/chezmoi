//go:build freebsd && !cgo

package cmd

import "chezmoi.io/chezmoi/v2/internal/chezmoi"

type keyringData struct{}

func (c *Config) keyringTemplateFunc(service, user string) string {
	chezmoi.SkipTemplateIf(c.skipSecrets)
	return ""
}
