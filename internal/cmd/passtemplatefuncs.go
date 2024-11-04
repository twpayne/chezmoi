package cmd

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type passConfig struct {
	Command string `json:"command" mapstructure:"command" yaml:"command"`
	cache   map[string][]byte
}

func (c *Config) passTemplateFunc(id string) string {
	output := mustValue(c.passOutput(id))
	firstLine, _, _ := bytes.Cut(output, []byte{'\n'})
	return string(bytes.TrimSpace(firstLine))
}

func (c *Config) passFieldsTemplateFunc(id string) map[string]string {
	output := mustValue(c.passOutput(id))

	result := make(map[string]string)
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		if key, value, ok := bytes.Cut(line, []byte{':'}); ok {
			result[string(bytes.TrimSpace(key))] = string(bytes.TrimSpace(value))
		}
	}
	return result
}

func (c *Config) passRawTemplateFunc(id string) string {
	return string(mustValue(c.passOutput(id)))
}

func (c *Config) passOutput(id string) ([]byte, error) {
	if output, ok := c.Pass.cache[id]; ok {
		return output, nil
	}

	args := []string{"show", id}
	cmd := exec.Command(c.Pass.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Pass.cache == nil {
		c.Pass.cache = make(map[string][]byte)
	}
	c.Pass.cache[id] = output

	return output, nil
}
