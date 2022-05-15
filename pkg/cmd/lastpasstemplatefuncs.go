package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"

	"github.com/coreos/go-semver/semver"
)

var (
	// chezmoi uses lpass show --json which was added in
	// https://github.com/lastpass/lastpass-cli/commit/e5a22e2eeef31ab6c54595616e0f57ca0a1c162d
	// and the first tag containing that commit is v1.3.0~6.
	lastpassMinVersion  = semver.Version{Major: 1, Minor: 3, Patch: 0}
	lastpassParseNoteRx = regexp.MustCompile(`\A([ A-Za-z]*):(.*)\z`)
	lastpassVersionArgs = []string{"--version"}
	lastpassVersionRx   = regexp.MustCompile(`^LastPass CLI v(\d+\.\d+\.\d+)`)
)

type lastpassConfig struct {
	Command   string
	versionOK bool
	cache     map[string][]map[string]interface{}
}

func (c *Config) lastpassTemplateFunc(id string) []map[string]interface{} {
	data, err := c.lastpassData(id)
	if err != nil {
		panic(err)
	}
	for _, d := range data {
		if note, ok := d["note"].(string); ok {
			d["note"], err = lastpassParseNote(note)
			if err != nil {
				panic(err)
			}
		}
	}
	return data
}

func (c *Config) lastpassRawTemplateFunc(id string) []map[string]interface{} {
	data, err := c.lastpassData(id)
	if err != nil {
		panic(err)
	}
	return data
}

func (c *Config) lastpassData(id string) ([]map[string]interface{}, error) {
	if !c.Lastpass.versionOK {
		if err := c.lastpassVersionCheck(); err != nil {
			return nil, err
		}
		c.Lastpass.versionOK = true
	}

	if data, ok := c.Lastpass.cache[id]; ok {
		return data, nil
	}

	output, err := c.lastpassOutput("show", "--json", id)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("%s: parse error: %w", output, err)
	}

	if c.Lastpass.cache == nil {
		c.Lastpass.cache = make(map[string][]map[string]interface{})
	}
	c.Lastpass.cache[id] = data
	return data, nil
}

func (c *Config) lastpassOutput(args ...string) ([]byte, error) {
	name := c.Lastpass.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.baseSystem.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Config) lastpassVersionCheck() error {
	output, err := c.lastpassOutput("--version")
	if err != nil {
		return err
	}
	m := lastpassVersionRx.FindSubmatch(output)
	if m == nil {
		return &extractVersionError{
			output: output,
		}
	}
	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		return err
	}
	if version.LessThan(lastpassMinVersion) {
		return &versionTooOldError{
			have: version,
			need: &lastpassMinVersion,
		}
	}
	return nil
}

func lastpassParseNote(note string) (map[string]string, error) {
	result := make(map[string]string)
	s := bufio.NewScanner(bytes.NewBufferString(note))
	key := ""
	for s.Scan() {
		if m := lastpassParseNoteRx.FindStringSubmatch(s.Text()); m != nil {
			keyComponents := strings.Split(m[1], " ")
			firstComponentRunes := []rune(keyComponents[0])
			firstComponentRunes[0] = unicode.ToLower(firstComponentRunes[0])
			keyComponents[0] = string(firstComponentRunes)
			key = strings.Join(keyComponents, "")
			result[key] = m[2] + "\n"
		} else {
			result[key] += s.Text() + "\n"
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
