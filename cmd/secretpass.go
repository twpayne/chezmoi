package cmd

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/twpayne/chezmoi/internal/chezmoi"
)

var passCmd = &cobra.Command{
	Use:     "pass [args...]",
	Short:   "Execute the pass CLI",
	PreRunE: config.ensureNoError,
	RunE:    config.runSecretPassCmd,
}

type passCmdConfig struct {
	Command  string
	unlocked bool
}

var passCache = make(map[string]string)

func init() {
	secretCmd.AddCommand(passCmd)

	config.Pass.Command = "pass"
	config.addTemplateFunc("pass", config.passFunc)
}

func (c *Config) runSecretPassCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Pass.Command, args...)
}

func (c *Config) passFunc(id string) string {
	if s, ok := passCache[id]; ok {
		return s
	}
	name := c.Pass.Command
	if !c.Pass.unlocked {
		args := []string{"grep", "^$"}
		cmd := exec.Command(name, args...)
		cmd.Stdin = c.Stdin
		cmd.Stdout = c.Stdout
		cmd.Stderr = c.Stderr
		if err := cmd.Run(); err != nil {
			panic(fmt.Errorf("pass: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
		}
		c.Pass.unlocked = true
	}
	args := []string{"show", id}
	cmd := exec.Command(name, args...)
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		panic(fmt.Errorf("pass: %s %s: %w", name, chezmoi.ShellQuoteArgs(args), err))
	}
	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}
	passCache[id] = password
	return passCache[id]
}
