package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Netflix/go-expect"
	"github.com/coreos/go-semver/semver"
	"golang.org/x/exp/slices"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

type keepassxcMode string

const (
	keepassxcModeCachePassword keepassxcMode = "cache-password"
	keepassxcModeOpen          keepassxcMode = "open"
)

type keepassxcAttributeCacheKey struct {
	entry     string
	attribute string
}

type keepassxcConfig struct {
	Command         string          `json:"command"  mapstructure:"command"  yaml:"command"`
	Database        chezmoi.AbsPath `json:"database" mapstructure:"database" yaml:"database"`
	Mode            keepassxcMode   `json:"mode"     mapstructure:"mode"     yaml:"mode"`
	Args            []string        `json:"args"     mapstructure:"args"     yaml:"args"`
	Prompt          bool            `json:"prompt"   mapstructure:"prompt"   yaml:"prompt"`
	cmd             *exec.Cmd
	console         *expect.Console
	prompt          string
	cache           map[string]map[string]string
	attachmentCache map[string]map[string]string
	attributeCache  map[keepassxcAttributeCacheKey]string
	password        string
}

var (
	keepassxcMinVersion = semver.Version{Major: 2, Minor: 7, Patch: 0}

	keepassxcEnterPasswordToUnlockDatabaseRx = regexp.MustCompile(`^Enter password to unlock .*: `)
	keepassxcPairRx                          = regexp.MustCompile(`^([A-Z]\w*):\s*(.*)$`)
	keepassxcPromptRx                        = regexp.MustCompile(`^.*> `)
)

func (c *Config) keepassxcAttachmentTemplateFunc(entry, name string) string {
	if data, ok := c.Keepassxc.attachmentCache[entry][name]; ok {
		return data
	}

	switch c.Keepassxc.Mode {
	case keepassxcModeCachePassword:
		// In cache password mode use --stdout to read the attachment data directly.
		output, err := c.keepassxcOutput("attachment-export", "--quiet", "--stdout", entry, name)
		if err != nil {
			panic(err)
		}
		return string(output)
	case keepassxcModeOpen:
		// In open mode write the attachment data to a temporary file.
		tempDir, err := c.tempDir("chezmoi-keepassxc")
		if err != nil {
			panic(err)
		}
		tempFilename := tempDir.JoinString(name).String()
		if _, err := c.keepassxcOutputOpen("attachment-export", "--quiet", entry, name, tempFilename); err != nil {
			panic(err)
		}
		data, err := os.ReadFile(tempFilename)
		if err != nil {
			panic(err)
		}
		if err := os.Remove(tempFilename); err != nil {
			panic(err)
		}
		return string(data)
	default:
		panic(fmt.Sprintf("%s: invalid mode", c.Keepassxc.Mode))
	}
}

