package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"chezmoi.io/chezmoi/internal/chezmoilog"
)

type protonPassConfig struct {
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
	outputCache map[string][]byte
}

func (c *Config) protonPassTemplateFunc(item string) string {
	args := []string{"item", "view", item}
	return string(mustValue(c.protonPassOutput(args)))
}

func (c *Config) protonPassJSONTemplateFunc(item string) any {
	args := []string{"item", "view", item, "--output=json"}
	output := mustValue(c.protonPassOutput(args))
	var result map[string]any
	must(json.Unmarshal(output, &result))
	return result
}

func (c *Config) protonPassOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.ProtonPass.outputCache[key]; ok {
		return data, nil
	}

	cmd := exec.Command(c.ProtonPass.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.ProtonPass.outputCache == nil {
		c.ProtonPass.outputCache = make(map[string][]byte)
	}
	c.ProtonPass.outputCache[key] = output
	return output, nil
}
