package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/coreos/go-semver/semver"

	"chezmoi.io/chezmoi/internal/chezmoilog"
)

type rbwConfig struct {
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
	outputCache map[string][]byte
}

var rbwMinVersion = semver.Version{Major: 1, Minor: 7, Patch: 0}

func (c *Config) rbwFieldsTemplateFunc(name string, extraArgs ...string) map[string]any {
	args := append([]string{"get", "--raw", name}, extraArgs...)
	output := mustValue(c.rbwOutput(args))
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

func (c *Config) rbwTemplateFunc(name string, extraArgs ...string) map[string]any {
	args := append([]string{"get", "--raw", name}, extraArgs...)
	output := mustValue(c.rbwOutput(args))
	var data map[string]any
	must(json.Unmarshal(output, &data))
	return data
}

func (c *Config) rbwOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.RBW.outputCache[key]; ok {
		return data, nil
	}

	cmd := exec.Command(c.RBW.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.RBW.outputCache == nil {
		c.RBW.outputCache = make(map[string][]byte)
	}
	c.RBW.outputCache[key] = output
	return output, nil
}