func (c *Config) keepassxcTemplateFunc(entry string) map[string]string {
	if data, ok := c.Keepassxc.cache[entry]; ok {
		return data
	}

	command := "show"
	args := []string{"--quiet", "--show-protected", entry}
	output, err := c.keepassxcOutput("show", args...)
	if err != nil {
		panic(err)
	}

	data, err := keepassxcParseOutput(output)
	if err != nil {
		// FIXME the error below should vary depending on the mode
		panic(newParseCmdOutputError(command, args, output, err))
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

	output, err := c.keepassxcOutput("show", entry, "--attributes", attribute, "--quiet", "--show-protected")
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

// keepassxcOutputCachePassword returns the output of command and args.
func (c *Config) keepassxcOutput(command string, args ...string) ([]byte, error) {
	if c.Keepassxc.Database.Empty() {
		panic(errors.New("keepassxc.database not set"))
	}

	switch c.Keepassxc.Mode {
	case keepassxcModeCachePassword:
		return c.keepassxcOutputCachePassword(command, args...)
	case keepassxcModeOpen:
		return c.keepassxcOutputOpen(command, args...)
	default:
		panic(fmt.Sprintf("%s: invalid mode", c.Keepassxc.Mode))
	}
}

// keepassxcOutputCachePassword returns the output of command and args,
// prompting the user for the password and caching it for later use.
func (c *Config) keepassxcOutputCachePassword(command string, args ...string) ([]byte, error) {
	cmdArgs := append([]string{command, c.Keepassxc.Database.String()}, args...)
	cmd := exec.Command(c.Keepassxc.Command, cmdArgs...) //nolint:gosec
	if c.Keepassxc.password == "" && c.Keepassxc.Prompt {
		password, err := c.readPassword(fmt.Sprintf("Enter password to unlock %s: ", c.Keepassxc.Database))
		if err != nil {
			return nil, err
		}
		c.Keepassxc.password = password
	}
	if c.Keepassxc.password != "" {
		cmd.Stdin = bytes.NewBufferString(c.Keepassxc.password + "\n")
	} else {
		cmd.Stdin = os.Stdin
	}
	cmd.Stderr = os.Stderr

	output, err := chezmoilog.LogCmdOutput(cmd)
	if err != nil {
		return nil, newCmdOutputError(cmd, output, err)
	}
	return output, nil
}

// keepassxcOutputOpen returns the output of command and args using an
// interactive connection via keepassxc-cli open command's interactive console.
func (c *Config) keepassxcOutputOpen(command string, args ...string) ([]byte, error) {
	// Create the console if it is not already created.
	if c.Keepassxc.console == nil {
		// Create the console.
		console, err := expect.NewConsole()
		if err != nil {
			return nil, err
		}

		// Start the keepassxc-cli open command.
		cmdArgs := append(slices.Clone(c.Keepassxc.Args), "open", c.Keepassxc.Database.String())
		cmd := exec.Command(c.Keepassxc.Command, cmdArgs...) //nolint:gosec
		cmd.Stdin = console.Tty()
		cmd.Stdout = console.Tty()
		cmd.Stderr = console.Tty()
		if err := chezmoilog.LogCmdStart(cmd); err != nil {
			return nil, err
		}

		if c.Keepassxc.Prompt {
			// Expect the password prompt, e.g. "Enter password to unlock $HOME/Passwords.kdbx: ".
			enterPasswordToUnlockPrompt, err := console.Expect(expect.Regexp(keepassxcEnterPasswordToUnlockDatabaseRx))
			if err != nil {
				return nil, err
			}

			// Read the password from the user, if necessary.
			var password string
			if c.Keepassxc.password != "" {
				password = c.Keepassxc.password
			} else {
				password, err = c.readPassword(enterPasswordToUnlockPrompt)
				if err != nil {
					return nil, err
				}
			}

			// Send the password.
			if _, err := console.SendLine(password); err != nil {
				return nil, err
			}

			// Wait for the end of the password prompt.
			if _, err := console.ExpectString("\r\n"); err != nil {
				return nil, err
			}
		}

		// Read the prompt, e.g "Passwords> ", so we can expect it later.
		output, err := console.Expect(expect.Regexp(keepassxcPromptRx))
		if err != nil {
			return nil, err
		}

		c.Keepassxc.cmd = cmd
		c.Keepassxc.console = console
		c.Keepassxc.prompt = keepassxcPromptRx.FindString(output)
	}

	// Send the command.
	line := strings.Join(append([]string{command}, args...), " ")
	if _, err := c.Keepassxc.console.SendLine(line); err != nil {
		return nil, err
	}

	// Read everything up to and including the prompt.
	output, err := c.Keepassxc.console.ExpectString(c.Keepassxc.prompt)
	if err != nil {
		return nil, err
	}

	// Trim the command from the output.
	output = strings.TrimPrefix(output, line+"\r\n")

	// Trim the prompt from the output.
	output = strings.TrimSuffix(output, c.Keepassxc.prompt)

	return []byte(output), nil
}

// keepassxcParseOutput parses a list of key-value pairs.
func keepassxcParseOutput(output []byte) (map[string]string, error) {
	data := make(map[string]string)
	s := bufio.NewScanner(bytes.NewReader(output))
	var key string
	for i := 0; s.Scan(); i++ {
		switch match := keepassxcPairRx.FindStringSubmatch(s.Text()); {
		case match != nil:
			key = match[1]
			data[key] = match[2]
		case key != "":
			data[key] += "\n" + s.Text()
		}
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// keepassxcClose closes any open connection to keepassxc-cli.
func (c *Config) keepassxcClose() error {
	// FIXME should we wait for EOF somewhere?
	if c.Keepassxc.console == nil {
		return nil
	}
	if _, err := c.Keepassxc.console.SendLine("exit"); err != nil {
		return err
	}
	if _, err := c.Keepassxc.console.ExpectString("exit\r\n"); err != nil {
		return err
	}
	if err := chezmoilog.LogCmdWait(c.Keepassxc.cmd); err != nil {
		return err
	}
	if err := c.Keepassxc.console.Close(); err != nil {
		return err
	}
	return nil
}
