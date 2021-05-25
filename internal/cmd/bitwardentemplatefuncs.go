package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type bitwardenConfig struct {
	Command     string
	outputCache map[string][]byte
}

func (c *Config) bitwardenFieldsTemplateFunc(args ...string) map[string]interface{} {
	output := c.bitwardenOutput(args)
	var data struct {
		Fields []map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", c.Bitwarden.Command, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	result := make(map[string]interface{})
	for _, field := range data.Fields {
		if name, ok := field["name"].(string); ok {
			result[name] = field
		}
	}
	return result
}

func (c *Config) bitwardenOutput(args []string) []byte {
	key := strings.Join(args, "\x00")
	if data, ok := c.Bitwarden.outputCache[key]; ok {
		return data
	}

	name := c.Bitwarden.Command
	args = append([]string{"get"}, args...)
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}

	if c.Bitwarden.outputCache == nil {
		c.Bitwarden.outputCache = make(map[string][]byte)
	}
	c.Bitwarden.outputCache[key] = output
	return output
}

func (c *Config) bitwardenTemplateFunc(args ...string) map[string]interface{} {
	output := c.bitwardenOutput(args)
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", c.Bitwarden.Command, chezmoi.ShellQuoteArgs(args), err, output))
		return nil
	}
	return data
}

func (c *Config) bitwardenAttachmentTemplateFunc(name, itemid string) string {
	return string(c.bitwardenOutput([]string{"attachment", name, "--itemid", itemid, "--raw"}))
}
