package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var passCommand = &cobra.Command{
	Use:   "pass",
	Short: "Execute the pass CLI",
	RunE:  makeRunE(config.runPassCommand),
}

// A PassCommandConfig is a configuration for the pass command.
type PassCommandConfig struct {
	Pass string
}

var passCache = make(map[string]string)

func init() {
	rootCommand.AddCommand(passCommand)
	config.Pass.Pass = "pass"
	config.addFunc("pass", config.passFunc)
}

func (c *Config) runPassCommand(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.Pass.Pass}, args...))
}

func (c *Config) passFunc(id string) string {
	if s, ok := passCache[id]; ok {
		return s
	}
	name := c.Pass.Pass
	args := []string{id}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("pass: %s %s: %v", name, strings.Join(args, " "), err))
	}
	password := strings.Split(string(output), "\n")[0]
	passCache[id] = password
	return passCache[id]
}
