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

var vaultCmd = &cobra.Command{
	Use:     "vault [args...]",
	Short:   "Execute the Hashicorp Vault CLI (vault)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runVaultCmd),
}

type vaultCmdConfig struct {
	Command string
}

var vaultCache = make(map[string]interface{})

func init() {
	config.Vault.Command = "vault"
	config.addTemplateFunc("vault", config.vaultFunc)

	secretCmd.AddCommand(vaultCmd)
}

func (c *Config) runVaultCmd(fs vfs.FS, args []string) error {
	return c.exec(fs, append([]string{c.Vault.Command}, args...))
}

func (c *Config) vaultFunc(key string) interface{} {
	if data, ok := vaultCache[key]; ok {
		return data
	}
	name := c.Vault.Command
	args := []string{"kv", "get", "-format=json", key}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("vault: %s %s: %w\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("vault: %s %s: %w\n%s", name, strings.Join(args, " "), err, output))
	}
	vaultCache[key] = data
	return data
}
