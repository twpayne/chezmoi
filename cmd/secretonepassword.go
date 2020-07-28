package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var onepasswordCmd = &cobra.Command{
	Use:     "onepassword [args...]",
	Short:   "Execute the 1Password CLI (op)",
	PreRunE: config.ensureNoError,
	RunE:    config.runOnepasswordCmd,
}

type onepasswordCmdConfig struct {
	Command string
}

var onepasswordOutputCache = make(map[string][]byte)

func init() {
	config.Onepassword.Command = "op"
	config.addTemplateFunc("onepassword", config.onepasswordFunc)
	config.addTemplateFunc("onepasswordDocument", config.onepasswordDocumentFunc)
	config.addTemplateFunc("onepasswordDetailsFields", config.onepasswordDetailsFieldsFunc)

	secretCmd.AddCommand(onepasswordCmd)
}

func (c *Config) runOnepasswordCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Onepassword.Command, args...)
}

func (c *Config) onepasswordOutput(args []string) []byte {
	key := strings.Join(args, "\x00")
	if output, ok := onepasswordOutputCache[key]; ok {
		return output
	}

	name := c.Onepassword.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}

	onepasswordOutputCache[key] = output
	return output
}

func (c *Config) onepasswordFunc(args ...string) map[string]interface{} {
	key, vault := onepasswordGetKeyAndVault(args)
	onepasswordArgs := []string{"get", "item", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	output := c.onepasswordOutput(onepasswordArgs)
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", c.Onepassword.Command, chezmoi.ShellQuoteArgs(onepasswordArgs), err, output))
	}
	return data
}

func (c *Config) onepasswordDocumentFunc(args ...string) string {
	key, vault := onepasswordGetKeyAndVault(args)
	onepasswordArgs := []string{"get", "document", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	output := c.onepasswordOutput(onepasswordArgs)
	return string(output)
}

func (c *Config) onepasswordDetailsFieldsFunc(args ...string) map[string]interface{} {
	key, vault := onepasswordGetKeyAndVault(args)
	onepasswordArgs := []string{"get", "item", key}
	if vault != "" {
		onepasswordArgs = append(onepasswordArgs, "--vault", vault)
	}
	output := c.onepasswordOutput(onepasswordArgs)
	var data struct {
		Details struct {
			Fields []map[string]interface{} `json:"fields"`
		} `json:"details"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", c.Onepassword.Command, chezmoi.ShellQuoteArgs(onepasswordArgs), err, output))
	}
	result := make(map[string]interface{})
	for _, field := range data.Details.Fields {
		if designation, ok := field["designation"].(string); ok {
			result[designation] = field
		}
	}
	return result
}

func onepasswordGetKeyAndVault(args []string) (string, string) {
	switch len(args) {
	case 1:
		return args[0], ""
	case 2:
		return args[0], args[1]
	default:
		panic(fmt.Sprintf("expected 1 or 2 arguments, got %d", len(args)))
	}
}
