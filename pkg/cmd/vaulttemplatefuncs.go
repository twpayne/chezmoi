package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
)

type vaultConfig struct {
	sync.Mutex
	Command string
	cache   map[string]interface{}
}

func (c *Config) vaultTemplateFunc(key string) interface{} {
	c.Vault.Lock()
	defer c.Vault.Unlock()

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
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(name, args), err, output))
		return nil
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(name, args), err, output))
		return nil
	}
	if c.Vault.cache == nil {
		c.Vault.cache = make(map[string]interface{})
	}
	c.Vault.cache[key] = data
	return data
}
