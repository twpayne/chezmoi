package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"unicode"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
	yaml "gopkg.in/yaml.v2"
)

type sourceVCSConfig struct {
	Command string
	Init    interface{}
	Pull    interface{}
}

// A Config represents a configuration.
type Config struct {
	configFile        string
	err               error
	SourceDir         string
	DestDir           string
	Umask             permValue
	DryRun            bool
	Remove            bool
	Verbose           bool
	Color             string
	GPG               chezmoi.GPG
	GPGRecipient      string
	SourceVCS         sourceVCSConfig
	Merge             mergeConfig
	Bitwarden         bitwardenCmdConfig
	GenericSecret     genericSecretCmdConfig
	KeePassXC         keePassXCCmdConfig
	Lastpass          lastpassCmdConfig
	Onepassword       onepasswordCmdConfig
	Vault             vaultCmdConfig
	Pass              passCmdConfig
	Data              map[string]interface{}
	colored           bool
	persistentState   chezmoi.PersistentState
	templateFuncs     template.FuncMap
	add               addCmdConfig
	data              dataCmdConfig
	dump              dumpCmdConfig
	edit              editCmdConfig
	init              initCmdConfig
	_import           importCmdConfig
	keyring           keyringCmdConfig
	remove            removeCmdConfig
	update            updateCmdConfig
	upgrade           upgradeCmdConfig
	stdin             io.Reader
	stdout            io.Writer
	stderr            io.Writer
	bds               *xdg.BaseDirectorySpecification
	scriptStateBucket []byte
}

var (
	formatMap = map[string]func(io.Writer, interface{}) error{
		"json": func(w io.Writer, value interface{}) error {
			e := json.NewEncoder(w)
			e.SetIndent("", "  ")
			return e.Encode(value)
		},
		"toml": func(w io.Writer, value interface{}) error {
			return toml.NewEncoder(w).Encode(value)
		},
		"yaml": func(w io.Writer, value interface{}) error {
			return yaml.NewEncoder(w).Encode(value)
		},
	}

	wellKnownAbbreviations = map[string]struct{}{
		"ANSI": {},
		"CPE":  {},
		"ID":   {},
		"URL":  {},
	}
)

// Stderr returns c's stderr.
func (c *Config) Stderr() io.Writer {
	if c.stderr != nil {
		return c.stderr
	}
	return os.Stderr
}

// Stdin returns c's stdin.
func (c *Config) Stdin() io.Reader {
	if c.stdin != nil {
		return c.stdin
	}
	return os.Stdin
}

// Stdout returns c's stdout.
func (c *Config) Stdout() io.Writer {
	if c.stdout != nil {
		return c.stdout
	}
	return os.Stdout
}

func (c *Config) addTemplateFunc(key string, value interface{}) {
	if c.templateFuncs == nil {
		c.templateFuncs = make(template.FuncMap)
	}
	if _, ok := c.templateFuncs[key]; ok {
		panic(fmt.Sprintf("Config.addTemplateFunc: %s already defined", key))
	}
	c.templateFuncs[key] = value
}

func (c *Config) applyArgs(fs vfs.FS, args []string, mutator chezmoi.Mutator) error {
	fs = vfs.NewReadOnlyFS(fs)
	ts, err := c.getTargetState(fs, nil)
	if err != nil {
		return err
	}
	applyOptions := &chezmoi.ApplyOptions{
		DestDir:           ts.DestDir,
		DryRun:            c.DryRun,
		Ignore:            ts.TargetIgnore.Match,
		PersistentState:   c.persistentState,
		Remove:            c.Remove,
		ScriptStateBucket: c.scriptStateBucket,
		Stdout:            c.Stdout(),
		Umask:             ts.Umask,
		Verbose:           c.Verbose,
	}
	if len(args) == 0 {
		return ts.Apply(fs, mutator, applyOptions)
	}
	entries, err := c.getEntries(fs, ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := entry.Apply(fs, mutator, applyOptions); err != nil {
			return err
		}
	}
	return nil
}

// ensureNoError ensures that no error was encountered when loading c.
func (c *Config) ensureNoError(cmd *cobra.Command, args []string) error {
	if c.err != nil {
		return errors.New("config contains errors, aborting")
	}
	return nil
}

func (c *Config) ensureSourceDirectory(fs chezmoi.PrivacyStater, mutator chezmoi.Mutator) error {
	if err := vfs.MkdirAll(mutator, filepath.Dir(c.SourceDir), 0777&^os.FileMode(c.Umask)); err != nil {
		return err
	}
	info, err := fs.Stat(c.SourceDir)
	switch {
	case err == nil && info.IsDir():
		if !chezmoi.IsPrivate(fs, c.SourceDir, os.FileMode(c.Umask)) {
			if err := mutator.Chmod(c.SourceDir, 0700&^os.FileMode(c.Umask)); err != nil {
				return err
			}
		}
		return nil
	case os.IsNotExist(err):
		return mutator.Mkdir(c.SourceDir, 0700&^os.FileMode(c.Umask))
	case err == nil:
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	default:
		return err
	}
}

func (c *Config) execEditor(argv ...string) error {
	return c.exec(append([]string{c.getEditor()}, argv...))
}

