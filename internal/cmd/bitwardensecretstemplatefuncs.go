package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type bitwardenSecretsConfig struct {
	outputCache map[string][]byte
	Command     string `json:"command" mapstructure:"command" yaml:"command"`
}

func (c *Config) bitwardenSecretsTemplateFunc(secretID string, additionalArgs ...string) any {
	args := []string{"secret", "get", secretID}
	switch len(additionalArgs) {
	case 0:
		// Do nothing.
	case 1:
		args = append(args, "--access-token", additionalArgs[0])
	default:
		panic(fmt.Errorf("expected 1 or 2 arguments, got %d", len(additionalArgs)+1))
	}
	output, err := c.bitwardenSecretsOutput(args)
	if err != nil {
		panic(err)
	}
	var data map[string]any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(newParseCmdOutputError(c.BitwardenSecrets.Command, args, output, err))
	}
	return data
}

func (c *Config) bitwardenSecretsOutput(args []string) ([]byte, error) {
	key := strings.Join(args, "\x00")
	if data, ok := c.Bitwarden.outputCache[key]; ok {
		return data, nil
	}

	name := c.BitwardenSecrets.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.BitwardenSecrets.outputCache == nil {
		c.BitwardenSecrets.outputCache = make(map[string][]byte)
	}
	c.BitwardenSecrets.outputCache[key] = output
	return output, nil
}
