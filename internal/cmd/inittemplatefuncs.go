package cmd

import (
	"os"

	"golang.org/x/term"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

func (c *Config) exitInitTemplateFunc(code int) string {
	panic(chezmoi.ExitCodeError(code))
}

func (c *Config) stdinIsATTYInitTemplateFunc() bool {
	if c.noTTY {
		return false
	}
	file, ok := c.stdin.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}

func (c *Config) writeToStdout(args ...string) string {
	for _, arg := range args {
		_ = mustValue(c.stdout.Write([]byte(arg)))
	}
	return ""
}
