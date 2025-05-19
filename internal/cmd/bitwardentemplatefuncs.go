package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type bitwardenConfig struct {
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
	outputCache map[string][]byte
}

func (c *Config) bitwardenAttachmentTemplateFunc(name, itemID string) string {
	return string(mustValue(c.bitwardenOutput([]string{"get", "attachment", name, "--itemid", itemID, "--raw"})))
}

func (c *Config) bitwardenAttachmentByRefTemplateFunc(name string, args ...string) string {
	output := mustValue(c.bitwardenOutput(append([]string{"get"}, args...)))
	var data struct {
		ID string `json:"id"`
	}
	must(json.Unmarshal(output, &data))
	return c.bitwardenAttachmentTemplateFunc(name, data.ID)
}

func (c *Config) bitwardenFieldsTemplateFunc(args ...string) map[string]any {
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
	output := mustValue(c.bitwardenOutput(append([]string{"get"}, args...)))
	var data map[string]any
	must(json.Unmarshal(output, &data))
	return data
}

func (c *Config) bitwardenOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Bitwarden.outputCache[key]; ok {
		return data, nil
	}

	name := c.Bitwarden.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Bitwarden.outputCache == nil {
		c.Bitwarden.outputCache = make(map[string][]byte)
	}
	c.Bitwarden.outputCache[key] = output
	return output, nil
}
