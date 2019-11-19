package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var passCmd = &cobra.Command{
	Use:     "pass [args...]",
	Short:   "Execute the pass CLI",
	PreRunE: config.ensureNoError,
	RunE:    config.runSecretPassCmd,
}

type passCmdConfig struct {
	Command string
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
	args := []string{"show", id}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		panic(fmt.Errorf("pass: %s %s: %w", name, strings.Join(args, " "), err))
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
