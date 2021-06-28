package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
)

var (
	// chezmoi uses gopass show --password which was added in
	// https://github.com/gopasspw/gopass/commit/8fa13d84e3656cfc4ee6717f5f485c9e471ad996
	// and the first tag containing that commit is v1.6.1.
	gopassMinVersion  = semver.Version{Major: 1, Minor: 6, Patch: 1}
	gopassVersionArgs = []string{"--version"}
	gopassVersionRx   = regexp.MustCompile(`gopass\s+(\d+\.\d+\.\d+)`)
)

type gopassConfig struct {
	Command   string
	versionOK bool
	cache     map[string]string
	rawCache  map[string][]byte
}

func (c *Config) gopassOutput(args ...string) ([]byte, error) {
	name := c.Gopass.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = c.stdin
	cmd.Stderr = c.stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Config) gopassRawTemplateFunc(id string) string {
	if !c.Gopass.versionOK {
		if err := c.gopassVersionCheck(); err != nil {
			returnTemplateError(err)
			return ""
		}
		c.Gopass.versionOK = true
	}

	if output, ok := c.Gopass.rawCache[id]; ok {
		return string(output)
	}

	args := []string{"show", id}
	output, err := c.gopassOutput(args...)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w", c.Gopass.Command, chezmoi.ShellQuoteArgs(args), err))
		return ""
	}

	if c.Gopass.rawCache == nil {
		c.Gopass.rawCache = make(map[string][]byte)
	}
	c.Gopass.rawCache[id] = output

	return string(output)
}

func (c *Config) gopassTemplateFunc(id string) string {
	if !c.Gopass.versionOK {
		if err := c.gopassVersionCheck(); err != nil {
			returnTemplateError(err)
			return ""
		}
		c.Gopass.versionOK = true
	}

	if password, ok := c.Gopass.cache[id]; ok {
		return password
	}

	args := []string{"show", "--password", id}
	output, err := c.gopassOutput(args...)
	if err != nil {
		returnTemplateError(fmt.Errorf("%s %s: %w", c.Gopass.Command, chezmoi.ShellQuoteArgs(args), err))
		return ""
	}

	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}

	if c.Gopass.cache == nil {
		c.Gopass.cache = make(map[string]string)
	}
	c.Gopass.cache[id] = password

	return password
}

func (c *Config) gopassVersionCheck() error {
	output, err := c.gopassOutput("--version")
	if err != nil {
		return err
	}
	m := gopassVersionRx.FindSubmatch(output)
	if m == nil {
		return fmt.Errorf("%s: could not extract version", output)
	}
	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		return err
	}
	if version.LessThan(gopassMinVersion) {
		return fmt.Errorf("version %s found, need version %s or later", version, gopassMinVersion)
	}
	return nil
}
