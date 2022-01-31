package cmd

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
)

type secretConfig struct {
	Command   string
	cache     map[string]string
	jsonCache map[string]interface{}
}

func (c *Config) secretTemplateFunc(args ...string) string {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.cache[key]; ok {
		return value
	}

	//nolint:gosec
	cmd := exec.Command(c.Secret.Command, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		raiseTemplateError(newCmdOutputError(cmd, output, err))
		return ""
	}

	value := string(bytes.TrimSpace(output))
	if c.Secret.cache == nil {
		c.Secret.cache = make(map[string]string)
	}
	c.Secret.cache[key] = value

	return value
}

func (c *Config) secretJSONTemplateFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if value, ok := c.Secret.jsonCache[key]; ok {
		return value
	}

	//nolint:gosec
	cmd := exec.Command(c.Secret.Command, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		raiseTemplateError(newCmdOutputError(cmd, output, err))
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(output, &value); err != nil {
		raiseTemplateError(newParseCmdOutputError(c.Secret.Command, args, output, err))
		return nil
	}

	if c.Secret.jsonCache == nil {
		c.Secret.jsonCache = make(map[string]interface{})
	}
	c.Secret.jsonCache[key] = value

	return value
}
