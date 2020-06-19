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

var bitwardenCache = make(map[string]interface{})

func init() {
	config.Bitwarden.Command = "bw"
	config.addTemplateFunc("bitwarden", config.bitwardenFunc)

	secretCmd.AddCommand(bitwardenCmd)
}

func (c *Config) runBitwardenCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Bitwarden.Command, args...)
}

func (c *Config) bitwardenFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if data, ok := bitwardenCache[key]; ok {
		return data
	}
	name := c.Bitwarden.Command
	args = append([]string{"get"}, args...)
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
	bitwardenCache[key] = data
	return data
}
