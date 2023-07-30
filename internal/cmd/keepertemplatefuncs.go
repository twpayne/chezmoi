package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type keeperConfig struct {
	Command     string   `json:"command" mapstructure:"command" yaml:"command"`
	Args        []string `json:"args"    mapstructure:"args"    yaml:"args"`
	outputCache map[string][]byte
}

func (c *Config) keeperTemplateFunc(record string) map[string]any {
	output, err := c.keeperOutput([]string{"get", "--format=json", record})
	if err != nil {
		panic(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		panic(err)
	}
	return result
}

func (c *Config) keeperDataFieldsTemplateFunc(record string) map[string]any {
	output, err := c.keeperOutput([]string{"get", "--format=json", record})
	if err != nil {
		panic(err)
	}
	var data struct {
		Data struct {
			Fields []struct {
				Type  string `json:"type"`
				Value any    `json:"value"`
			} `json:"fields"`
		} `json:"data"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(err)
	}
	result := make(map[string]any)
	for _, field := range data.Data.Fields {
		result[field.Type] = field.Value
	}
	return result
}

func (c *Config) keeperFindPasswordTemplateFunc(record string) string {
	output, err := c.keeperOutput([]string{"find-password", record})
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimSpace(output))
}

func (c *Config) keeperOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Keeper.outputCache[key]; ok {
		return data, nil
	}

	name := c.Keeper.Command
	args = append(args, c.Keeper.Args...)
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Keeper.outputCache == nil {
		c.Keeper.outputCache = make(map[string][]byte)
	}
	c.Keeper.outputCache[key] = output
	return output, nil
}
