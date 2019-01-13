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
	Short: "Execute the Bitwarden CLI",
	RunE:  makeRunE(config.runBitwardenCommand),
}

type bitwardenCommandConfig struct {
	BW      string
	Session string
}

var bitwardenCache = make(map[string]interface{})

func init() {
	rootCommand.AddCommand(bitwardenCommand)

	persistentFlags := rootCommand.PersistentFlags()
	persistentFlags.StringVar(&config.Bitwarden.Session, "bitwarden-session", "", "bitwarden session")

	config.Bitwarden.BW = "bw"
	config.addFunc("bitwarden", config.bitwardenFunc)
}

func (c *Config) runBitwardenCommand(fs vfs.FS, args []string) error {
	if c.Bitwarden.Session != "" {
		args = append([]string{"--session", c.Bitwarden.Session}, args...)
	}
	return c.exec(append([]string{c.Bitwarden.BW}, args...))
}

func (c *Config) bitwardenFunc(args ...string) interface{} {
	key := strings.Join(args, "\x00")
	if data, ok := bitwardenCache[key]; ok {
		return data
	}
	name := c.Bitwarden.BW
	if c.Bitwarden.Session != "" {
		args = append([]string{"--session", c.Bitwarden.Session}, args...)
	}
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
