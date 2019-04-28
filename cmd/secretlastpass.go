package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
)

var lastpassCmd = &cobra.Command{
	Use:     "lastpass [args...]",
	Short:   "Execute the LastPass CLI (lpass)",
	PreRunE: config.ensureNoError,
	RunE:    makeRunE(config.runLastpassCmd),
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

func (c *Config) runLastpassCmd(fs vfs.FS, args []string) error {
	return c.exec(append([]string{c.Lastpass.Command}, args...))
}

func (c *Config) lastpassOutput(args ...string) ([]byte, error) {
	name := c.Lastpass.Command
	if c.Verbose {
		fmt.Printf("%s %s\n", name, strings.Join(args, " "))
	}
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("lastpass: %s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
	return output, nil
}

func (c *Config) lastpassFunc(id string) interface{} {
	c.Lastpass.versionCheckOnce.Do(func() {
		if err := c.lastpassVersionCheck(); err != nil {
			chezmoi.ReturnTemplateFuncError(err)
		}
	})
	if data, ok := lastPassCache[id]; ok {
		return data
	}
	output, err := c.lastpassOutput("show", "--json", id)
	if err != nil {
		chezmoi.ReturnTemplateFuncError(err)
	}
	var data []map[string]interface{}
	if err := json.Unmarshal(output, &data); err != nil {
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("lastpass: parse error: %v\n%q", err, output))
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
		chezmoi.ReturnTemplateFuncError(fmt.Errorf("lastpassParseNote: %v", err))
	}
	return result
}
