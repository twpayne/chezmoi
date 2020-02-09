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
	"sync"
	"unicode"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

var lastpassCmd = &cobra.Command{
	Use:     "lastpass [args...]",
	Short:   "Execute the LastPass CLI (lpass)",
	PreRunE: config.ensureNoError,
	RunE:    config.runLastpassCmd,
}

var (
	// chezmoi uses lpass show --json which was added in
	// https://github.com/lastpass/lastpass-cli/commit/e5a22e2eeef31ab6c54595616e0f57ca0a1c162d
	// and the first tag containing that commit is v1.3.0~6.
	lastpassMinVersion      = semver.Version{Major: 1, Minor: 3, Patch: 0}
	lastpassParseNoteRegexp = regexp.MustCompile(`\A([ A-Za-z]*):(.*)\z`)
	lastpassVersionArgs     = []string{"--version"}
	lastpassVersionRegexp   = regexp.MustCompile(`^LastPass CLI v(\d+\.\d+\.\d+)`)
)

type lastpassCmdConfig struct {
	Command          string
	versionCheckOnce sync.Once
}

var lastPassCache = make(map[string]interface{})

func init() {
	config.Lastpass.Command = "lpass"
	config.addTemplateFunc("lastpass", config.lastpassFunc)

	secretCmd.AddCommand(lastpassCmd)
}

func (c *Config) runLastpassCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Lastpass.Command, args...)
}

func (c *Config) lastpassOutput(args ...string) ([]byte, error) {
	name := c.Lastpass.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Config) lastpassFunc(id string) interface{} {
	c.Lastpass.versionCheckOnce.Do(func() {
		panicOnError(c.lastpassVersionCheck())
	})
	if data, ok := lastPassCache[id]; ok {
		return data
	}
	output, err := c.lastpassOutput("show", "--json", id)
	panicOnError(err)
	var data []map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		panic(fmt.Errorf("lastpass: parse error: %w\n%q", err, output))
	}
	for _, d := range data {
		if note, ok := d["note"].(string); ok {
			d["note"] = lastpassParseNote(note)
		}
	}
	lastPassCache[id] = data
	return data
}

func (c *Config) lastpassVersionCheck() error {
	output, err := c.lastpassOutput(lastpassVersionArgs...)
	if err != nil {
		return err
	}
	m := lastpassVersionRegexp.FindSubmatch(output)
	if m == nil {
		return fmt.Errorf("lastpass: could not extract version from %q", output)
	}
	version, err := semver.NewVersion(string(m[1]))
	if err != nil {
		return err
	}
	if version.LessThan(lastpassMinVersion) {
		return fmt.Errorf("lastpass: version %s found, need version %s or later", version, lastpassMinVersion)
	}
	return nil
}

func lastpassParseNote(note string) map[string]string {
	result := make(map[string]string)
	s := bufio.NewScanner(bytes.NewBufferString(note))
	key := ""
	for s.Scan() {
		if m := lastpassParseNoteRegexp.FindStringSubmatch(s.Text()); m != nil {
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
		panic(fmt.Errorf("lastpassParseNote: %w", err))
	}
	return result
}
