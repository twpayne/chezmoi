package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type secretConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
	cache   map[string][]byte
}

func (c *Config) secretTemplateFunc(args ...string) string {
	return string(bytes.TrimSpace(mustValue(c.secretOutput(args))))
}

func (c *Config) secretJSONTemplateFunc(args ...string) any {
	output := mustValue(c.secretOutput(args))

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

	args = append(slices.Clone(c.Secret.Args), args...)
	cmd := exec.Command(c.Secret.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Secret.cache == nil {
		c.Secret.cache = make(map[string][]byte)
	}
	c.Secret.cache[key] = output

	return output, nil
}
