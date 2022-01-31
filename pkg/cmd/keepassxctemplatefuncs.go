package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type keepassxcAttributeCacheKey struct {
	entry     string
	attribute string
}

type keepassxcConfig struct {
	Command        string
	Database       chezmoi.AbsPath
	Args           []string
	version        *semver.Version
	cache          map[string]map[string]string
	attributeCache map[keepassxcAttributeCacheKey]string
	password       string
}

var (
	keepassxcPairRx                      = regexp.MustCompile(`^([^:]+):\s*(.*)$`)
	keepassxcNeedShowProtectedArgVersion = semver.Version{Major: 2, Minor: 5, Patch: 1}
)

func (c *Config) keepassxcTemplateFunc(entry string) map[string]string {
	if data, ok := c.Keepassxc.cache[entry]; ok {
		return data
	}

	if c.Keepassxc.Database.Empty() {
		raiseTemplateError(errors.New("keepassxc.database not set"))
		return nil
	}

	args := []string{"show"}
	version, err := c.keepassxcVersion()
	if err != nil {
		raiseTemplateError(err)
		return nil
	}
	if version.Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database.String(), entry)
	output, err := c.keepassxcOutput(c.Keepassxc.Command, args)
	if err != nil {
		raiseTemplateError(err)
		return nil
	}

	data, err := keypassxcParseOutput(output)
	if err != nil {
		raiseTemplateError(newParseCmdOutputError(c.Keepassxc.Command, args, output, err))
		return nil
	}

	if c.Keepassxc.cache == nil {
		c.Keepassxc.cache = make(map[string]map[string]string)
	}
	c.Keepassxc.cache[entry] = data

	return data
}

func (c *Config) keepassxcAttributeTemplateFunc(entry, attribute string) string {
	key := keepassxcAttributeCacheKey{
		entry:     entry,
		attribute: attribute,
	}
	if data, ok := c.Keepassxc.attributeCache[key]; ok {
		return data
	}

	if c.Keepassxc.Database.Empty() {
		raiseTemplateError(errors.New("keepassxc.database not set"))
		return ""
	}

	args := []string{"show", "--attributes", attribute, "--quiet"}
	version, err := c.keepassxcVersion()
	if err != nil {
		raiseTemplateError(err)
		return ""
	}
	if version.Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database.String(), entry)
	output, err := c.keepassxcOutput(c.Keepassxc.Command, args)
	if err != nil {
		raiseTemplateError(err)
		return ""
	}

	outputStr := string(bytes.TrimSpace(output))
	if c.Keepassxc.attributeCache == nil {
		c.Keepassxc.attributeCache = make(map[keepassxcAttributeCacheKey]string)
	}
	c.Keepassxc.attributeCache[key] = outputStr

	return outputStr
}

func (c *Config) keepassxcOutput(name string, args []string) ([]byte, error) {
	if c.Keepassxc.password == "" {
		password, err := c.readPassword(fmt.Sprintf("Insert password to unlock %s: ", c.Keepassxc.Database))
		if err != nil {
			return nil, err
		}
		c.Keepassxc.password = password
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = bytes.NewBufferString(c.Keepassxc.password + "\n")
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}
	return output, nil
}

func keypassxcParseOutput(output []byte) (map[string]string, error) {
	data := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(output))
	for i := 0; s.Scan(); i++ {
		if i == 0 {
			continue
		}
		match := keepassxcPairRx.FindStringSubmatch(s.Text())
		if match == nil {
			return nil, fmt.Errorf("%s: parse error", s.Text())
		}
		data[match[1]] = match[2]
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

func (c *Config) keepassxcVersion() (*semver.Version, error) {
	if c.Keepassxc.version != nil {
		return c.Keepassxc.version, nil
	}

	name := c.Keepassxc.Command
	args := []string{"--version"}
	cmd := exec.Command(name, args...)
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}

	c.Keepassxc.version, err = semver.NewVersion(string(bytes.TrimSpace(output)))
	if err != nil {
		return nil, fmt.Errorf("cannot parse version %s: %w", output, err)
	}
	return c.Keepassxc.version, nil
}
