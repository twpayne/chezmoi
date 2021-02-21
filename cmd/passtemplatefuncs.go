package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type passConfig struct {
	Command string
	cache   map[string]string
}

func (c *Config) passTemplateFunc(id string) string {
	if s, ok := c.Pass.cache[id]; ok {
		return s
	}
	name := c.Pass.Command
	args := []string{"show", id}
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
		return ""
	}
	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}
	if c.Pass.cache == nil {
		c.Pass.cache = make(map[string]string)
	}
	c.Pass.cache[id] = password
	return password
}
