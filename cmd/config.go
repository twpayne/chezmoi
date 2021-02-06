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
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-semver/semver"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	vfs "github.com/twpayne/go-vfs"
	xdg "github.com/twpayne/go-xdg/v3"
	bolt "go.etcd.io/bbolt"
	yaml "gopkg.in/yaml.v2"

	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/git"
)

const commitMessageTemplateAsset = "assets/templates/COMMIT_MESSAGE.tmpl"

var whitespaceRegexp = regexp.MustCompile(`\s+`)

type sourceVCSConfig struct {
	Command    string
	AutoCommit bool
	AutoPush   bool
	Init       interface{}
	NotGit     bool
	Pull       interface{}
}

type templateConfig struct {
	Options []string
}

// A Config represents a configuration.
type Config struct {
	configFile        string
	err               error
	fs                vfs.FS
	mutator           chezmoi.Mutator
	SourceDir         string
	DestDir           string
	Umask             permValue
	DryRun            bool
	Follow            bool
	Remove            bool
	Verbose           bool
	Color             string
	Debug             bool
	NoScript          bool
	GPG               chezmoi.GPG
	GPGRecipient      string
	SourceVCS         sourceVCSConfig
	Template          templateConfig
	Merge             mergeConfig
	Bitwarden         bitwardenCmdConfig
	CD                cdCmdConfig
	Diff              diffCmdConfig
	GenericSecret     genericSecretCmdConfig
	Gopass            gopassCmdConfig
	KeePassXC         keePassXCCmdConfig
	Lastpass          lastpassCmdConfig
	Onepassword       onepasswordCmdConfig
	Vault             vaultCmdConfig
	Pass              passCmdConfig
	Data              map[string]interface{}
	colored           bool
	maxDiffDataSize   int
	templateFuncs     template.FuncMap
	add               addCmdConfig
	archive           archiveCmdConfig
	completion        completionCmdConfig
	data              dataCmdConfig
	dump              dumpCmdConfig
	edit              editCmdConfig
	executeTemplate   executeTemplateCmdConfig
	_import           importCmdConfig
	init              initCmdConfig
	keyring           keyringCmdConfig
	managed           managedCmdConfig
	purge             purgeCmdConfig
	remove            removeCmdConfig
	update            updateCmdConfig
	upgrade           upgradeCmdConfig
	Stdin             io.Reader
	Stdout            io.Writer
	Stderr            io.Writer
	bds               *xdg.BaseDirectorySpecification
	scriptStateBucket []byte

	//nolint:structcheck,unused
	ioregData ioregData
}

// A configOption sets an option on a Config.
type configOption func(*Config)

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

	identifierRegexp = regexp.MustCompile(`\A[\pL_][\pL\p{Nd}_]*\z`)

	assets = make(map[string][]byte)
)

