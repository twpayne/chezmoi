package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
)

type passConfig struct {
	Command string
	cache   map[string][]byte
}

func (c *Config) passOutput(id string) []byte {
	if output, ok := c.Pass.cache[id]; ok {
		return output
	}

	name := c.Pass.Command
	args := []string{"show", id}
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s: %w", shellQuoteCommand(name, args), err))
		return nil
	}

	if c.Pass.cache == nil {
		c.Pass.cache = make(map[string][]byte)
	}
	c.Pass.cache[id] = output

	return output
}

func (c *Config) passTemplateFunc(id string, fields ...string) string {
	output := c.passOutput(id)

	if len(fields) > 0 {
		// Search fields in the form of "field: value"
		for _, field := range fields {
			for _, line := range bytes.Split(output, []byte("\n")) {
				if bytes.HasPrefix(line, append([]byte(field), ':')) {
					return string(bytes.TrimSpace(bytes.TrimPrefix(line, append([]byte(field), ':'))))
				}
			}
		}

		returnTemplateError(fmt.Errorf("found none of the fields in pass entry"))
		return ""
	}

	if index := bytes.IndexByte(output, '\n'); index != -1 {
		return string(output[:index])
	}
	return string(output)
}

func (c *Config) passRawTemplateFunc(id string) string {
	return string(c.passOutput(id))
}
