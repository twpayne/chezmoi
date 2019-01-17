package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var bitwardenCommand = &cobra.Command{
	Use:   "bitwarden",
	Short: "Execute the Bitwarden CLI (bw)",
	RunE:  makeRunE(config.runBitwardenCommand),
}

type bitwardenCommandConfig struct {
	Bw string
}

var bitwardenCache = make(map[string]interface{})

func init() {
	config.Bitwarden.Bw = "bw"
	config.addTemplateFunc("bitwarden", config.bitwardenFunc)

	secretCommand.AddCommand(bitwardenCommand)
}

func (c *Config) runBitwardenCommand(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.Bitwarden.Bw}, args...))
}

func (c *Config) bitwardenFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if data, ok := bitwardenCache[key]; ok {
		return data
	}
	name := c.Bitwarden.Bw
	args = append([]string{"get"}, args...)
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("bitwarden: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("bitwarden: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	bitwardenCache[key] = data
	return data
}
