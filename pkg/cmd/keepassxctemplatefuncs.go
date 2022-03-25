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
	Command         string
	Database        chezmoi.AbsPath
	Args            []string
	version         *semver.Version
	cache           map[string]map[string]string
	attachmentCache map[string]map[string]string
	attributeCache  map[keepassxcAttributeCacheKey]string
	password        string
}

var (
	keepassxcPairRx                      = regexp.MustCompile(`^([A-Z]\w*):\s*(.*)$`)
	keepassxcNeedShowProtectedArgVersion = semver.Version{Major: 2, Minor: 5, Patch: 1}
)

func (c *Config) keepassxcAttachmentTemplateFunc(entry, name string) string {
	if data, ok := c.Keepassxc.attachmentCache[entry][name]; ok {
		return data
	}

	args := []string{"attachment-export", "--stdout"}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database.String(), entry, name)

	output, err := c.keepassxcOutput(c.Keepassxc.Command, args)
	if err != nil {
		panic(err)
	}

	return string(output)
}

func (c *Config) keepassxcTemplateFunc(entry string) map[string]string {
	if data, ok := c.Keepassxc.cache[entry]; ok {
		return data
	}

	args := []string{"show"}
	version, err := c.keepassxcVersion()
	if err != nil {
		panic(err)
	}
	if version.Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database.String(), entry)
	output, err := c.keepassxcOutput(c.Keepassxc.Command, args)
	if err != nil {
		panic(err)
	}

	data, err := keepassxcParseOutput(output)
	if err != nil {
		panic(newParseCmdOutputError(c.Keepassxc.Command, args, output, err))
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

	args := []string{"show", "--attributes", attribute, "--quiet"}
	version, err := c.keepassxcVersion()
	if err != nil {
		panic(err)
	}
	if version.Compare(keepassxcNeedShowProtectedArgVersion) >= 0 {
		args = append(args, "--show-protected")
	}
	args = append(args, c.Keepassxc.Args...)
	args = append(args, c.Keepassxc.Database.String(), entry)
	output, err := c.keepassxcOutput(c.Keepassxc.Command, args)
	if err != nil {
		panic(err)
	}

	outputStr := string(bytes.TrimSpace(output))
	if c.Keepassxc.attributeCache == nil {
		c.Keepassxc.attributeCache = make(map[keepassxcAttributeCacheKey]string)
	}
	c.Keepassxc.attributeCache[key] = outputStr

	return outputStr
}

func (c *Config) keepassxcOutput(name string, args []string) ([]byte, error) {
	if c.Keepassxc.Database.Empty() {
		panic(errors.New("keepassxc.database not set"))
	}

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

func keepassxcParseOutput(output []byte) (map[string]string, error) {
	data := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(output))
	var key string
	for i := 0; s.Scan(); i++ {
		switch match := keepassxcPairRx.FindStringSubmatch(s.Text()); {
		case match != nil:
			key = match[1]
			data[key] = match[2]
		case match == nil && key != "":
			data[key] += "\n" + s.Text()
		}
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
		return nil, &parseVersionError{
			output: output,
			err:    err,
		}
	}
	return c.Keepassxc.version, nil
}
