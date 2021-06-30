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
	key, vault, account := onepasswordGetKeyAndVaultAndAccount(args)
	onepasswordArgs := []string{"get", "item", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	if account != "" {
		onepasswordArgs = append(onepasswordArgs, "--account", account)
	}
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
	key, vault, account := onepasswordGetKeyAndVaultAndAccount(args)
	onepasswordArgs := []string{"get", "document", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	if account != "" {
		onepasswordArgs = append(onepasswordArgs, "--account", account)
	}
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
	key, vault, account := onepasswordGetKeyAndVaultAndAccount(args)
	onepasswordArgs := []string{"get", "item", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	if account != "" {
		onepasswordArgs = append(onepasswordArgs, "--account", account)
	}
	output := c.onepasswordOutput(onepasswordArgs)
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w\n%s", c.Onepassword.Command, chezmoi.ShellQuoteArgs(onepasswordArgs), err, output))
		return nil
	}
	return data
}

func onepasswordGetKeyAndVaultAndAccount(args []string) (string, string, string) {
	switch len(args) {
	case 1:
		return args[0], "", ""
	case 2:
		return args[0], args[1], ""
	case 3:
		return args[0], args[1], args[2]
	default:
		returnTemplateError(fmt.Errorf("expected 1 or 2 or 3 arguments, got %d", len(args)))
		return "", "", ""
	}
}
