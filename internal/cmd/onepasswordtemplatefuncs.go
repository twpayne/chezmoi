package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

type onepasswordConfig struct {
	Command     string
	outputCache map[string][]byte
}

func (c *Config) onepasswordDetailsFieldsTemplateFunc(args ...string) map[string]interface{} {
	onepasswordArgs := getOnepasswordArgs([]string{"get", "item"}, args)
	output := c.onepasswordOutput(onepasswordArgs)
	var data struct {
		Details struct {
			Fields []map[string]interface{} `json:"fields"`
		} `json:"details"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", c.Onepassword.Command, chezmoi.ShellQuoteArgs(onepasswordArgs), err, output))
		return nil
	}
	result := make(map[string]interface{})
	for _, field := range data.Details.Fields {
		if designation, ok := field["designation"].(string); ok {
			result[designation] = field
		}
	}
	return result
}

func (c *Config) onepasswordDocumentTemplateFunc(args ...string) string {
	onepasswordArgs := getOnepasswordArgs([]string{"get", "document"}, args)
	output := c.onepasswordOutput(onepasswordArgs)
	return string(output)
}

func (c *Config) onepasswordOutput(args []string) []byte {
	key := strings.Join(args, "\x00")
	if output, ok := c.Onepassword.outputCache[key]; ok {
		return output
	}

	name := c.Onepassword.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w: %s", name, chezmoi.ShellQuoteArgs(args), err, bytes.TrimSpace(stderr.Bytes())))
		return nil
	}

	if c.Onepassword.outputCache == nil {
		c.Onepassword.outputCache = make(map[string][]byte)
	}
	c.Onepassword.outputCache[key] = output
	return output
}

func (c *Config) onepasswordTemplateFunc(args ...string) map[string]interface{} {
	onepasswordArgs := getOnepasswordArgs([]string{"get", "item"}, args)
	output := c.onepasswordOutput(onepasswordArgs)
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", c.Onepassword.Command, chezmoi.ShellQuoteArgs(onepasswordArgs), err, output))
		return nil
	}
	return data
}

func getOnepasswordArgs(baseArgs, args []string) []string {
	if len(args) < 1 || len(args) > 3 {
		returnTemplateError(fmt.Errorf("expected 1, 2, or 3 arguments, got %d", len(args)))
		return nil
	}
	baseArgs = append(baseArgs, args[0])
	if len(args) > 1 {
		baseArgs = append(baseArgs, "--vault", args[1])
	}
	if len(args) > 2 {
		baseArgs = append(baseArgs, "--account", args[2])
	}
	return baseArgs
}
