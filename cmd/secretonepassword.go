package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

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

var (
	onepasswordCache         = make(map[string]interface{})
	onepasswordDocumentCache = make(map[string]string)
)

func init() {
	config.Onepassword.Command = "op"
	config.addTemplateFunc("onepassword", config.onepasswordFunc)
	config.addTemplateFunc("onepasswordDocument", config.onepasswordDocumentFunc)

	secretCmd.AddCommand(onepasswordCmd)
}

func (c *Config) runOnepasswordCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Onepassword.Command, args...)
}

func (c *Config) onepasswordFunc(args ...string) interface{} {
	if len(args) < 1 || len(args) > 2 {
		panic(fmt.Sprintf("expected 1 or 2 arguments, got %d", len(args)))
	}

	item := args[0]
	vault := ""
	key := item
	if len(args) == 2 {
		vault = args[1]
		key += "\x00" + vault
	}

	if data, ok := onepasswordCache[key]; ok {
		return data
	}
	name := c.Onepassword.Command
	args = []string{"get", "item", item}
	if vault != "" {
		args = append(args, "--vault", vault)
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	onepasswordCache[key] = data
	return data
}

func (c *Config) onepasswordDocumentFunc(args ...string) interface{} {
	if len(args) < 1 || len(args) > 2 {
		panic(fmt.Sprintf("expected 1 or 2 arguments, got %d", len(args)))
	}

	item := args[0]
	vault := ""
	key := item
	if len(args) == 2 {
		vault = args[1]
		key += "\x00" + vault
	}

	if output, ok := onepasswordDocumentCache[key]; ok {
		return output
	}
	name := c.Onepassword.Command
	args = []string{"get", "document", item}
	if vault != "" {
		args = append(args, "--vault", vault)
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", name, chezmoi.ShellQuoteArgs(args), err, output))
	}
	onepasswordDocumentCache[item] = string(output)
	return string(output)
}
