package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type passConfig struct {
	Command string
	cache   map[string][]byte
}

func (c *Config) passOutput(id string) []byte {
	if output, ok := c.Pass.cache[id]; ok {
		return output
	}

	name := c.Pass.Command
	args := []string{"show", id}
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s: %w", shellQuoteCommand(name, args), err))
		return nil
	}

	if c.Pass.cache == nil {
		c.Pass.cache = make(map[string][]byte)
	}
	c.Pass.cache[id] = output

	return output
}

func (c *Config) passTemplateFunc(id string) string {
	output, _, _ := chezmoi.CutBytes(c.passOutput(id), []byte{'\n'})
	return string(bytes.TrimSpace(output))
}

func (c *Config) passFieldsTemplateFunc(id string) map[string]string {
	output := c.passOutput(id)
	result := make(map[string]string)
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		if key, value, ok := chezmoi.CutBytes(line, []byte{':'}); ok {
			result[string(bytes.TrimSpace(key))] = string(bytes.TrimSpace(value))
		}
	}
	return result
}

func (c *Config) passRawTemplateFunc(id string) string {
	return string(c.passOutput(id))
}
