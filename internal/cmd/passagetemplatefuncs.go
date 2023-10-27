package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"

	"github.com/coreos/go-semver/semver"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

var (
	// just requiring the current ( as of 2023-10-27 ) version of passage.
	passageMinVersion  = semver.Version{Major: 1, Minor: 7, Patch: 4}
	passageShowArgs    = []string{"show"}
	passageVersionArgs = []string{"version"}
	passageVersionRx   = regexp.MustCompile(`=\s+v(\d+\.\d+\.\d+)`)
)

type passageConfig struct {
	Command   string `json:"command" mapstructure:"command" yaml:"command"`
	versionOK bool
	cache     map[string]string
	rawCache  map[string][]byte
}

func (c *Config) passageTemplateFunc(id string) string {
	c.passageValid()

	if password, ok := c.Passage.cache[id]; ok {
		return password
	}

	output := c.fetchPassage(id)
	passwordBytes, _, _ := bytes.Cut(output, []byte{'\n'})
	password := string(passwordBytes)

	if c.Passage.cache == nil {
		c.Passage.cache = map[string]string{}
	}
	c.Passage.cache[id] = password

	return password
}

func (c *Config) passageRawTemplateFunc(id string) string {
	c.passageValid()

	if output, ok := c.Passage.rawCache[id]; ok {
		return string(output)
	}

	output := c.fetchPassage(id)
	if c.Passage.rawCache == nil {
		c.Passage.rawCache = map[string][]byte{}
	}
	c.Passage.rawCache[id] = output

	return string(output)
}

func (c *Config) fetchPassage(id string) []byte {
	var args []string
	args = append(args, passageShowArgs...)
	args = append(args, id)
	output, err := c.passageOutput(args...)
	if err != nil {
		panic(err)
	}
	return output
}

func (c *Config) passageValid() {
	if !c.Passage.versionOK {
		if err := c.passageVersionCheck(); err != nil {
			panic(err)
		}
		c.Passage.versionOK = true
	}
}

func (c *Config) passageOutput(args ...string) ([]byte, error) {
	name := c.Passage.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}
	return output, nil
}

func (c *Config) passageVersionCheck() error {
	output, err := c.passageOutput(passageVersionArgs...)
	if err != nil {
		return err
	}
	m := passageVersionRx.FindSubmatch(output)
	if m == nil {
		return &extractVersionError{
			output: output,
		}
	}

	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		return err
	}
	if version.LessThan(passageMinVersion) {
		return &versionTooOldError{
			have: version,
			need: &passageMinVersion,
		}
	}
	return nil
}
