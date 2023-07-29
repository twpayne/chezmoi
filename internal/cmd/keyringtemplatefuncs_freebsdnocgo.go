//go:build freebsd && !cgo

package cmd

type keyringData struct{}

func (c *Config) keyringTemplateFunc(service, user string) string {
	return ""
}
