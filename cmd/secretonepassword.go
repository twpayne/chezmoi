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

var onepasswordCmd = &cobra.Command{
	Use:     "onepassword [args...]",
	Short:   "Execute the 1Password CLI (op)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runOnepasswordCmd),
}

type onepasswordCmdConfig struct {
	Command string
}

var onepasswordCache = make(map[string]interface{})

func init() {
	config.Onepassword.Command = "op"
	config.addTemplateFunc("onepassword", config.onepasswordFunc)

	secretCmd.AddCommand(onepasswordCmd)
}

func (c *Config) runOnepasswordCmd(fs vfs.FS, args []string) error {
	return c.exec(fs, append([]string{c.Onepassword.Command}, args...))
}

func (c *Config) onepasswordFunc(item string) interface{} {
	if data, ok := onepasswordCache[item]; ok {
		return data
	}
	name := c.Onepassword.Command
	args := []string{"get", "item", item}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		panic(fmt.Errorf("onepassword: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	var data interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("onepassword: %s %s: %v\n%s", name, strings.Join(args, " "), err, output))
	}
	onepasswordCache[item] = data
	return data
}
