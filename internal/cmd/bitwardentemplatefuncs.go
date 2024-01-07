package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type bitwardenConfig struct {
	outputCache map[string][]byte
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
}

func (c *Config) bitwardenAttachmentTemplateFunc(name, itemID string) string {
	output, err := c.bitwardenOutput([]string{"get", "attachment", name, "--itemid", itemID, "--raw"})
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) bitwardenAttachmentByRefTemplateFunc(name string, args ...string) string {
	output, err := c.bitwardenOutput(append([]string{"get"}, args...))
	if err != nil {
		panic(err)
	}
	var data struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Bitwarden.Command, args, output, err))
	}
	return c.bitwardenAttachmentTemplateFunc(name, data.ID)
}

func (c *Config) bitwardenFieldsTemplateFunc(args ...string) map[string]any {
	output, err := c.bitwardenOutput(append([]string{"get"}, args...))
	if err != nil {
		panic(err)
	}
	var data struct {
		Fields []map[string]any `json:"fields"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Bitwarden.Command, args, output, err))
	}
	result := make(map[string]any)
	for _, field := range data.Fields {
		if name, ok := field["name"].(string); ok {
			result[name] = field
		}
	}
	return result
}

func (c *Config) bitwardenTemplateFunc(args ...string) map[string]any {
	output, err := c.bitwardenOutput(append([]string{"get"}, args...))
	if err != nil {
		panic(err)
	}
	var data map[string]any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.Bitwarden.Command, args, output, err))
	}
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
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Bitwarden.outputCache == nil {
		c.Bitwarden.outputCache = make(map[string][]byte)
	}
	c.Bitwarden.outputCache[key] = output
	return output, nil
}
