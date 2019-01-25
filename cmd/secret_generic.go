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

var genericSecretCommand = &cobra.Command{
	Use:   "generic [args...]",
	Short: "Execute a generic secret command",
	RunE:  makeRunE(config.runGenericSecretCommand),
}

type genericSecretCommandConfig struct {
	Command string
}

var genericSecretCache = make(map[string][]byte)

func init() {
	config.addTemplateFunc("secret", config.secretFunc)
	config.addTemplateFunc("secretJSON", config.secretJSONFunc)

	secretCommand.AddCommand(genericSecretCommand)
}

func (c *Config) runGenericSecretCommand(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.GenericSecret.Command}, args...))
}

func (c *Config) secretFunc(args ...string) interface{} {
	// FIXME factor out common functionality with secretJSONFunc
	key := strings.Join(args, "\x00")
	if output, ok := genericSecretCache[key]; ok {
		return output
	}
	name := c.GenericSecret.Command
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("secret: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	genericSecretCache[key] = output
	return strings.TrimSpace(string(output))
}

func (c *Config) secretJSONFunc(args ...string) interface{} {
	// FIXME factor out common functionality with secretFunc
	key := strings.Join(args, "\x00")
	if output, ok := genericSecretCache[key]; ok {
		return output
	}
	name := c.GenericSecret.Command
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("secretJSON: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("secretJSON: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	genericSecretCache[key] = output
	return data
}
