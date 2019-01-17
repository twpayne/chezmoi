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

var onepasswordCommand = &cobra.Command{
	Use:   "onepassword",
	Short: "Execute the 1Password CLI",
	RunE:  makeRunE(config.runOnePasswordCommand),
}

type onepasswordCommandConfig struct {
	Op string
}

var onepasswordCache = make(map[string]interface{})

func init() {
	config.OnePassword.Op = "op"
	config.addFunc("onepassword", config.onepasswordFunc)

	_, err := exec.LookPath(config.OnePassword.Op)
	if err == nil {
		// op is installed
		secretCommand.AddCommand(onepasswordCommand)
	}
}

func (c *Config) runOnePasswordCommand(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.OnePassword.Op}, args...))
}

func (c *Config) onepasswordFunc(item string) interface{} {
	if data, ok := onepasswordCache[item]; ok {
		return data
	}
	name := c.OnePassword.Op
	args := []string{"get", "item", item}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("onepassword: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("onepassword: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	onepasswordCache[item] = data
	return data
}
