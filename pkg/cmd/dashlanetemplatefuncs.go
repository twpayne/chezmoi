package cmd

import (
	"encoding/json"
	"os"
	"os/exec"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

type dashlaneConfig struct {
	Command       string   `json:"command" mapstructure:"command" yaml:"command"`
	Args          []string `json:"args"    mapstructure:"args"    yaml:"args"`
	cacheNote     map[string]any
	cachePassword map[string]any
}

func (c *Config) dashlaneNoteTemplateFunc(filter string) any {
	if data, ok := c.Dashlane.cacheNote[filter]; ok {
		return data
	}

	if c.Dashlane.cacheNote == nil {
		c.Dashlane.cacheNote = make(map[string]any)
	}

	output, err := c.dashlaneOutput("note", filter)
	if err != nil {
		panic(err)
	}

	data := string(output)

	c.Dashlane.cacheNote[filter] = data
	return data
}

func (c *Config) dashlanePasswordTemplateFunc(filter string) any {
	if data, ok := c.Dashlane.cachePassword[filter]; ok {
		return data
	}

	if c.Dashlane.cachePassword == nil {
		c.Dashlane.cachePassword = make(map[string]any)
	}

	output, err := c.dashlaneOutput("password", "--output", "json", filter)
	if err != nil {
		panic(err)
	}

	var data any
	if err := json.Unmarshal(output, &data); err != nil {
		panic(err)
	}

	c.Dashlane.cachePassword[filter] = data
	return data
}

func (c *Config) dashlaneOutput(args ...string) ([]byte, error) {
	name := c.Dashlane.Command
	cmd := exec.Command(name, append(c.Dashlane.Args, args...)...)
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}
