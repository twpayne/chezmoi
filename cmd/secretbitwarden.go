package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var bitwardenCmd = &cobra.Command{
	Use:     "bitwarden [args...]",
	Short:   "Execute the Bitwarden CLI (bw)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runBitwardenCmd),
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

func (c *Config) runBitwardenCmd(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.Bitwarden.Command}, args...))
}

func (c *Config) bitwardenFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if data, ok := bitwardenCache[key]; ok {
		return data
	}
	name := c.Bitwarden.Command
	args = append([]string{"get"}, args...)
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("bitwarden: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("bitwarden: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	bitwardenCache[key] = data
	return data
}
