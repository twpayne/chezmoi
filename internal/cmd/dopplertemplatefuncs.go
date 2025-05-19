package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type dopplerConfig struct {
	Command     string   `json:"command" mapstructure:"command" yaml:"command"`
	Args        []string `json:"args"    mapstructure:"args"    yaml:"args"`
	Project     string   `json:"project" mapstructure:"project" yaml:"project"`
	Config      string   `json:"config"  mapstructure:"config"  yaml:"config"`
	outputCache map[string][]byte
}

func (c *Config) dopplerTemplateFunc(key string, additionalArgs ...string) any {
	if len(additionalArgs) > 2 {
		// Add one to the number of received arguments as the key
		// is the first argument.
		panic(fmt.Errorf("expected 1 to 3 arguments, got %d", len(additionalArgs)+1))
	}

	args := c.appendDopplerAdditionalArgs([]string{"secrets", "download", "--json", "--no-file"}, additionalArgs)

	data := mustValue(c.dopplerOutput(args))
	var value map[string]any
	must(json.Unmarshal(data, &value))

	secret, ok := value[key]
	if !ok {
		panic(fmt.Errorf("could not find requested secret: %s", key))
	}

	return secret
}

func (c *Config) dopplerProjectJSONTemplateFunc(additionalArgs ...string) any {
	if len(additionalArgs) > 2 {
		panic(fmt.Errorf("expected 0 to 2 arguments, got %d", len(additionalArgs)))
	}
	args := c.appendDopplerAdditionalArgs([]string{"secrets", "download", "--json", "--no-file"}, additionalArgs)

	data := mustValue(c.dopplerOutput(args))
	var value any
	must(json.Unmarshal(data, &value))
	return value
}

func (c *Config) appendDopplerAdditionalArgs(args, additionalArgs []string) []string {
	if len(additionalArgs) > 0 && additionalArgs[0] != "" {
		args = append(args, "--project", additionalArgs[0])
	} else if c.Doppler.Project != "" {
		args = append(args, "--project", c.Doppler.Project)
	}
	if len(additionalArgs) > 1 && additionalArgs[1] != "" {
		args = append(args, "--config", additionalArgs[1])
	} else if c.Doppler.Config != "" {
		args = append(args, "--config", c.Doppler.Config)
	}

	return args
}

func (c *Config) dopplerOutput(args []string) ([]byte, error) {
	args = append(slices.Clone(c.Doppler.Args), args...)
	key := strings.Join(args, "\x00")
	if data, ok := c.Doppler.outputCache[key]; ok {
		return data, nil
	}
	cmd := exec.Command(c.Doppler.Command, args...)
	// Always run the doppler command in the destination path because doppler uses
	// relative paths to find its .doppler.json config file.
	cmd.Dir = c.DestDirAbsPath.String()
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.Doppler.outputCache == nil {
		c.Doppler.outputCache = make(map[string][]byte)
	}
	c.Doppler.outputCache[key] = output
	return output, nil
}
