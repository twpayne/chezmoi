package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"
	"unicode"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg"
	yaml "gopkg.in/yaml.v2"
)

// A Config represents a configuration.
type Config struct {
	SourceDir        string
	DestDir          string
	Umask            octalIntValue
	DryRun           bool
	Verbose          bool
	SourceVCSCommand string
	Bitwarden        bitwardenCommandConfig
	LastPass         LastPassCommandConfig
	Data             map[string]interface{}
	funcs            template.FuncMap
	add              addCommandConfig
	data             dataCommandConfig
	dump             dumpCommandConfig
	edit             editCommandConfig
	_import          importCommandConfig
	keyring          keyringCommandConfig
}

var (
	formatMap = map[string]func(io.Writer, interface{}) error{
		"json": func(w io.Writer, value interface{}) error {
			e := json.NewEncoder(w)
			e.SetIndent("", "  ")
			return e.Encode(value)
		},
		"yaml": func(w io.Writer, value interface{}) error {
			return yaml.NewEncoder(w).Encode(value)
		},
	}

	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": struct{}{},
		"CPE":  struct{}{},
		"URL":  struct{}{},
	}
)

func (c *Config) addFunc(key string, value interface{}) {
	if c.funcs == nil {
		c.funcs = make(template.FuncMap)
	}
	if _, ok := c.funcs[key]; ok {
		panic(fmt.Sprintf("Config.addFunc: %s already defined", key))
	}
	c.funcs[key] = value
}

func (c *Config) applyArgs(fs vfs.FS, args []string, mutator chezmoi.Mutator) error {
	ts, err := c.getTargetState(fs)
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return ts.Apply(fs, mutator)
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := entry.Apply(fs, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, mutator); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) exec(argv []string) error {
	path, err := exec.LookPath(argv[0])
	if err != nil {
		return err
	}
	if c.Verbose {
		fmt.Printf("exec %s\n", strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	return syscall.Exec(path, argv, os.Environ())
}

func (c *Config) execEditor(argv ...string) error {
	return c.exec(append([]string{c.getEditor()}, argv...))
}

func (c *Config) getDefaultMutator(fs vfs.FS) chezmoi.Mutator {
	var mutator chezmoi.Mutator
	if c.DryRun {
		mutator = chezmoi.NullMutator
	} else {
		mutator = chezmoi.NewFSMutator(fs, c.DestDir)
	}
	if c.Verbose {
		mutator = chezmoi.NewLoggingMutator(os.Stdout, mutator)
	}
	return mutator
}

func (c *Config) getEditor() string {
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "vi"
}

func (c *Config) getEntries(ts *chezmoi.TargetState, args []string) ([]chezmoi.Entry, error) {
	entries := []chezmoi.Entry{}
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		entry, err := ts.Get(targetPath)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, fmt.Errorf("%s: not in source state", arg)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (c *Config) getTargetState(fs vfs.FS) (*chezmoi.TargetState, error) {
	defaultData, err := getDefaultData(fs)
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"chezmoi": defaultData,
	}
	for key, value := range c.Data {
		data[key] = value
	}
	ts := chezmoi.NewTargetState(c.DestDir, os.FileMode(c.Umask), c.SourceDir, data, c.funcs)
	readOnlyFS := vfs.NewReadOnlyFS(fs)
	if err := ts.Populate(readOnlyFS); err != nil {
		return nil, err
	}
	return ts, nil
}

func (c *Config) runEditor(argv ...string) error {
	editor := c.getEditor()
	if c.Verbose {
		fmt.Printf("%s %s\n", editor, strings.Join(argv, " "))
	}
	if c.DryRun {
		return nil
	}
	cmd := exec.Command(editor, argv...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getDefaultConfigFile(x *xdg.XDG, homeDir string) string {
	// Search XDG config directories first.
	for _, configDir := range x.ConfigDirs {
		for _, extension := range viper.SupportedExts {
			configFilePath := filepath.Join(configDir, "chezmoi", "chezmoi."+extension)
			if _, err := os.Stat(configFilePath); err == nil {
				return configFilePath
			}
		}
	}
	// Fallback to XDG default.
	return filepath.Join(x.ConfigHome, "chezmoi", "chezmoi.yaml")
}

func getDefaultData(fs vfs.FS) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"arch": runtime.GOARCH,
		"os":   runtime.GOOS,
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, err
	}
	data["username"] = currentUser.Username

	// user.LookupGroupId looks up a group by gid. If CGO is enabled, then this
	// uses an underlying C library call (e.g. getgrgid_r on Linux) and is
	// trustworthy. If CGO is disabled then the fallback implementation only
	// searches /etc/group, which is typically empty if an external directory
	// service is being used, and so the lookup fails. So, if
	// user.LookupGroupId returns an error, only return an error if CGO is
	// enabled.
	group, err := user.LookupGroupId(currentUser.Gid)
	if err == nil {
		data["group"] = group.Name
	} else if err != nil && cgoEnabled {
		// Only return an error if CGO is enabled.
		return nil, err
	}

	homedir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	data["homedir"] = homedir

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	data["hostname"] = hostname

	osRelease, err := getOSRelease(fs)
	if err == nil {
		data["osRelease"] = upperSnakeCaseToCamelCaseMap(osRelease)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return data, nil
}

func getDefaultSourceDir(x *xdg.XDG, homeDir string) string {
	// Check for XDG data directories first.
	for _, dataDir := range x.DataDirs {
		sourceDir := filepath.Join(dataDir, "chezmoi")
		if _, err := os.Stat(sourceDir); err == nil {
			return sourceDir
		}
	}
	// Fallback to XDG default.
	return filepath.Join(x.DataHome, "chezmoi")
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

func makeRunE(runCommand func(vfs.FS, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runCommand(vfs.OSFS, args)
	}
}

func printErrorAndExit(err error) {
	fmt.Printf("chezmoi: %v\n", err)
	os.Exit(1)
}

func prompt(s, choices string) (byte, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		_, err := fmt.Printf("%s [%s]? ", s, strings.Join(strings.Split(choices, ""), ","))
		if err != nil {
			return 0, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		if len(line) == 2 && strings.IndexByte(choices, line[0]) != -1 {
			return line[0], nil
		}
	}
}

// titilize returns s, titilized.
func titilize(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	return string(append([]rune{unicode.ToTitle(runes[0])}, runes[1:]...))
}

// upperSnakeCaseToCamelCase converts a string in UPPER_SNAKE_CASE to
// camelCase.
func upperSnakeCaseToCamelCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else if !isWellKnownAbbreviation(word) {
			words[i] = titilize(strings.ToLower(word))
		}
	}
	return strings.Join(words, "")
}

// upperSnakeCaseToCamelCaseKeys returns m with all keys converted from
// UPPER_SNAKE_CASE to camelCase.
func upperSnakeCaseToCamelCaseMap(m map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[upperSnakeCaseToCamelCase(k)] = v
	}
	return result
}
