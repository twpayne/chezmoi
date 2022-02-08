package cmd

import (
	"bytes"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type passConfig struct {
	Command string
	cache   map[string][]byte
}

func (c *Config) passTemplateFunc(id string) string {
	output, err := c.passOutput(id)
	if err != nil {
		panic(err)
	}
	firstLine, _, _ := chezmoi.CutBytes(output, []byte{'\n'})
	return string(bytes.TrimSpace(firstLine))
}

func (c *Config) passFieldsTemplateFunc(id string) map[string]string {
	output, err := c.passOutput(id)
	if err != nil {
		panic(err)
	}

	result := make(map[string]string)
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		if key, value, ok := chezmoi.CutBytes(line, []byte{':'}); ok {
			result[string(bytes.TrimSpace(key))] = string(bytes.TrimSpace(value))
		}
	}
	return result
}

func (c *Config) passRawTemplateFunc(id string) string {
	output, err := c.passOutput(id)
	if err != nil {
		panic(err)
	}
	return string(output)
}

func (c *Config) passOutput(id string) ([]byte, error) {
	if output, ok := c.Pass.cache[id]; ok {
		return output, nil
	}

	args := []string{"show", id}
	//nolint:gosec
	cmd := exec.Command(c.Pass.Command, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Pass.cache == nil {
		c.Pass.cache = make(map[string][]byte)
	}
	c.Pass.cache[id] = output

	return output, nil
}
