package cmd

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type vaultConfig struct {
	Command string
	cache   map[string]interface{}
}

func (c *Config) vaultTemplateFunc(key string) interface{} {
	if data, ok := c.Vault.cache[key]; ok {
		return data
	}

	args := []string{"kv", "get", "-format=json", key}
	//nolint:gosec
	cmd := exec.Command(c.Vault.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		panic(newCmdOutputError(cmd, output, err))
	}

	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Vault.Command, args, output, err))
	}

	if c.Vault.cache == nil {
		c.Vault.cache = make(map[string]interface{})
	}
	c.Vault.cache[key] = data

	return data
}
