package cmd

import (
	"chezmoi.io/chezmoi/v2/internal/chezmoi"
)

func (c *Config) exitInitTemplateFunc(code int) string {
	panic(chezmoi.ExitCodeError(code))
}

func (c *Config) writeToStdout(args ...string) string {
	for _, arg := range args {
		_ = mustValue(c.stdout.Write([]byte(arg)))
	}
	return ""
}