// newConfig creates a new Config with the given options.
func newConfig(options ...configOption) *Config {
	c := &Config{
		Umask: permValue(chezmoi.GetUmask()),
		Color: "auto",
		SourceVCS: sourceVCSConfig{
			Command: "git",
		},
		Template: templateConfig{
			Options: chezmoi.DefaultTemplateOptions,
		},
		Diff: diffCmdConfig{
			Format: "chezmoi",
		},
		Merge: mergeConfig{
			Command: "vimdiff",
		},
		GPG: chezmoi.GPG{
			Command: "gpg",
		},
		maxDiffDataSize:   1 * 1024 * 1024, // 1MB
		templateFuncs:     sprig.TxtFuncMap(),
		scriptStateBucket: []byte("script"),
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
	}
	for _, option := range options {
		option(c)
	}
	return c
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

func (c *Config) applyArgs(args []string, persistentState chezmoi.PersistentState) error {
	fs := vfs.NewReadOnlyFS(c.fs)
	ts, err := c.getTargetState(nil)
	if err != nil {
		return err
	}
	applyOptions := &chezmoi.ApplyOptions{
		DestDir:           ts.DestDir,
		DryRun:            c.DryRun,
		Ignore:            ts.TargetIgnore.Match,
		PersistentState:   persistentState,
		Remove:            c.Remove,
		ScriptStateBucket: c.scriptStateBucket,
		Stdout:            c.Stdout,
		Umask:             ts.Umask,
		Verbose:           c.Verbose,
		NoScript:          c.NoScript,
	}
	if len(args) == 0 {
		return ts.Apply(fs, c.mutator, c.Follow, applyOptions)
	}
	entries, err := c.getEntries(ts, args)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := entry.Apply(fs, c.mutator, c.Follow, applyOptions); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) autoCommit(vcs VCS) error {
	addArgs := vcs.AddArgs(".")
	if addArgs == nil {
		return fmt.Errorf("%s: autocommit not supported", c.SourceVCS.Command)
	}
	if err := c.run(c.SourceDir, c.SourceVCS.Command, addArgs...); err != nil {
		return err
	}
	output, err := c.output(c.SourceDir, c.SourceVCS.Command, vcs.StatusArgs()...)
	if err != nil {
		return err
	}
	status, err := vcs.ParseStatusOutput(output)
	if err != nil {
		return err
	}
	if gitStatus, ok := status.(*git.Status); ok && gitStatus.Empty() {
		return nil
	}
	commitMessageText, err := getAsset(commitMessageTemplateAsset)
	if err != nil {
		return err
	}
	commitMessageTmpl, err := template.New("commit_message").Funcs(c.templateFuncs).Parse(string(commitMessageText))
	if err != nil {
		return err
	}
	sb := &strings.Builder{}
	if err := commitMessageTmpl.Execute(sb, status); err != nil {
		return err
	}
	commitArgs := vcs.CommitArgs(sb.String())
	return c.run(c.SourceDir, c.SourceVCS.Command, commitArgs...)
}

func (c *Config) autoCommitAndAutoPush(cmd *cobra.Command, args []string) error {
	vcs, err := c.getVCS()
	if err != nil {
		return err
	}
	if c.DryRun {
		return nil
	}
	if c.SourceVCS.AutoCommit || c.SourceVCS.AutoPush {
		if err := c.autoCommit(vcs); err != nil {
			return err
		}
	}
	if c.SourceVCS.AutoPush {
		if err := c.autoPush(vcs); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) autoPush(vcs VCS) error {
	pushArgs := vcs.PushArgs()
	if pushArgs == nil {
		return fmt.Errorf("%s: autopush not supported", c.SourceVCS.Command)
	}
	return c.run(c.SourceDir, c.SourceVCS.Command, pushArgs...)
}

// ensureNoError ensures that no error was encountered when loading c.
func (c *Config) ensureNoError(cmd *cobra.Command, args []string) error {
	if c.err != nil {
		return errors.New("config contains errors, aborting")
	}
	return nil
}

func (c *Config) ensureSourceDirectory() error {
	info, err := c.fs.Stat(c.SourceDir)
	switch {
	case err == nil && info.IsDir():
		private, err := chezmoi.IsPrivate(c.fs, c.SourceDir, true)
		if err != nil {
			return err
		}
		if !private {
			if err := c.mutator.Chmod(c.SourceDir, 0o700&^os.FileMode(c.Umask)); err != nil {
				return err
			}
		}
		return nil
	case os.IsNotExist(err):
		if err := vfs.MkdirAll(c.mutator, filepath.Dir(c.SourceDir), 0o777&^os.FileMode(c.Umask)); err != nil {
			return err
		}
		return c.mutator.Mkdir(c.SourceDir, 0o700&^os.FileMode(c.Umask))
	case err == nil:
		return fmt.Errorf("%s: not a directory", c.SourceDir)
	default:
		return err
	}
}

func (c *Config) getData() (map[string]interface{}, error) {
	defaultData, err := c.getDefaultData()
	if err != nil {
		return nil, err
	}
	data := map[string]interface{}{
		"chezmoi": defaultData,
	}
	for key, value := range c.Data {
		data[key] = value
	}
	return data, nil
}

func (c *Config) getDefaultData() (map[string]interface{}, error) {
	data := map[string]interface{}{
		"arch":      runtime.GOARCH,
		"os":        runtime.GOOS,
		"sourceDir": c.SourceDir,
	}

	// Determine the user's username and group, if possible.
	//
	// user.Current and user.LookupGroupId in Go's standard library are
	// generally unreliable, so work around errors if possible, or ignore them.
	//
	// If CGO is disabled, then the Go standard library falls back to parsing
	// /etc/passwd and /etc/group, which will return incorrect results without
	// error if the system uses an alternative password database such as NIS or
	// LDAP.
	//
	// If CGO is enabled then user.Current and user.LookupGroupId will use the
	// underlying libc functions, namely getpwuid_r and getgrnam_r. If linked
	// with glibc this will return the correct result. If linked with musl then
	// they will use musl's implementation which, like Go's non-CGO
	// implementation, also only parses /etc/passwd and /etc/group and so also
	// returns incorrect results without error if NIS or LDAP are being used.
	//
	// Since neither the username nor the group are likely widely used in
	// templates, leave these variables unset if their values cannot be
	// determined. Unset variables will trigger template errors if used,
	// alerting the user to the problem and allowing them to find alternative
	// solutions.

	// First, attempt to determine the current user using user.Current, falling
	// back to the $USER environment variable if set, and otherwise leaving
	// username unset.
	currentUser, err := user.Current()
	if err == nil {
		data["username"] = currentUser.Username
	} else if user, ok := os.LookupEnv("USER"); ok {
		data["username"] = user
	}

	// If the current user could be determined, then attempt to lookup the group
	// id. There is no fallback.
	if currentUser != nil {
		if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
			data["group"] = group.Name
		}
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

	osRelease, err := getOSRelease(c.fs)
	if err == nil {
		if osRelease != nil {
			data["osRelease"] = upperSnakeCaseToCamelCaseMap(osRelease)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	kernelInfo, err := getKernelInfo(c.fs)
	if err == nil && kernelInfo != nil {
		data["kernel"] = kernelInfo
	} else if err != nil {
		return nil, err
	}

	return data, nil
}

func (c *Config) getEditor() (string, []string) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}
	components := whitespaceRegexp.Split(editor, -1)
	return components[0], components[1:]
}

func (c *Config) getEntries(ts *chezmoi.TargetState, args []string) ([]chezmoi.Entry, error) {
	entries := []chezmoi.Entry{}
	for _, arg := range args {
		targetPath, err := filepath.Abs(arg)
		if err != nil {
			return nil, err
		}
		entry, err := ts.Get(c.fs, targetPath)
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

func (c *Config) getPersistentState(options *bolt.Options) (chezmoi.PersistentState, error) {
	persistentStateFile := c.getPersistentStateFile()
	if options == nil {
		options = &bolt.Options{}
	}
	if options.Timeout == 0 {
		options.Timeout = 2 * time.Second
	}
	if c.DryRun {
		options.ReadOnly = true
	}
	state, err := chezmoi.NewBoltPersistentState(c.fs, persistentStateFile, options)
	if errors.Is(err, bolt.ErrTimeout) {
		return nil, fmt.Errorf("failed to lock database: %w", err)
	}
	return state, err
}

func (c *Config) getPersistentStateFile() string {
	if c.configFile != "" {
		return filepath.Join(filepath.Dir(c.configFile), "chezmoistate.boltdb")
	}
	for _, configDir := range c.bds.ConfigDirs {
		persistentStateFile := filepath.Join(configDir, "chezmoi", "chezmoistate.boltdb")
		if _, err := os.Stat(persistentStateFile); err == nil {
			return persistentStateFile
		}
	}
	return filepath.Join(filepath.Dir(getDefaultConfigFile(c.bds)), "chezmoistate.boltdb")
}

func (c *Config) getTargetState(populateOptions *chezmoi.PopulateOptions) (*chezmoi.TargetState, error) {
	fs := vfs.NewReadOnlyFS(c.fs)

	data, err := c.getData()
	if err != nil {
		return nil, err
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

	ts := chezmoi.NewTargetState(
		chezmoi.WithDestDir(destDir),
		chezmoi.WithGPG(&c.GPG),
		chezmoi.WithSourceDir(c.SourceDir),
		chezmoi.WithTemplateData(data),
		chezmoi.WithTemplateFuncs(c.templateFuncs),
		chezmoi.WithTemplateOptions(c.Template.Options),
		chezmoi.WithUmask(os.FileMode(c.Umask)),
	)
	if err := ts.Populate(fs, populateOptions); err != nil {
		return nil, err
	}
	if Version != nil && !isDevVersion(Version) && ts.MinVersion != nil && Version.LessThan(*ts.MinVersion) {
		return nil, fmt.Errorf("chezmoi version %s too old, source state requires at least %s", Version, ts.MinVersion)
	}
	return ts, nil
}

func (c *Config) getVCS() (VCS, error) {
	vcs, ok := vcses[filepath.Base(c.SourceVCS.Command)]
	if !ok {
		return nil, fmt.Errorf("%s: unsupported source VCS command", c.SourceVCS.Command)
	}
	return vcs, nil
}

func (c *Config) output(dir, name string, argv ...string) ([]byte, error) {
	cmd := exec.Command(name, argv...)
	if dir != "" {
		var err error
		cmd.Dir, err = c.fs.RawPath(dir)
		if err != nil {
			return nil, err
		}
	}
	return c.mutator.IdempotentCmdOutput(cmd)
}

//nolint:unparam
func (c *Config) prompt(s, choices string) (byte, error) {
	r := bufio.NewReader(c.Stdin)
	for {
		_, err := fmt.Printf("%s [%s]? ", s, strings.Join(strings.Split(choices, ""), ","))
		if err != nil {
			return 0, err
		}
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if len(line) == 1 && strings.IndexByte(choices, line[0]) != -1 {
			return line[0], nil
		}
	}
}

// run runs name argv... in dir.
func (c *Config) run(dir, name string, argv ...string) error {
	cmd := exec.Command(name, argv...)
	if dir != "" {
		var err error
		cmd.Dir, err = c.fs.RawPath(dir)
		if err != nil {
			return err
		}
	}
	cmd.Stdin = c.Stdin
	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stdout
	return c.mutator.RunCmd(cmd)
}

func (c *Config) runEditor(argv ...string) error {
	editorName, editorArgs := c.getEditor()
	return c.run("", editorName, append(editorArgs, argv...)...)
}

func (c *Config) validateData() error {
	return validateKeys(config.Data, identifierRegexp)
}

func getAsset(name string) ([]byte, error) {
	asset, ok := assets[name]
	if !ok {
		return nil, fmt.Errorf("%s: not found", name)
	}
	return asset, nil
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

// isDevVersion returns true if version is a development version (i.e. that the
// major, minor, and patch version numbers are all zero).
func isDevVersion(v *semver.Version) bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
}

// isWellKnownAbbreviation returns true if word is a well known abbreviation.
func isWellKnownAbbreviation(word string) bool {
	_, ok := wellKnownAbbreviations[word]
	return ok
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
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

// validateKeys ensures that all keys in data match re.
func validateKeys(data interface{}, re *regexp.Regexp) error {
	switch data := data.(type) {
	case map[string]interface{}:
		for key, value := range data {
			if !re.MatchString(key) {
				return fmt.Errorf("invalid key: %q", key)
			}
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, value := range data {
			if err := validateKeys(value, re); err != nil {
				return err
			}
		}
	}
	return nil
}
