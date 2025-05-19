package cmd

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type vaultConfig struct {
	Command string `json:"command" mapstructure:"command" yaml:"command"`
	cache   map[string]any
}

func (c *Config) vaultTemplateFunc(key string) any {
	if data, ok := c.Vault.cache[key]; ok {
		return data
	}

	args := []string{"kv", "get", "-format=json", key}
	cmd := exec.Command(c.Vault.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	var data any
	must(json.Unmarshal(output, &data))

	if c.Vault.cache == nil {
		c.Vault.cache = make(map[string]any)
	}
	c.Vault.cache[key] = data

	return data
}
