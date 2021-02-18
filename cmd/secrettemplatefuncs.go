package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

type secretConfig struct {
	Command   string
	cache     map[string]string
	jsonCache map[string]interface{}
}

func (c *Config) secretJSONTemplateFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.jsonCache[key]; ok {
		return value
	}
	name := c.Secret.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	var value interface{}
	if err := json.Unmarshal(output, &value); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	if c.Secret.jsonCache == nil {
		c.Secret.jsonCache = make(map[string]interface{})
	}
	c.Secret.jsonCache[key] = value
	return value
}

func (c *Config) secretTemplateFunc(args ...string) string {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.cache[key]; ok {
		return value
	}
	name := c.Secret.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return ""
	}
	value := string(bytes.TrimSpace(output))
	if c.Secret.cache == nil {
		c.Secret.cache = make(map[string]string)
	}
	c.Secret.cache[key] = value
	return value
}
