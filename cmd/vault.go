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

var vaultCommand = &cobra.Command{
	Use:   "vault",
	Short: "Execute the Vault CLI",
	RunE:  makeRunE(config.runVaultCommand),
}

type vaultCommandConfig struct {
	Vault string
}

var vaultCache = make(map[string]interface{})

func init() {
	rootCommand.AddCommand(vaultCommand)

	config.Vault.Vault = "vault"
	config.addFunc("vault", config.vaultFunc)
}

func (c *Config) runVaultCommand(fs vfs.FS, args []string) error {
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
