package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/pete-woods/go-expect"
	"github.com/tobischo/gokeepasslib/v3"

	"chezmoi.io/chezmoi/v2/internal/chezmoi"
	"chezmoi.io/chezmoi/v2/internal/chezmoilog"
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
	promptStr       string //nolint:unused
	cache           map[string]map[string]string
	attachmentCache map[string]map[string]string
	attributeCache  map[keepassxcAttributeCacheKey]string
	password        string
}

var (
	keepassxcMinVersion = semver.Version{Major: 2, Minor: 7, Patch: 0}
	keepassxcPairRx     = regexp.MustCompile(`\A([A-Z]\w*):\s*(.*?)\r?\n\z`)
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

	data := keepassxcParseOutput(output)

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

// keepassxcParseOutput parses a list of key-value pairs.
func keepassxcParseOutput(output []byte) map[string]string {
	data := make(map[string]string)
	var key string
	for line := range bytes.Lines(output) {
		switch match := keepassxcPairRx.FindSubmatch(line); {
		case match != nil:
			key = string(match[1])
			data[key] = string(match[2])
		case key != "":
			data[key] += "\n" + string(bytes.TrimSuffix(line, []byte{'\n'}))
		}
	}
	return data
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

// keepassxcBuiltinBuildCache build the builtin cache using a given mapper
// function.
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
