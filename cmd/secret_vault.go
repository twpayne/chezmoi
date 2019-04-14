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

var vaultCmd = &cobra.Command{
	Use:     "vault [args...]",
	Short:   "Execute the Hashicorp Vault CLI (vault)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runVaultCmd),
}

type vaultCmdConfig struct {
	Vault string
}

var vaultCache = make(map[string]interface{})

func init() {
	config.Vault.Vault = "vault"
	config.addTemplateFunc("vault", config.vaultFunc)

	secretCmd.AddCommand(vaultCmd)
}

func (c *Config) runVaultCmd(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.Vault.Vault}, args...))
}

func (c *Config) vaultFunc(key string) interface{} {
	if data, ok := vaultCache[key]; ok {
		return data
	}
	name := c.Vault.Vault
	args := []string{"kv", "get", "-format=json", key}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("vault: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("vault: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	vaultCache[key] = data
	return data
}
