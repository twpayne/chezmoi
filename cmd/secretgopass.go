package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var gopassCmd = &cobra.Command{
	Use:     "gopass [args...]",
	Short:   "Execute the gopass CLI",
	PreRunE: config.ensureNoError,
	RunE:    config.runSecretGopassCmd,
}

type gopassCmdConfig struct {
	Command string `json:"command" toml:"command" yaml:"command"`
}

var gopassCache = make(map[string]string)

func init() {
	secretCmd.AddCommand(gopassCmd)

	config.Gopass.Command = "gopass"
	config.addTemplateFunc("gopass", config.gopassFunc)
}

func (c *Config) runSecretGopassCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Pass.Command, args...)
}

func (c *Config) gopassFunc(id string) string {
	if s, ok := gopassCache[id]; ok {
		return s
	}
	name := c.Gopass.Command
	args := []string{"show", id}
	cmd := exec.Command(name, args...)
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("gopass: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}
	gopassCache[id] = password
	return gopassCache[id]
}
