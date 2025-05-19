package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/twpayne/go-expect"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

type keepassxcMode string

const (
	keepassxcModeBuiltin       keepassxcMode = "builtin"
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
	promptStr       string
	cache           map[string]map[string]string
	attachmentCache map[string]map[string]string
	attributeCache  map[keepassxcAttributeCacheKey]string
	password        string
}

var (
	keepassxcMinVersion = semver.Version{Major: 2, Minor: 7, Patch: 0}

	keepassxcEnterPasswordToUnlockDatabaseRx             = regexp.MustCompile(`^Enter password to unlock .*: `)
	keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx = regexp.MustCompile(
		"^Please present or touch your \\S+ to continue\\.\r\n",
	)
	keepassxcAnyResponseRx = regexp.MustCompile(`(?m)\A.*\r\n`)
	keepassxcPairRx        = regexp.MustCompile(`^([A-Z]\w*):\s*(.*)$`)
	keepassxcPromptRx      = regexp.MustCompile(`^.*> `)
)

func (c *Config) keepassxcAttachmentTemplateFunc(entry, name string) string {
	if data, ok := c.Keepassxc.attachmentCache[entry][name]; ok {
		return data
	}

	switch c.Keepassxc.Mode {
	case keepassxcModeCachePassword:
		// In cache password mode use --stdout to read the attachment data directly.
		return string(mustValue(c.keepassxcOutput("attachment-export", "--quiet", "--stdout", entry, name)))
	case keepassxcModeOpen:
		// In open mode write the attachment data to a temporary file.
		tempDir := mustValue(c.tempDir("chezmoi-keepassxc"))
		tempFilename := tempDir.JoinString("attachment-" + strconv.FormatInt(time.Now().UnixNano(), 10)).String()
		_ = mustValue(c.keepassxcOutputOpen("attachment-export", "--quiet", entry, name, tempFilename))
		data := mustValue(os.ReadFile(tempFilename))
		must(os.Remove(tempFilename))
		return string(data)
	case keepassxcModeBuiltin:
		c.Keepassxc.cache[entry] = c.keepassxcBuiltinExtractValues(entry, keepassxcBuiltinMapAttachmentCache)
		if data, ok := c.Keepassxc.cache[entry]; ok {
			if att, ok := data[name]; ok {
				return att
			}
		}
		panic(fmt.Sprintf("attachment %s of entry %s not found", name, entry))
	default:
		panic(fmt.Sprintf("%s: invalid mode", c.Keepassxc.Mode))
	}
}

func (c *Config) keepassxcTemplateFunc(entry string) map[string]string {
	if c.Keepassxc.cache == nil {
		c.Keepassxc.cache = make(map[string]map[string]string)
	}

	if data, ok := c.Keepassxc.cache[entry]; ok {
		return data
	}

	if c.Keepassxc.Mode == keepassxcModeBuiltin {
		c.Keepassxc.cache[entry] = c.keepassxcBuiltinExtractValues(entry, keepassxcBuiltinMapValueCache)
		return c.Keepassxc.cache[entry]
	}

	args := []string{"--quiet", "--show-protected", entry}
	output := mustValue(c.keepassxcOutput("show", args...))

	data := mustValue(keepassxcParseOutput(output))

	c.Keepassxc.cache[entry] = data

	return data
}

func (c *Config) keepassxcAttributeTemplateFunc(entry, attribute string) string {
	if c.Keepassxc.Mode == keepassxcModeBuiltin {
		// builtin stores attributes in cache
		if c.Keepassxc.cache == nil {
			c.Keepassxc.cache = make(map[string]map[string]string)
		}
		if data, ok := c.Keepassxc.cache[entry]; ok {
			return data[attribute]
		}
		c.Keepassxc.cache[entry] = c.keepassxcBuiltinExtractValues(entry, keepassxcBuiltinMapValueCache)
		if data, ok := c.Keepassxc.cache[entry]; ok {
			return data[attribute]
		}
		panic(fmt.Sprintf("attribute %s of entry %s not found", entry, attribute))
	}

	key := keepassxcAttributeCacheKey{
		entry:     entry,
		attribute: attribute,
	}
	if data, ok := c.Keepassxc.attributeCache[key]; ok {
		return data
	}

	output := mustValue(c.keepassxcOutput("show", entry, "--attributes", attribute, "--quiet", "--show-protected"))
	outputStr := string(bytes.TrimSpace(output))
	if c.Keepassxc.attributeCache == nil {
		c.Keepassxc.attributeCache = make(map[keepassxcAttributeCacheKey]string)
	}
	c.Keepassxc.attributeCache[key] = outputStr

	return outputStr
}