func (c *Config) getDefaultMutator(fs vfs.FS) chezmoi.Mutator {
	var mutator chezmoi.Mutator
	if c.DryRun {
		mutator = chezmoi.NullMutator{}
	} else {
		mutator = chezmoi.NewFSMutator(fs)
	}
	if c.Verbose {
		mutator = chezmoi.NewLoggingMutator(c.Stdout(), mutator, c.colored)
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

func (c *Config) getEntries(fs vfs.Stater, ts *chezmoi.TargetState, args []string) ([]chezmoi.Entry, error) {
	entries := []chezmoi.Entry{}
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		entry, err := ts.Get(fs, targetPath)
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

func (c *Config) getTargetState(fs vfs.FS, populateOptions *chezmoi.PopulateOptions) (*chezmoi.TargetState, error) {
	fs = vfs.NewReadOnlyFS(fs)
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

	destDir := c.DestDir
	if destDir != "" {
		destDir, err = filepath.Abs(c.DestDir)
		if err != nil {
			return nil, err
		}
	}

	// For backwards compatibility, prioritize gpgRecipient over gpg.recipient.
	if c.GPGRecipient != "" {
		c.GPG.Recipient = c.GPGRecipient
	}

	ts := chezmoi.NewTargetState(destDir, os.FileMode(c.Umask), c.SourceDir, data, c.templateFuncs, &c.GPG)
	if err := ts.Populate(fs, populateOptions); err != nil {
		return nil, err
	}
	if Version != nil && ts.MinVersion != nil && Version.LessThan(*ts.MinVersion) {
		return nil, fmt.Errorf("chezmoi version %s too old, source state requires at least %s", Version, ts.MinVersion)
	}
	return ts, nil
}

func (c *Config) getVCSInfo() (*vcsInfo, error) {
	vcsInfo, ok := vcsInfos[filepath.Base(c.SourceVCS.Command)]
	if !ok {
		return nil, fmt.Errorf("%s: unsupported source VCS command", c.SourceVCS.Command)
	}
	return vcsInfo, nil
}

//nolint:unparam
func (c *Config) prompt(s, choices string) (byte, error) {
	r := bufio.NewReader(c.Stdin())
	for {
		_, err := fmt.Printf("%s [%s]? ", s, strings.Join(strings.Split(choices, ""), ","))
		if err != nil {
			return 0, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) == 1 && strings.IndexByte(choices, line[0]) != -1 {
			return line[0], nil
		}
	}
}

// run runs name argv... in dir.
func (c *Config) run(dir, name string, argv ...string) error {
	if c.Verbose {
		if dir == "" {
			fmt.Printf("%s %s\n", name, strings.Join(argv, " "))
		} else {
			fmt.Printf("( cd %s && %s %s )\n", dir, name, strings.Join(argv, " "))
		}
	}
	if c.DryRun {
		return nil
	}
	cmd := exec.Command(name, argv...)
	cmd.Dir = dir
	cmd.Stdin = c.Stdin()
	cmd.Stdout = c.Stdout()
	cmd.Stderr = c.Stdout()
	return cmd.Run()
}

func (c *Config) runEditor(argv ...string) error {
	return c.run("", c.getEditor(), argv...)
}

func (c *Config) warn(s string) {
	fmt.Fprintf(c.Stderr(), "warning: %s\n", s)
}

func getDefaultConfigFile(bds *xdg.BaseDirectorySpecification) string {
	// Search XDG Base Directory Specification config directories first.
	for _, configDir := range bds.ConfigDirs {
		for _, extension := range viper.SupportedExts {
			configFilePath := filepath.Join(configDir, "chezmoi", "chezmoi."+extension)
			if _, err := os.Stat(configFilePath); err == nil {
				return configFilePath
			}
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return filepath.Join(bds.ConfigHome, "chezmoi", "chezmoi.toml")
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
	} else if cgoEnabled && runtime.GOOS != "windows" {
		// Only return an error if CGO is enabled and the platform is
		// non-Windows (groups don't really mean much on Windows).
		return nil, err
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	data["homedir"] = homedir

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	data["fullHostname"] = hostname
	data["hostname"] = strings.SplitN(hostname, ".", 2)[0]

	osRelease, err := getOSRelease(fs)
	if err == nil {
		data["osRelease"] = upperSnakeCaseToCamelCaseMap(osRelease)
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return data, nil
}

func getDefaultSourceDir(bds *xdg.BaseDirectorySpecification) string {
	// Check for XDG Base Directory Specification data directories first.
	for _, dataDir := range bds.DataDirs {
		sourceDir := filepath.Join(dataDir, "chezmoi")
		if _, err := os.Stat(sourceDir); err == nil {
			return sourceDir
		}
	}
	// Fallback to XDG Base Directory Specification default.
	return filepath.Join(bds.DataHome, "chezmoi")
}

func getPersistentStateFile(bds *xdg.BaseDirectorySpecification, configFile string) string {
	for _, configDir := range bds.ConfigDirs {
		persistentStateFile := filepath.Join(configDir, "chezmoi", "chezmoistate.boltdb")
		if _, err := os.Stat(persistentStateFile); err == nil {
			return persistentStateFile
		}
	}
	return filepath.Join(filepath.Dir(configFile), "chezmoistate.boltdb")
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

func makeRunE(runCmd func(vfs.FS, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runCmd(vfs.OSFS, args)
	}
}

func printErrorAndExit(err error) {
	fmt.Printf("chezmoi: %v\n", err)
	os.Exit(1)
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
