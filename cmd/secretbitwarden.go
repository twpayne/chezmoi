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

var bitwardenCmd = &cobra.Command{
	Use:     "bitwarden [args...]",
	Short:   "Execute the Bitwarden CLI (bw)",
	PreRunE: config.ensureNoError,
	RunE:    config.runBitwardenCmd,
}

type bitwardenCmdConfig struct {
	Command string
}

var bitwardenOutputCache = make(map[string][]byte)

func init() {
	config.Bitwarden.Command = "bw"
	config.addTemplateFunc("bitwarden", config.bitwardenFunc)
	config.addTemplateFunc("bitwardenFields", config.bitwardenFieldsFunc)

	secretCmd.AddCommand(bitwardenCmd)
}

func (c *Config) runBitwardenCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Bitwarden.Command, args...)
}

func (c *Config) bitwardenOutput(args []string) []byte {
	key := strings.Join(args, "\x00")
	if output, ok := bitwardenOutputCache[key]; ok {
		return output
	}

	//nolint:gosec
	cmd := exec.Command(c.Bitwarden.Command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", c.Bitwarden.Command, chezmoi.ShellQuoteArgs(args), err, output))
	}

	bitwardenOutputCache[key] = output
	return output
}

func (c *Config) bitwardenFunc(args ...string) map[string]interface{} {
	output := c.bitwardenOutput(append([]string{"get"}, args...))
	var data map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", c.Bitwarden.Command, chezmoi.ShellQuoteArgs(args), err, output))
	}
	return data
}

func (c *Config) bitwardenFieldsFunc(args ...string) map[string]interface{} {
	output := c.bitwardenOutput(append([]string{"get"}, args...))
	var data struct {
		Fields []map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("%s %s: %w\n%s", c.Bitwarden.Command, chezmoi.ShellQuoteArgs(args), err, output))
	}
	result := make(map[string]interface{})
	for _, field := range data.Fields {
		if name, ok := field["name"].(string); ok {
			result[name] = field
		}
	}
	return result
}