// keepassxcOutput returns the output of command and args.
func (c *Config) keepassxcOutput(command string, args ...string) ([]byte, error) {
	if c.Keepassxc.Database.IsEmpty() {
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
	cmdArgs := []string{command}
	cmdArgs = append(cmdArgs, c.Keepassxc.Args...)
	cmdArgs = append(cmdArgs, c.Keepassxc.Database.String())
	cmdArgs = append(cmdArgs, args...)
	cmd := exec.Command(c.Keepassxc.Command, cmdArgs...)
	if c.Keepassxc.password == "" && c.Keepassxc.Prompt {
		password, err := c.readPassword(fmt.Sprintf("Enter password to unlock %s: ", c.Keepassxc.Database), "password")
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

	output, err := chezmoilog.LogCmdOutput(c.logger, cmd)
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
		cmdArgs := []string{"open"}
		cmdArgs = append(cmdArgs, c.Keepassxc.Args...)
		cmdArgs = append(cmdArgs, c.Keepassxc.Database.String())
		cmd := exec.Command(c.Keepassxc.Command, cmdArgs...)
		env := os.Environ()
		// Ensure prompt is in English.
		env = append(env, "LANGUAGE=en")
		// Reduce injection of terminal control characters.
		env = slices.DeleteFunc(env, func(s string) bool {
			return strings.HasPrefix(s, "TERM=")
		})
		cmd.Env = env
		cmd.Stdin = console.Tty()
		cmd.Stdout = console.Tty()
		cmd.Stderr = console.Tty()
		if err := chezmoilog.LogCmdStart(c.logger, cmd); err != nil {
			return nil, err
		}

		if c.Keepassxc.Prompt {
			// Expect the password or YubiKey response.
			response, err := console.Expect(
				expect.Regexp(keepassxcEnterPasswordToUnlockDatabaseRx),
				expect.Regexp(keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx),
				expect.Regexp(keepassxcAnyResponseRx),
			)
			if err != nil {
				return nil, err
			}

			switch {
			case keepassxcEnterPasswordToUnlockDatabaseRx.MatchString(response):
				// Read the password from the user, if necessary.
				var password string
				if c.Keepassxc.password != "" {
					password = c.Keepassxc.password
				} else {
					password, err = c.readPassword(response, "password")
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
			case keepassxcPleasePresentOrTouchYourYubiKeyToContinueRx.MatchString(response):
				if _, err := console.ExpectString("\r\n"); err != nil {
					return nil, err
				}
				if _, err := c.stderr.Write([]byte(strings.TrimSpace(response) + "\n")); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("%q: unexpected response", response)
			}
		}

		// Read the response, e.g "Passwords> ", so we can expect it later.
		response, err := console.Expect(
			expect.Regexp(keepassxcPromptRx),
			expect.Regexp(keepassxcAnyResponseRx),
		)
		if err != nil {
			return nil, err
		}
		if !keepassxcPromptRx.MatchString(response) {
			return nil, fmt.Errorf("%q: unexpected response", response)
		}

		c.Keepassxc.cmd = cmd
		c.Keepassxc.console = console
		c.Keepassxc.promptStr = keepassxcPromptRx.FindString(response)
	}

	// Build the command line. Strings with spaces and other non-word characters
	// need to be quoted.
	quotedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		quotedArgs = append(quotedArgs, maybeQuote(arg))
	}
	line := strings.Join(append([]string{command}, quotedArgs...), " ")

	// Send the line.
	if _, err := c.Keepassxc.console.SendLine(line); err != nil {
		return nil, err
	}

	// Read everything up to and including the prompt.
	output, err := c.Keepassxc.console.ExpectString(c.Keepassxc.promptStr)
	if err != nil {
		return nil, err
	}
	outputLines := strings.Split(output, "\r\n")

	// Trim the echoed command from the output, which is the first line.
	if len(outputLines) > 0 {
		outputLines = outputLines[1:]
	}

	// Trim the prompt from the output, which is the last line.
	if len(outputLines) > 0 {
		outputLines = outputLines[:len(outputLines)-1]
	}

	return []byte(strings.Join(outputLines, "\r\n")), nil
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
	if err := chezmoilog.LogCmdWait(c.logger, c.Keepassxc.cmd); err != nil {
		return err
	}
	return c.Keepassxc.console.Close()
}

// keepassxcBuiltinExtractValues extract builtin values.
func (c *Config) keepassxcBuiltinExtractValues(
	entry string,
	mapper func(db *gokeepasslib.Database, entry gokeepasslib.Entry) map[string]string,
) map[string]string {
	file, err := os.Open(c.Keepassxc.Database.String())
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if c.Keepassxc.password == "" && c.Keepassxc.Prompt {
		password, err := c.readPassword(fmt.Sprintf("Enter password to unlock %s: ", c.Keepassxc.Database), "password")
		if err != nil {
			panic(err)
		}
		c.Keepassxc.password = password
	}

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(c.Keepassxc.password)
	err = gokeepasslib.NewDecoder(file).Decode(db)
	if err != nil {
		panic(err)
	}
	err = db.UnlockProtectedEntries()
	if err != nil {
		panic(err)
	}
	return keepassxcBuiltinBuildCache(db, "", db.Content.Root.Groups[0].Groups, nil, entry, mapper)
}

// keepassxcBuiltinBuildCache build the builtin cache using a given mapper function.
func keepassxcBuiltinBuildCache(
	db *gokeepasslib.Database,
	path string,
	groups []gokeepasslib.Group,
	mapData map[string]string,
	entry string,
	mapper func(db *gokeepasslib.Database, entry gokeepasslib.Entry) map[string]string,
) map[string]string {
	for _, group := range groups {
		if len(group.Entries) > 0 {
			for _, groupEntry := range group.Entries {
				var elements []string
				if path != "" {
					elements = append(elements, path)
				}
				elements = append(elements, group.Name, groupEntry.GetTitle())
				key := strings.Join(elements, "/")
				if entry == key {
					return mapper(db, groupEntry)
				}
			}
		}
		if len(group.Groups) > 0 {
			mapData = keepassxcBuiltinBuildCache(
				db,
				strings.TrimPrefix(fmt.Sprintf("%s/%s", path, group.Name), "/"),
				group.Groups,
				mapData,
				entry,
				mapper,
			)
			if len(mapData) > 0 {
				return mapData
			}
		}
	}
	return map[string]string{}
}

// keepassxcBuiltinMapValueCache map builtin entries for values and attributes.
func keepassxcBuiltinMapValueCache(_ *gokeepasslib.Database, entry gokeepasslib.Entry) map[string]string {
	m := make(map[string]string)
	for _, value := range entry.Values {
		m[value.Key] = value.Value.Content
	}
	return m
}

// keepassxcBuiltinMapAttachmentCache map builtin entries for attachments.
func keepassxcBuiltinMapAttachmentCache(db *gokeepasslib.Database, entry gokeepasslib.Entry) map[string]string {
	m := make(map[string]string)
	for _, bin := range entry.Binaries {
		b := db.FindBinary(bin.Value.ID)
		if b != nil {
			str, err := b.GetContentString()
			if err != nil {
				panic(err)
			}
			m[bin.Name] = str
		}
	}
	return m
}
