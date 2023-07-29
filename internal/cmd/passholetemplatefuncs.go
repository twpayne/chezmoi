package cmd

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/coreos/go-semver/semver"
	"golang.org/x/exp/slices"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type passholeCacheKey struct {
	path  string
	field string
}

type passholeConfig struct {
	Command  string   `json:"command" mapstructure:"command" yaml:"command"`
	Args     []string `json:"args"    mapstructure:"args"    yaml:"args"`
	Prompt   bool     `json:"prompt"  mapstructure:"prompt"  yaml:"prompt"`
	cache    map[passholeCacheKey]string
	password string
}

var passholeMinVersion = semver.Version{Major: 1, Minor: 10, Patch: 0}

func (c *Config) passholeTemplateFunc(path, field string) string {
	key := passholeCacheKey{
		path:  path,
		field: field,
	}
	if value, ok := c.Passhole.cache[key]; ok {
		return value
	}

	args := slices.Clone(c.Passhole.Args)
	var stdin io.Reader
	if c.Passhole.Prompt {
		if c.Passhole.password == "" {
			password, err := c.readPassword("Enter database password: ")
			if err != nil {
				panic(err)
			}
			c.Passhole.password = password
		}
		args = append(args, "--password", "-")
		stdin = bytes.NewBufferString(c.Passhole.password + "\n")
	}
	args = append(args, "show", "--field", field, path)
	output, err := c.passholeOutput(c.Passhole.Command, args, stdin)
	if err != nil {
		panic(err)
	}

	if c.Passhole.cache == nil {
		c.Passhole.cache = make(map[passholeCacheKey]string)
	}
	c.Passhole.cache[key] = output
	return output
}

func (c *Config) passholeOutput(name string, args []string, stdin io.Reader) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return "", newCmdOutputError(cmd, output, err)
	}
	return string(output), nil
}
