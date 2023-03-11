package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type rbwConfig struct {
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
	outputCache map[string][]byte
}

func (c *Config) rbwTemplateFunc(name string) map[string]any {
	args := []string{"get", "--raw", name}
	output, err := c.rbwOutput(args)
	if err != nil {
		panic(err)
	}
	var data map[string]any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.RBW.Command, args, output, err))
	}
	return data
}

func (c *Config) rbwOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.RBW.outputCache[key]; ok {
		return data, nil
	}

	cmd := exec.Command(c.RBW.Command, args...) //nolint:gosec
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.RBW.outputCache == nil {
		c.RBW.outputCache = make(map[string][]byte)
	}
	c.RBW.outputCache[key] = output
	return output, nil
}
