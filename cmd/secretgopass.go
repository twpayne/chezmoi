package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sync"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
)

var (
	// chezmoi uses gopass show --password which was added in
	// https://github.com/gopasspw/gopass/commit/8fa13d84e3656cfc4ee6717f5f485c9e471ad996
	// and the first tag containing that commit is v1.6.1.
	gopassMinVersion    = semver.Version{Major: 1, Minor: 6, Patch: 1}
	gopassVersionArgs   = []string{"--version"}
	gopassVersionRegexp = regexp.MustCompile(`gopass\s+(\d+\.\d+\.\d+)`)
)

var gopassCmd = &cobra.Command{
	Use:     "gopass [args...]",
	Short:   "Execute the gopass CLI",
	PreRunE: config.ensureNoError,
	RunE:    config.runSecretGopassCmd,
}

type gopassCmdConfig struct {
	Command          string
	versionCheckOnce sync.Once
}

var gopassCache = make(map[string]string)

func init() {
	secretCmd.AddCommand(gopassCmd)

	config.Gopass.Command = "gopass"
	config.addTemplateFunc("gopass", config.gopassFunc)
}

func (c *Config) runSecretGopassCmd(cmd *cobra.Command, args []string) error {
	return c.run("", c.Gopass.Command, args...)
}

func (c *Config) gopassOutput(args ...string) ([]byte, error) {
	name := c.Gopass.Command
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	output, err := c.mutator.IdempotentCmdOutput(cmd)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Config) gopassFunc(id string) string {
	c.Gopass.versionCheckOnce.Do(func() {
		panicOnError(c.gopassVersionCheck())
	})
	if s, ok := gopassCache[id]; ok {
		return s
	}
	output, err := c.gopassOutput("show", "--password", id)
	panicOnError(err)
	var password string
	if index := bytes.IndexByte(output, '\n'); index != -1 {
		password = string(output[:index])
	} else {
		password = string(output)
	}
	gopassCache[id] = password
	return gopassCache[id]
}

func (c *Config) gopassVersionCheck() error {
	output, err := c.gopassOutput(gopassVersionArgs...)
	if err != nil {
		return err
	}
	m := gopassVersionRegexp.FindSubmatch(output)
	if m == nil {
		return fmt.Errorf("could not extract version from %q", output)
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
