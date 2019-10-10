package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var gopassCmd = &cobra.Command{
	Use:     "gopass [args...]",
	Short:   "Execute the gopass CLI",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runSecretGopassCmd),
}

type gopassCmdConfig struct {
	Command string
}

var gopassCache = make(map[string]string)

func init() {
	secretCmd.AddCommand(gopassCmd)

	config.Gopass.Command = "gopass"
	config.addTemplateFunc("gopass", config.gopassFunc)
}

func (c *Config) runSecretGopassCmd(fs vfs.FS, args []string) error {
	return c.exec(fs, append([]string{c.Pass.Command}, args...))
}

func (c *Config) gopassFunc(id string) string {
	if s, ok := gopassCache[id]; ok {
		return s
	}
	name := c.Gopass.Command
	args := []string{"show", id}
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).Output()
	if err != nil {
		panic(fmt.Errorf("gopass: %s %s: %w", name, strings.Join(args, " "), err))
	}
	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}
	gopassCache[id] = password
	return gopassCache[id]
}
