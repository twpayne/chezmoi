package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"chezmoi.io/chezmoi/internal/chezmoilog"
)

type bitwardenConfig struct {
	Command     string   `json:"command" mapstructure:"command" yaml:"command"`
	Unlock      autoBool `json:"unlock"  mapstructure:"unlock"  yaml:"unlock"`
	session     string
	outputCache map[string][]byte
}

func (c *Config) bitwardenAttachmentTemplateFunc(name, itemID string) string {
	must(c.bitwardenMaybeUnlock())
	return string(mustValue(c.bitwardenOutput([]string{"get", "attachment", name, "--itemid", itemID, "--raw"})))
}

func (c *Config) bitwardenAttachmentByRefTemplateFunc(name string, args ...string) string {
	must(c.bitwardenMaybeUnlock())
	output := mustValue(c.bitwardenOutput(append([]string{"get"}, args...)))
	var data struct {
		ID string `json:"id"`
	}
	must(json.Unmarshal(output, &data))
	return c.bitwardenAttachmentTemplateFunc(name, data.ID)
}

func (c *Config) bitwardenFieldsTemplateFunc(args ...string) map[string]any {
	must(c.bitwardenMaybeUnlock())
	output := mustValue(c.bitwardenOutput(append([]string{"get"}, args...)))
	var data struct {
		Fields []map[string]any `json:"fields"`
	}
	must(json.Unmarshal(output, &data))
	result := make(map[string]any)
	for _, field := range data.Fields {
		if name, ok := field["name"].(string); ok {
			result[name] = field
		}
	}
	return result
}

func (c *Config) bitwardenTemplateFunc(args ...string) map[string]any {
	must(c.bitwardenMaybeUnlock())
	output := mustValue(c.bitwardenOutput(append([]string{"get"}, args...)))
	var data map[string]any
	must(json.Unmarshal(output, &data))
	return data
}

func (c *Config) bitwardenLock() error {
	if c.Bitwarden.session == "" {
		return nil
	}
	output, err := c.bitwardenOutput([]string{"lock"})
	if err != nil {
		return fmt.Errorf("%w: %s", err, output)
	}
	return nil
}

func (c *Config) bitwardenMaybeUnlock() error {
	unlock := c.Bitwarden.Unlock.Value(func() bool {
		_, ok := os.LookupEnv("BW_SESSION")
		return !ok
	})
	if !unlock {
		return nil
	}
	output, err := c.bitwardenOutput([]string{"unlock", "--raw"})
	if err != nil {
		return err
	}
	c.Bitwarden.session = string(bytes.TrimSpace(output))
	return nil
}

func (c *Config) bitwardenOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Bitwarden.outputCache[key]; ok {
		return data, nil
	}

	output, err := c.bitwardenUncachedOutput(args)
	if err != nil {
		return nil, err
	}

	if c.Bitwarden.outputCache == nil {
		c.Bitwarden.outputCache = make(map[string][]byte)
	}
	c.Bitwarden.outputCache[key] = output
	return output, nil
}

func (c *Config) bitwardenUncachedOutput(args []string) ([]byte, error) {
	name := c.Bitwarden.Command
	cmd := exec.Command(name, args...)
	if c.Bitwarden.session != "" {
		cmd.Env = append(os.Environ(), "BW_SESSION="+c.Bitwarden.session)
	}
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}
	return output, err
}
