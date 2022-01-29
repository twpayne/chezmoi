package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type bitwardenConfig struct {
	Command     string
	outputCache map[string][]byte
}

func (c *Config) bitwardenAttachmentTemplateFunc(name, itemid string) string {
	output, err := c.bitwardenOutput([]string{"attachment", name, "--itemid", itemid, "--raw"})
	if err != nil {
		returnTemplateError(err)
		return ""
	}
	return string(output)
}

func (c *Config) bitwardenFieldsTemplateFunc(args ...string) map[string]interface{} {
	output, err := c.bitwardenOutput(args)
	if err != nil {
		returnTemplateError(err)
		return nil
	}
	var data struct {
		Fields []map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Bitwarden.Command, args), err, output))
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

func (c *Config) bitwardenTemplateFunc(args ...string) map[string]interface{} {
	output, err := c.bitwardenOutput(args)
	if err != nil {
		returnTemplateError(err)
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s: %w\n%s", shellQuoteCommand(c.Bitwarden.Command, args), err, output))
		return nil
	}
	return data
}

func (c *Config) bitwardenOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Bitwarden.outputCache[key]; ok {
		return data, nil
	}

	name := c.Bitwarden.Command
	args = append([]string{"get"}, args...)
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, fmt.Errorf("%s: %w\n%s", shellQuoteCommand(name, args), err, output)
	}

	if c.Bitwarden.outputCache == nil {
		c.Bitwarden.outputCache = make(map[string][]byte)
	}
	c.Bitwarden.outputCache[key] = output
	return output, nil
}
