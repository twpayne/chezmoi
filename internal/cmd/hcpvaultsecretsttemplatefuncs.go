package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type hcpVaultSecretConfig struct {
	Command         string   `json:"command"         mapstructure:"command"         yaml:"command"`
	Args            []string `json:"args"            mapstructure:"args"            yaml:"args"`
	ApplicationName string   `json:"applicationName" mapstructure:"applicationName" yaml:"applicationName"`
	OrganizationID  string   `json:"organizationId"  mapstructure:"organizationId"  yaml:"organizationId"`
	ProjectID       string   `json:"projectId"       mapstructure:"projectId"       yaml:"projectId"`
	outputCache     map[string][]byte
}

var vltMinVersion = semver.Version{Major: 0, Minor: 2, Patch: 1}

func (c *Config) hcpVaultSecretTemplateFunc(key string, additionalArgs ...string) string {
	args := mustValue(c.appendHCPVaultSecretsAdditionalArgs(
		[]string{"secrets", "get", "--plaintext"},
		additionalArgs,
	))
	return string(mustValue(c.vltOutput(append(args, key))))
}

func (c *Config) hcpVaultSecretJSONTemplateFunc(key string, additionalArgs ...string) any {
	args := mustValue(c.appendHCPVaultSecretsAdditionalArgs(
		[]string{"secrets", "get", "--format", "json"},
		additionalArgs,
	))
	data := mustValue(c.vltOutput(append(args, key)))
	var value any
	must(json.Unmarshal(data, &value))
	return value
}

func (c *Config) appendHCPVaultSecretsAdditionalArgs(args, additionalArgs []string) ([]string, error) {
	if len(additionalArgs) > 0 && additionalArgs[0] != "" {
		args = append(args, "--app-name", additionalArgs[0])
	} else if c.HCPVaultSecrets.ApplicationName != "" {
		args = append(args, "--app-name", c.HCPVaultSecrets.ApplicationName)
	}
	if len(additionalArgs) > 1 && additionalArgs[1] != "" {
		args = append(args, "--project", additionalArgs[1])
	} else if c.HCPVaultSecrets.ProjectID != "" {
		args = append(args, "--project", c.HCPVaultSecrets.ProjectID)
	}
	if len(additionalArgs) > 2 && additionalArgs[2] != "" {
		args = append(args, "--organization", additionalArgs[2])
	} else if c.HCPVaultSecrets.OrganizationID != "" {
		args = append(args, "--organization", c.HCPVaultSecrets.OrganizationID)
	}
	if len(additionalArgs) > 3 {
		// Add one to the number of received arguments as the hcpVaultSecret
		// and hcpVaultSecretJson template functions report this error and take
		// the key as the first argument.
		return nil, fmt.Errorf("expected 1 to 4 arguments, got %d", len(additionalArgs)+1)
	}
	return args, nil
}

func (c *Config) vltOutput(args []string) ([]byte, error) {
	args = append(slices.Clone(c.HCPVaultSecrets.Args), args...)
	key := strings.Join(args, "\x00")
	if data, ok := c.HCPVaultSecrets.outputCache[key]; ok {
		return data, nil
	}

	cmd := exec.Command(c.HCPVaultSecrets.Command, args...)
	// Always run the vlt command in the destination path because vlt uses
	// relative paths to find its .vlt.json config file.
	cmd.Dir = c.DestDirAbsPath.String()
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	if c.HCPVaultSecrets.outputCache == nil {
		c.HCPVaultSecrets.outputCache = make(map[string][]byte)
	}
	c.HCPVaultSecrets.outputCache[key] = output
	return output, nil
}
