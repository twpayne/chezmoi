package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type secretConfig struct {
	Command   string
	cache     map[string]string
	jsonCache map[string]any
}

func (c *Config) secretTemplateFunc(args ...string) string {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.cache[key]; ok {
		return value
	}

	//nolint:gosec
	cmd := exec.Command(c.Secret.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	value := string(bytes.TrimSpace(output))
	if c.Secret.cache == nil {
		c.Secret.cache = make(map[string]string)
	}
	c.Secret.cache[key] = value

	return value
}

func (c *Config) secretJSONTemplateFunc(args ...string) any {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.jsonCache[key]; ok {
		return value
	}

	//nolint:gosec
	cmd := exec.Command(c.Secret.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	var value any
	if err := json.Unmarshal(output, &value); err != nil {
		panic(newParseCmdOutputError(c.Secret.Command, args, output, err))
	}

	if c.Secret.jsonCache == nil {
		c.Secret.jsonCache = make(map[string]any)
	}
	c.Secret.jsonCache[key] = value

	return value
}
