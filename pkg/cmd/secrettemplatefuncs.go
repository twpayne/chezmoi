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
	Command string
	Args    []string
	cache   map[string][]byte
}

func (c *Config) secretTemplateFunc(args ...string) string {
	output, err := c.secretOutput(args)
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimSpace(output))
}

func (c *Config) secretJSONTemplateFunc(args ...string) any {
	output, err := c.secretOutput(args)
	if err != nil {
		panic(err)
	}

	var value any
	if err := json.Unmarshal(output, &value); err != nil {
		panic(newParseCmdOutputError(c.Secret.Command, args, output, err))
	}
	return value
}

func (c *Config) secretOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if output, ok := c.Secret.cache[key]; ok {
		return output, nil
	}

	args = append(c.Secret.Args, args...)
	cmd := exec.Command(c.Secret.Command, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Secret.cache == nil {
		c.Secret.cache = make(map[string][]byte)
	}
	c.Secret.cache[key] = output

	return output, nil
}
