package chezmoicmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type vaultConfig struct {
	Command string
	cache   map[string]interface{}
}

func (c *Config) vaultTemplateFunc(key string) interface{} {
	if data, ok := c.Vault.cache[key]; ok {
		return data
	}
	name := c.Vault.Command
	args := []string{"kv", "get", "-format=json", key}
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	if c.Vault.cache == nil {
		c.Vault.cache = make(map[string]interface{})
	}
	c.Vault.cache[key] = data
	return data
}
