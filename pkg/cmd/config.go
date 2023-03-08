package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-semver/semver"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/mitchellh/mapstructure"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-xdg/v6"
	cobracompletefig "github.com/withfig/autocomplete-tools/integrations/cobra"
	"go.uber.org/multierr"
	"golang.org/x/exp/maps"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"

	"github.com/twpayne/chezmoi/v2/assets/templates"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
	"github.com/twpayne/chezmoi/v2/pkg/git"
	"github.com/twpayne/chezmoi/v2/pkg/shell"
)

const (
	logComponentKey                  = "component"
	logComponentValueEncryption      = "encryption"
	logComponentValuePersistentState = "persistentState"
	logComponentValueSourceState     = "sourceState"
	logComponentValueSystem          = "system"
)

type doPurgeOptions struct {
	binary          bool
	cache           bool
	config          bool
	persistentState bool
	sourceDir       bool
	workingTree     bool
}

type commandConfig struct {
	Command string   `json:"command" mapstructure:"command" yaml:"command"`
	Args    []string `json:"args" mapstructure:"args" yaml:"args"`
}

type hookConfig struct {
	Pre  commandConfig `json:"pre" mapstructure:"pre" yaml:"pre"`
	Post commandConfig `json:"post" mapstructure:"post" yaml:"post"`
}

type templateConfig struct {
	Options []string `json:"options" mapstructure:"options" yaml:"options"`
}

type warningsConfig struct {
	ConfigFileTemplateHasChanged bool `json:"configFileTemplateHasChanged" mapstructure:"configFileTemplateHasChanged" yaml:"configFileTemplateHasChanged"` //nolint:lll
}

// ConfigFile contains all data settable in the config file.
type ConfigFile struct {
	// Global configuration.
	CacheDirAbsPath    chezmoi.AbsPath                 `json:"cacheDir" mapstructure:"cacheDir" yaml:"cacheDir"`
	Color              autoBool                        `json:"color" mapstructure:"color" yaml:"color"`
	Data               map[string]any                  `json:"data" mapstructure:"data" yaml:"data"`
	Format             writeDataFormat                 `json:"format" mapstructure:"format" yaml:"format"`
	DestDirAbsPath     chezmoi.AbsPath                 `json:"destDir" mapstructure:"destDir" yaml:"destDir"`
	GitHub             gitHubConfig                    `json:"gitHub" mapstructure:"gitHub" yaml:"gitHub"`
	Hooks              map[string]hookConfig           `json:"hooks" mapstructure:"hooks" yaml:"hooks"`
	Interpreters       map[string]*chezmoi.Interpreter `json:"interpreters" mapstructure:"interpreters" yaml:"interpreters"` //nolint:lll
	Mode               chezmoi.Mode                    `json:"mode" mapstructure:"mode" yaml:"mode"`
	Pager              string                          `json:"pager" mapstructure:"pager" yaml:"pager"`
	PINEntry           pinEntryConfig                  `json:"pinentry" mapstructure:"pinentry" yaml:"pinentry"`
	Progress           autoBool                        `json:"progress" mapstructure:"progress" yaml:"progress"`
	Safe               bool                            `json:"safe" mapstructure:"safe" yaml:"safe"`
	ScriptEnv          map[string]string               `json:"scriptEnv" mapstructure:"scriptEnv" yaml:"scriptEnv"`
	ScriptTempDir      chezmoi.AbsPath                 `json:"scriptTempDir" mapstructure:"scriptTempDir" yaml:"scriptTempDir"` //nolint:lll
	SourceDirAbsPath   chezmoi.AbsPath                 `json:"sourceDir" mapstructure:"sourceDir" yaml:"sourceDir"`
	Template           templateConfig                  `json:"template" mapstructure:"template" yaml:"template"`
	TextConv           textConv                        `json:"textConv" mapstructure:"textConv" yaml:"textConv"`
	Umask              fs.FileMode                     `json:"umask" mapstructure:"umask" yaml:"umask"`
	UseBuiltinAge      autoBool                        `json:"useBuiltinAge" mapstructure:"useBuiltinAge" yaml:"useBuiltinAge"` //nolint:lll
	UseBuiltinGit      autoBool                        `json:"useBuiltinGit" mapstructure:"useBuiltinGit" yaml:"useBuiltinGit"` //nolint:lll
	Verbose            bool                            `json:"verbose" mapstructure:"verbose" yaml:"verbose"`
	Warnings           warningsConfig                  `json:"warnings" mapstructure:"warnings" yaml:"warnings"`
	WorkingTreeAbsPath chezmoi.AbsPath                 `json:"workingTree" mapstructure:"workingTree" yaml:"workingTree"`

	// Password manager configurations.
	AWSSecretsManager awsSecretsManagerConfig `json:"awsSecretsManager" mapstructure:"awsSecretsManager" yaml:"awsSecretsManager"` //nolint:lll
	Bitwarden         bitwardenConfig         `json:"bitwarden" mapstructure:"bitwarden" yaml:"bitwarden"`
	Dashlane          dashlaneConfig          `json:"dashlane" mapstructure:"dashlane" yaml:"dashlane"`
	Ejson             ejsonConfig             `json:"ejson" mapstructure:"ejson" yaml:"ejson"`
	Gopass            gopassConfig            `json:"gopass" mapstructure:"gopass" yaml:"gopass"`
	Keepassxc         keepassxcConfig         `json:"keepassxc" mapstructure:"keepassxc" yaml:"keepassxc"`
	Keeper            keeperConfig            `json:"keeper" mapstructure:"keeper" yaml:"keeper"`
	Lastpass          lastpassConfig          `json:"lastpass" mapstructure:"lastpass" yaml:"lastpass"`
	Onepassword       onepasswordConfig       `json:"onepassword" mapstructure:"onepassword" yaml:"onepassword"`
	Pass              passConfig              `json:"pass" mapstructure:"pass" yaml:"pass"`
	Passhole          passholeConfig          `json:"passhole" mapstructure:"passhole" yaml:"passhole"`
	RBW               rbwConfig               `json:"rbw" mapstructure:"rbw" yaml:"rbw"`
	Secret            secretConfig            `json:"secret" mapstructure:"secret" yaml:"secret"`
	Vault             vaultConfig             `json:"vault" mapstructure:"vault" yaml:"vault"`

	// Encryption configurations.
	Encryption string                `json:"encryption" mapstructure:"encryption" yaml:"encryption"`
	Age        chezmoi.AgeEncryption `json:"age" mapstructure:"age" yaml:"age"`
	GPG        chezmoi.GPGEncryption `json:"gpg" mapstructure:"gpg" yaml:"gpg"`

	// Command configurations.
	Add        addCmdConfig        `json:"add" mapstructure:"add" yaml:"add"`
	CD         cdCmdConfig         `json:"cd" mapstructure:"cd" yaml:"cd"`
	Completion completionCmdConfig `json:"completion" mapstructure:"completion" yaml:"completion"`
	Diff       diffCmdConfig       `json:"diff" mapstructure:"diff" yaml:"diff"`
	Edit       editCmdConfig       `json:"edit" mapstructure:"edit" yaml:"edit"`
	Git        gitCmdConfig        `json:"git" mapstructure:"git" yaml:"git"`
	Merge      mergeCmdConfig      `json:"merge" mapstructure:"merge" yaml:"merge"`
	Status     statusCmdConfig     `json:"status" mapstructure:"status" yaml:"status"`
	Update     updateCmdConfig     `json:"update" mapstructure:"update" yaml:"update"`
	Verify     verifyCmdConfig     `json:"verify" mapstructure:"verify" yaml:"verify"`
}

// A Config represents a configuration.
type Config struct {
	ConfigFile

	// Global configuration.
	configFormat     readDataFormat
	cpuProfile       chezmoi.AbsPath
	debug            bool
	dryRun           bool
	force            bool
	homeDir          string
	interactive      bool
	keepGoing        bool
	noPager          bool
	noTTY            bool
	outputAbsPath    chezmoi.AbsPath
	refreshExternals chezmoi.RefreshExternals
	sourcePath       bool
	templateFuncs    template.FuncMap

	// Password manager data.
	gitHub  gitHubData
	keyring keyringData

	// Command configurations, not settable in the config file.
	apply           applyCmdConfig
	archive         archiveCmdConfig
	chattr          chattrCmdConfig
	dump            dumpCmdConfig
	executeTemplate executeTemplateCmdConfig
	_import         importCmdConfig
	init            initCmdConfig
	managed         managedCmdConfig
	mergeAll        mergeAllCmdConfig
	purge           purgeCmdConfig
	reAdd           reAddCmdConfig
	remove          removeCmdConfig
	secret          secretCmdConfig
	state           stateCmdConfig
	unmanaged       unmanagedCmdConfig
	upgrade         upgradeCmdConfig

	// Version information.
	version     semver.Version
	versionInfo VersionInfo
	versionStr  string

	// Configuration.
	fileSystem             vfs.FS
	bds                    *xdg.BaseDirectorySpecification
	configFileAbsPath      chezmoi.AbsPath
	configFileAbsPathErr   error
	baseSystem             chezmoi.System
	sourceSystem           chezmoi.System
	destSystem             chezmoi.System
	persistentStateAbsPath chezmoi.AbsPath
	persistentState        chezmoi.PersistentState
	httpClient             *http.Client
	logger                 *zerolog.Logger

	// Computed configuration.
	homeDirAbsPath      chezmoi.AbsPath
	encryption          chezmoi.Encryption
	sourceDirAbsPath    chezmoi.AbsPath
	sourceDirAbsPathErr error
	sourceState         *chezmoi.SourceState
	sourceStateErr      error
	templateData        *templateData
	runEnv              []string

	stdin       io.Reader
	stdout      io.Writer
	stderr      io.Writer
	bufioReader *bufio.Reader

	tempDirs map[string]chezmoi.AbsPath

	ioregData ioregData

	restoreWindowsConsole func() error
}

type templateData struct {
	Arch           string          `json:"arch"`
	Args           []string        `json:"args"`
	CacheDir       chezmoi.AbsPath `json:"cacheDir"`
	Command        string          `json:"command"`
	ConfigFile     chezmoi.AbsPath `json:"configFile"`
	Executable     chezmoi.AbsPath `json:"executable"`
	FQDNHostname   string          `json:"fqdnHostname"`
	GID            string          `json:"gid"`
	Group          string          `json:"group"`
	HomeDir        chezmoi.AbsPath `json:"homeDir"`
	Hostname       string          `json:"hostname"`
	Kernel         map[string]any  `json:"kernel"`
	OS             string          `json:"os"`
	OSRelease      map[string]any  `json:"osRelease"`
	SourceDir      chezmoi.AbsPath `json:"sourceDir"`
	UID            string          `json:"uid"`
	Username       string          `json:"username"`
	Version        map[string]any  `json:"version"`
	WindowsVersion map[string]any  `json:"windowsVersion"`
	WorkingTree    chezmoi.AbsPath `json:"workingTree"`
}

// A configOption sets and option on a Config.
type configOption func(*Config) error

type configState struct {
	ConfigTemplateContentsSHA256 chezmoi.HexBytes `json:"configTemplateContentsSHA256" yaml:"configTemplateContentsSHA256"` //nolint:lll,tagliatelle
}

var (
	chezmoiRelPath             = chezmoi.NewRelPath("chezmoi")
	persistentStateFileRelPath = chezmoi.NewRelPath("chezmoistate.boltdb")
	httpCacheDirRelPath        = chezmoi.NewRelPath("httpcache")

	configStateKey = []byte("configState")

	defaultAgeEncryptionConfig = chezmoi.AgeEncryption{
		Command: "age",
		Suffix:  ".age",
	}
	defaultGPGEncryptionConfig = chezmoi.GPGEncryption{
		Command: "gpg",
		Suffix:  ".asc",
	}

	whitespaceRx = regexp.MustCompile(`\s+`)
)

// newConfig creates a new Config with the given options.
func newConfig(options ...configOption) (*Config, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	homeDirAbsPath, err := chezmoi.NormalizePath(userHomeDir)
	if err != nil {
		return nil, err
	}

	bds, err := xdg.NewBaseDirectorySpecification()
	if err != nil {
		return nil, err
	}

	c := &Config{
		ConfigFile: newConfigFile(bds),

		// Global configuration.
		homeDir:       userHomeDir,
		templateFuncs: sprig.TxtFuncMap(),

		// Command configurations.
		apply: applyCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		archive: archiveCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		dump: dumpCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		executeTemplate: executeTemplateCmdConfig{
			stdinIsATTY: true,
		},
		_import: importCmdConfig{
			destination: homeDirAbsPath,
			filter:      chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
		},
		init: initCmdConfig{
			data:              true,
			filter:            chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			guessRepoURL:      true,
			recurseSubmodules: true,
		},
		managed: managedCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			pathStyle: pathStyleRelative,
		},
		mergeAll: mergeAllCmdConfig{
			recursive: true,
		},
		reAdd: reAddCmdConfig{
			filter: chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
		},
		unmanaged: unmanagedCmdConfig{
			pathStyle: pathStyleRelative,
		},
		upgrade: upgradeCmdConfig{
			owner: gitHubOwner,
			repo:  gitHubRepo,
		},

		// Configuration.
		fileSystem: vfs.OSFS,
		bds:        bds,

		// Computed configuration.
		homeDirAbsPath: homeDirAbsPath,

		tempDirs: make(map[string]chezmoi.AbsPath),

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	// Override sprig's toPrettyJson template function. Delete it from the
	// template function map first to avoid a duplication function panic.
	delete(c.templateFuncs, "toPrettyJson")

	// The completion template function is added in persistentPreRunRootE as
	// it needs a *cobra.Command, which we don't yet have.
	for key, value := range map[string]any{
		"awsSecretsManager":        c.awsSecretsManagerTemplateFunc,
		"awsSecretsManagerRaw":     c.awsSecretsManagerRawTemplateFunc,
		"bitwarden":                c.bitwardenTemplateFunc,
		"bitwardenAttachment":      c.bitwardenAttachmentTemplateFunc,
		"bitwardenAttachmentByRef": c.bitwardenAttachmentByRefTemplateFunc,
		"bitwardenFields":          c.bitwardenFieldsTemplateFunc,
		"comment":                  c.commentTemplateFunc,
		"dashlaneNote":             c.dashlaneNoteTemplateFunc,
		"dashlanePassword":         c.dashlanePasswordTemplateFunc,
		"decrypt":                  c.decryptTemplateFunc,
		"deleteValueAtPath":        c.deleteValueAtPathTemplateFunc,
		"ejsonDecrypt":             c.ejsonDecryptTemplateFunc,
		"ejsonDecryptWithKey":      c.ejsonDecryptWithKeyTemplateFunc,
		"encrypt":                  c.encryptTemplateFunc,
		"eqFold":                   c.eqFoldTemplateFunc,
		"fromIni":                  c.fromIniTemplateFunc,
		"fromJsonc":                c.fromJsoncTemplateFunc,
		"fromToml":                 c.fromTomlTemplateFunc,
		"fromYaml":                 c.fromYamlTemplateFunc,
		"gitHubKeys":               c.gitHubKeysTemplateFunc,
		"gitHubLatestRelease":      c.gitHubLatestReleaseTemplateFunc,
		"gitHubLatestTag":          c.gitHubLatestTagTemplateFunc,
		"glob":                     c.globTemplateFunc,
		"gopass":                   c.gopassTemplateFunc,
		"gopassRaw":                c.gopassRawTemplateFunc,
		"hexDecode":                c.hexDecodeTemplateFunc,
		"hexEncode":                c.hexEncodeTemplateFunc,
		"include":                  c.includeTemplateFunc,
		"includeTemplate":          c.includeTemplateTemplateFunc,
		"ioreg":                    c.ioregTemplateFunc,
		"joinPath":                 c.joinPathTemplateFunc,
		"jq":                       c.jqTemplateFunc,
		"keepassxc":                c.keepassxcTemplateFunc,
		"keepassxcAttachment":      c.keepassxcAttachmentTemplateFunc,
		"keepassxcAttribute":       c.keepassxcAttributeTemplateFunc,
		"keeper":                   c.keeperTemplateFunc,
		"keeperDataFields":         c.keeperDataFieldsTemplateFunc,
		"keeperFindPassword":       c.keeperFindPasswordTemplateFunc,
		"keyring":                  c.keyringTemplateFunc,
		"lastpass":                 c.lastpassTemplateFunc,
		"lastpassRaw":              c.lastpassRawTemplateFunc,
		"lookPath":                 c.lookPathTemplateFunc,
		"lstat":                    c.lstatTemplateFunc,
		"mozillaInstallHash":       c.mozillaInstallHashTemplateFunc,
		"onepassword":              c.onepasswordTemplateFunc,
		"onepasswordDetailsFields": c.onepasswordDetailsFieldsTemplateFunc,
		"onepasswordDocument":      c.onepasswordDocumentTemplateFunc,
		"onepasswordItemFields":    c.onepasswordItemFieldsTemplateFunc,
		"onepasswordRead":          c.onepasswordReadTemplateFunc,
		"output":                   c.outputTemplateFunc,
		"pass":                     c.passTemplateFunc,
		"passFields":               c.passFieldsTemplateFunc,
		"passRaw":                  c.passRawTemplateFunc,
		"passhole":                 c.passholeTemplateFunc,
		"pruneEmptyDicts":          c.pruneEmptyDictsTemplateFunc,
		"quoteList":                c.quoteListTemplateFunc,
		"rbw":                      c.rbwTemplateFunc,
		"rbwFields":                c.rbwFieldsTemplateFunc,
		"replaceAllRegex":          c.replaceAllRegexTemplateFunc,
		"secret":                   c.secretTemplateFunc,
		"secretJSON":               c.secretJSONTemplateFunc,
		"setValueAtPath":           c.setValueAtPathTemplateFunc,
		"stat":                     c.statTemplateFunc,
		"toIni":                    c.toIniTemplateFunc,
		"toPrettyJson":             c.toPrettyJsonTemplateFunc,
		"toToml":                   c.toTomlTemplateFunc,
		"toYaml":                   c.toYamlTemplateFunc,
		"vault":                    c.vaultTemplateFunc,
	} {
		c.addTemplateFunc(key, value)
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	c.homeDirAbsPath, err = chezmoi.NormalizePath(c.homeDir)
	if err != nil {
		return nil, err
	}
	c.configFileAbsPath, c.configFileAbsPathErr = c.defaultConfigFile(c.fileSystem, c.bds)
	c.SourceDirAbsPath, err = c.defaultSourceDir(c.fileSystem, c.bds)
	if err != nil {
		return nil, err
	}
	c.DestDirAbsPath = c.homeDirAbsPath
	c._import.destination = c.homeDirAbsPath

	return c, nil
}

// Close closes resources associated with c.
func (c *Config) Close() error {
	var err error
	for _, tempDirAbsPath := range c.tempDirs {
		err2 := os.RemoveAll(tempDirAbsPath.String())
		c.logger.Err(err2).
			Stringer("tempDir", tempDirAbsPath).
			Msg("RemoveAll")
		err = multierr.Append(err, err2)
	}
	pprof.StopCPUProfile()
	return err
}

// addTemplateFunc adds the template function with the given key and value
// to c. It panics if there is already an existing template function with the
// same key.
func (c *Config) addTemplateFunc(key string, value any) {
	if _, ok := c.templateFuncs[key]; ok {
		panic(fmt.Sprintf("%s: already defined", key))
	}
	c.templateFuncs[key] = value
}

type applyArgsOptions struct {
	cmd          *cobra.Command
	filter       *chezmoi.EntryTypeFilter
	init         bool
	recursive    bool
	umask        fs.FileMode
	preApplyFunc chezmoi.PreApplyFunc
}

// applyArgs is the core of all commands that make changes to a target system.
// It checks config file freshness, reads the source state, and then applies the
// source state for each target entry in args. If args is empty then the source
// state is applied to all target entries.
func (c *Config) applyArgs(
	ctx context.Context, targetSystem chezmoi.System, targetDirAbsPath chezmoi.AbsPath, args []string,
	options applyArgsOptions,
) error {
	if options.init {
		if err := c.createAndReloadConfigFile(options.cmd); err != nil {
			return err
		}
	}

	var currentConfigTemplateContentsSHA256 []byte
	configTemplate, err := c.findConfigTemplate()
	if err != nil {
		return err
	}
	if configTemplate != nil {
		currentConfigTemplateContentsSHA256 = chezmoi.SHA256Sum(configTemplate.contents)
	}
	var previousConfigTemplateContentsSHA256 []byte
	if configStateData, err := c.persistentState.Get(chezmoi.ConfigStateBucket, configStateKey); err != nil {
		return err
	} else if configStateData != nil {
		var configState configState
		if err := json.Unmarshal(configStateData, &configState); err != nil {
			return err
		}
		previousConfigTemplateContentsSHA256 = []byte(configState.ConfigTemplateContentsSHA256)
	}
	configTemplatesEmpty := currentConfigTemplateContentsSHA256 == nil && previousConfigTemplateContentsSHA256 == nil
	configTemplateContentsUnchanged := configTemplatesEmpty ||
		bytes.Equal(currentConfigTemplateContentsSHA256, previousConfigTemplateContentsSHA256)
	if !configTemplateContentsUnchanged {
		if c.force {
			if configTemplate == nil {
				if err := c.persistentState.Delete(chezmoi.ConfigStateBucket, configStateKey); err != nil {
					return err
				}
			} else {
				configStateValue, err := json.Marshal(configState{
					ConfigTemplateContentsSHA256: chezmoi.HexBytes(currentConfigTemplateContentsSHA256),
				})
				if err != nil {
					return err
				}
				if err := c.persistentState.Set(chezmoi.ConfigStateBucket, configStateKey, configStateValue); err != nil {
					return err
				}
			}
		} else if c.Warnings.ConfigFileTemplateHasChanged {
			c.errorf("warning: config file template has changed, run chezmoi init to regenerate config file\n")
		}
	}

	sourceState, err := c.getSourceState(ctx, options.cmd)
	if err != nil {
		return err
	}

	var targetRelPaths chezmoi.RelPaths
	switch {
	case len(args) == 0:
		targetRelPaths = sourceState.TargetRelPaths()
	case c.sourcePath:
		targetRelPaths, err = c.targetRelPathsBySourcePath(sourceState, args)
		if err != nil {
			return err
		}
	default:
		targetRelPaths, err = c.targetRelPaths(sourceState, args, targetRelPathsOptions{
			mustBeInSourceState: true,
			recursive:           options.recursive,
		})
		if err != nil {
			return err
		}
	}

	applyOptions := chezmoi.ApplyOptions{
		Filter:       options.filter,
		PreApplyFunc: options.preApplyFunc,
		Umask:        options.umask,
	}

	keptGoingAfterErr := false
	for _, targetRelPath := range targetRelPaths {
		switch err := sourceState.Apply(
			targetSystem, c.destSystem, c.persistentState, targetDirAbsPath, targetRelPath, applyOptions,
		); {
		case errors.Is(err, chezmoi.Skip):
			continue
		case err != nil && c.keepGoing:
			c.errorf("%v\n", err)
			keptGoingAfterErr = true
		case err != nil:
			return err
		}
	}

	switch err := sourceState.PostApply(targetSystem, targetDirAbsPath, targetRelPaths); {
	case err != nil && c.keepGoing:
		c.errorf("%v\n", err)
		keptGoingAfterErr = true
	case err != nil:
		return err
	}

	if keptGoingAfterErr {
		return chezmoi.ExitCodeError(1)
	}

	return nil
}

// checkVersion checks that chezmoi is at least the required version for the
// source state.
func (c *Config) checkVersion() error {
	versionAbsPath := c.SourceDirAbsPath.JoinString(chezmoi.VersionName)
	switch data, err := c.baseSystem.ReadFile(versionAbsPath); {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return err
	default:
		minVersion, err := semver.NewVersion(strings.TrimSpace(string(data)))
		if err != nil {
			return fmt.Errorf("%s: %q: %w", versionAbsPath, data, err)
		}
		var zeroVersion semver.Version
		if c.version != zeroVersion && c.version.LessThan(*minVersion) {
			return &chezmoi.TooOldError{
				Need: *minVersion,
				Have: c.version,
			}
		}
	}
	return nil
}

// cmdOutput returns the of running the command name with args in dirAbsPath.
func (c *Config) cmdOutput(dirAbsPath chezmoi.AbsPath, name string, args []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	if !dirAbsPath.Empty() {
		dirRawAbsPath, err := c.baseSystem.RawPath(dirAbsPath)
		if err != nil {
			return nil, err
		}
		cmd.Dir = dirRawAbsPath.String()
	}
	return chezmoilog.LogCmdOutput(cmd)
}

// colorAutoFunc detects whether color should be used.
func (c *Config) colorAutoFunc() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if stdout, ok := c.stdout.(*os.File); ok {
		return term.IsTerminal(int(stdout.Fd()))
	}
	return false
}

// createAndReloadConfigFile creates a config file if it there is a config file
// template and reloads it.
func (c *Config) createAndReloadConfigFile(cmd *cobra.Command) error {
	// Find config template, execute it, and create config file.
	configTemplate, err := c.findConfigTemplate()
	if err != nil {
		return err
	}

	if configTemplate == nil {
		if err := c.persistentState.Delete(chezmoi.ConfigStateBucket, configStateKey); err != nil {
			return err
		}
		return nil
	}

	configFileContents, err := c.createConfigFile(configTemplate.targetRelPath, configTemplate.contents, cmd)
	if err != nil {
		return err
	}

	// Validate the configMap.
	var configFile ConfigFile
	if err := c.decodeConfigBytes(configTemplate.format, configFileContents, &configFile); err != nil {
		return fmt.Errorf("%s: %w", configTemplate.sourceAbsPath, err)
	}

	// Write the config.
	configPath := c.init.configPath
	if c.init.configPath.Empty() {
		configPath = chezmoi.NewAbsPath(c.bds.ConfigHome).Join(chezmoiRelPath, configTemplate.targetRelPath)
	}
	if err := chezmoi.MkdirAll(c.baseSystem, configPath.Dir(), fs.ModePerm); err != nil {
		return err
	}
	if err := c.baseSystem.WriteFile(configPath, configFileContents, 0o600); err != nil {
		return err
	}

	configStateValue, err := json.Marshal(configState{
		ConfigTemplateContentsSHA256: chezmoi.HexBytes(chezmoi.SHA256Sum(configTemplate.contents)),
	})
	if err != nil {
		return err
	}
	if err := c.persistentState.Set(chezmoi.ConfigStateBucket, configStateKey, configStateValue); err != nil {
		return err
	}

	// Reload the config.
	if err := c.decodeConfigBytes(configTemplate.format, configFileContents, &c.ConfigFile); err != nil {
		return fmt.Errorf("%s: %w", configTemplate.sourceAbsPath, err)
	}
	if err := c.setEncryption(); err != nil {
		return err
	}

	return nil
}

// createConfigFile creates a config file using a template and returns its
// contents.
func (c *Config) createConfigFile(filename chezmoi.RelPath, data []byte, cmd *cobra.Command) ([]byte, error) {
	funcMap := make(template.FuncMap)
	chezmoi.RecursiveMerge(funcMap, c.templateFuncs)
	initTemplateFuncs := map[string]any{
		"exit":             c.exitInitTemplateFunc,
		"promptBool":       c.promptBoolInitTemplateFunc,
		"promptBoolOnce":   c.promptBoolOnceInitTemplateFunc,
		"promptInt":        c.promptIntInitTemplateFunc,
		"promptIntOnce":    c.promptIntOnceInitTemplateFunc,
		"promptString":     c.promptStringInitTemplateFunc,
		"promptStringOnce": c.promptStringOnceInitTemplateFunc,
		"stdinIsATTY":      c.stdinIsATTYInitTemplateFunc,
		"writeToStdout":    c.writeToStdout,
	}
	chezmoi.RecursiveMerge(funcMap, initTemplateFuncs)

	tmpl, err := chezmoi.ParseTemplate(filename.String(), data, funcMap, chezmoi.TemplateOptions{
		Options: append([]string(nil), c.Template.Options...),
	})
	if err != nil {
		return nil, err
	}

	templateData := c.getTemplateDataMap(cmd)
	if c.init.data {
		chezmoi.RecursiveMerge(templateData, c.Data)
	}
	return tmpl.Execute(templateData)
}

// defaultConfigFile returns the default config file according to the XDG Base
// Directory Specification.
func (c *Config) defaultConfigFile(fileSystem vfs.FS, bds *xdg.BaseDirectorySpecification) (chezmoi.AbsPath, error) {
	// Search XDG Base Directory Specification config directories first.
CONFIG_DIR:
	for _, configDir := range bds.ConfigDirs {
		configDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(configDir, c.homeDirAbsPath)
		if err != nil {
			return chezmoi.EmptyAbsPath, err
		}

		dirEntries, err := fileSystem.ReadDir(configDirAbsPath.JoinString("chezmoi").String())
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue CONFIG_DIR
		case err != nil:
			return chezmoi.EmptyAbsPath, err
		}

		dirEntryNames := make(map[string]struct{}, len(dirEntries))
		for _, dirEntry := range dirEntries {
			dirEntryNames[dirEntry.Name()] = struct{}{}
		}

		var names []string
		for _, extension := range chezmoi.FormatExtensions {
			name := "chezmoi." + extension
			if _, ok := dirEntryNames[name]; ok {
				names = append(names, name)
			}
		}

		switch len(names) {
		case 0:
			// Do nothing.
		case 1:
			return configDirAbsPath.JoinString("chezmoi", names[0]), nil
		default:
			configFileAbsPathStrs := make([]string, 0, len(names))
			for _, name := range names {
				configFileAbsPathStr := configDirAbsPath.JoinString("chezmoi", name)
				configFileAbsPathStrs = append(configFileAbsPathStrs, configFileAbsPathStr.String())
			}
			return chezmoi.EmptyAbsPath, fmt.Errorf("multiple config files: %s", englishList(configFileAbsPathStrs))
		}
	}

	// Fallback to XDG Base Directory Specification default.
	configHomeAbsPath, err := chezmoi.NewAbsPathFromExtPath(bds.ConfigHome, c.homeDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return configHomeAbsPath.JoinString("chezmoi", "chezmoi.toml"), nil
}

// decodeConfigBytes decodes data in format into configFile.
func (c *Config) decodeConfigBytes(format chezmoi.Format, data []byte, configFile *ConfigFile) error {
	var configMap map[string]any
	if err := format.Unmarshal(data, &configMap); err != nil {
		return err
	}
	return c.decodeConfigMap(configMap, configFile)
}

// decodeConfigFile decodes the config file at configFileAbsPath into
// configFile.
func (c *Config) decodeConfigFile(configFileAbsPath chezmoi.AbsPath, configFile *ConfigFile) error {
	var format chezmoi.Format
	if c.configFormat == "" {
		var err error
		format, err = chezmoi.FormatFromAbsPath(configFileAbsPath)
		if err != nil {
			return err
		}
	} else {
		format = c.configFormat.Format()
	}

	configFileContents, err := c.fileSystem.ReadFile(configFileAbsPath.String())
	if err != nil {
		return fmt.Errorf("%s: %w", configFileAbsPath, err)
	}

	if err := c.decodeConfigBytes(format, configFileContents, configFile); err != nil {
		return fmt.Errorf("%s: %w", configFileAbsPath, err)
	}

	return nil
}

// decodeConfigMap decodes configMap into configFile.
func (c *Config) decodeConfigMap(configMap map[string]any, configFile *ConfigFile) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			chezmoi.StringSliceToEntryTypeSetHookFunc(),
			chezmoi.StringToAbsPathHookFunc(),
			StringOrBoolToAutoBoolHookFunc(),
		),
		Result: configFile,
	})
	if err != nil {
		return err
	}
	if err := decoder.Decode(configMap); err != nil {
		return err
	}
	return nil
}

// defaultPreApplyFunc is the default pre-apply function. If the target entry
// has changed since chezmoi last wrote it then it prompts the user for the
// action to take.
func (c *Config) defaultPreApplyFunc(
	targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState,
) error {
	c.logger.Info().
		Stringer("targetRelPath", targetRelPath).
		Object("targetEntryState", targetEntryState).
		Object("lastWrittenEntryState", lastWrittenEntryState).
		Object("actualEntryState", actualEntryState).
		Msg("defaultPreApplyFunc")

	switch {
	case c.force:
		return nil
	case targetEntryState.Equivalent(actualEntryState):
		return nil
	}

	if c.interactive {
		prompt := fmt.Sprintf("Apply %s", targetRelPath)
		var choices []string
		actualContents := actualEntryState.Contents()
		targetContents := targetEntryState.Contents()
		if actualContents != nil || targetContents != nil {
			choices = append(choices, "diff")
		}
		choices = append(choices, choicesYesNoAllQuit...)
		for {
			switch choice, err := c.promptChoice(prompt, choices); {
			case err != nil:
				return err
			case choice == "diff":
				if err := c.diffFile(
					targetRelPath,
					actualContents, actualEntryState.Mode,
					targetContents, targetEntryState.Mode,
				); err != nil {
					return err
				}
			case choice == "yes":
				return nil
			case choice == "no":
				return chezmoi.Skip
			case choice == "all":
				c.interactive = false
				return nil
			case choice == "quit":
				return chezmoi.ExitCodeError(0)
			default:
				panic(fmt.Sprintf("%s: unexpected choice", choice))
			}
		}
	}

	switch {
	case targetEntryState.Overwrite():
		return nil
	case targetEntryState.Type == chezmoi.EntryStateTypeScript:
		return nil
	case lastWrittenEntryState == nil:
		return nil
	case lastWrittenEntryState.Equivalent(actualEntryState):
		return nil
	}

	prompt := fmt.Sprintf("%s has changed since chezmoi last wrote it", targetRelPath)
	var choices []string
	actualContents := actualEntryState.Contents()
	targetContents := targetEntryState.Contents()
	if actualContents != nil || targetContents != nil {
		choices = append(choices, "diff")
	}
	choices = append(choices, "overwrite", "all-overwrite", "skip", "quit")
	for {
		switch choice, err := c.promptChoice(prompt, choices); {
		case err != nil:
			return err
		case choice == "diff":
			if err := c.diffFile(
				targetRelPath,
				actualContents, actualEntryState.Mode,
				targetContents, targetEntryState.Mode,
			); err != nil {
				return err
			}
		case choice == "overwrite":
			return nil
		case choice == "all-overwrite":
			c.force = true
			return nil
		case choice == "skip":
			return chezmoi.Skip
		case choice == "quit":
			return chezmoi.ExitCodeError(0)
		default:
			panic(fmt.Sprintf("%s: unexpected choice", choice))
		}
	}
}

// defaultSourceDir returns the default source directory according to the XDG
// Base Directory Specification.
func (c *Config) defaultSourceDir(fileSystem vfs.Stater, bds *xdg.BaseDirectorySpecification) (chezmoi.AbsPath, error) {
	// Check for XDG Base Directory Specification data directories first.
	for _, dataDir := range bds.DataDirs {
		dataDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(dataDir, c.homeDirAbsPath)
		if err != nil {
			return chezmoi.EmptyAbsPath, err
		}
		sourceDirAbsPath := dataDirAbsPath.Join(chezmoiRelPath)
		if _, err := fileSystem.Stat(sourceDirAbsPath.String()); err == nil {
			return sourceDirAbsPath, nil
		}
	}
	// Fallback to XDG Base Directory Specification default.
	dataHomeAbsPath, err := chezmoi.NewAbsPathFromExtPath(bds.DataHome, c.homeDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return dataHomeAbsPath.Join(chezmoiRelPath), nil
}

type destAbsPathInfosOptions struct {
	follow         bool
	ignoreNotExist bool
	recursive      bool
}

// destAbsPathInfos returns the os/fs.FileInfos for each destination entry in
// args, recursing into subdirectories and following symlinks if configured in
// options.
func (c *Config) destAbsPathInfos(
	sourceState *chezmoi.SourceState, args []string, options destAbsPathInfosOptions,
) (map[chezmoi.AbsPath]fs.FileInfo, error) {
	destAbsPathInfos := make(map[chezmoi.AbsPath]fs.FileInfo)
	for _, arg := range args {
		arg = filepath.Clean(arg)
		destAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		if _, err := c.targetRelPath(destAbsPath); err != nil {
			return nil, err
		}
		if options.recursive {
			walkFunc := func(destAbsPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
				switch {
				case options.ignoreNotExist && errors.Is(err, fs.ErrNotExist):
					return nil
				case err != nil:
					return err
				}
				if options.follow && fileInfo.Mode().Type() == fs.ModeSymlink {
					fileInfo, err = c.destSystem.Stat(destAbsPath)
					if err != nil {
						return err
					}
				}
				return sourceState.AddDestAbsPathInfos(destAbsPathInfos, c.destSystem, destAbsPath, fileInfo)
			}
			if err := chezmoi.Walk(c.destSystem, destAbsPath, walkFunc); err != nil {
				return nil, err
			}
		} else {
			var fileInfo fs.FileInfo
			if options.follow {
				fileInfo, err = c.destSystem.Stat(destAbsPath)
			} else {
				fileInfo, err = c.destSystem.Lstat(destAbsPath)
			}
			switch {
			case options.ignoreNotExist && errors.Is(err, fs.ErrNotExist):
				continue
			case err != nil:
				return nil, err
			}
			if err := sourceState.AddDestAbsPathInfos(destAbsPathInfos, c.destSystem, destAbsPath, fileInfo); err != nil {
				return nil, err
			}
		}
	}
	return destAbsPathInfos, nil
}

// diffFile outputs the diff between fromData and fromMode and toData and toMode
// at path.
func (c *Config) diffFile(
	path chezmoi.RelPath,
	fromData []byte, fromMode fs.FileMode,
	toData []byte, toMode fs.FileMode,
) error {
	builder := strings.Builder{}
	unifiedEncoder := diff.NewUnifiedEncoder(&builder, diff.DefaultContextLines)
	color := c.Color.Value(c.colorAutoFunc)
	if color {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	if fromMode.IsRegular() {
		var err error
		fromData, err = c.TextConv.convert(path.String(), fromData)
		if err != nil {
			return err
		}
	}
	if toMode.IsRegular() {
		var err error
		toData, err = c.TextConv.convert(path.String(), toData)
		if err != nil {
			return err
		}
	}
	diffPatch, err := chezmoi.DiffPatch(path, fromData, fromMode, toData, toMode)
	if err != nil {
		return err
	}
	if err := unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return c.pageOutputString(builder.String(), c.Diff.Pager)
}

// editor returns the path to the user's editor and any extra arguments.
func (c *Config) editor(args []string) (string, []string, error) {
	editCommand := c.Edit.Command
	editArgs := c.Edit.Args

	// If the user has set an edit command then use it.
	if editCommand != "" {
		return editCommand, append(editArgs, args...), nil
	}

	// Prefer $VISUAL over $EDITOR and fallback to the OS's default editor.
	editCommand = firstNonEmptyString(
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		defaultEditor,
	)

	return parseCommand(editCommand, append(editArgs, args...))
}

// errorf writes an error to stderr.
func (c *Config) errorf(format string, args ...any) {
	fmt.Fprintf(c.stderr, "chezmoi: "+format, args...)
}

// execute creates a new root command and executes it with args.
func (c *Config) execute(args []string) error {
	rootCmd, err := c.newRootCmd()
	if err != nil {
		return err
	}
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// filterInput reads from args (or the standard input if args is empty),
// transforms it with f, and writes the output.
func (c *Config) filterInput(args []string, f func([]byte) ([]byte, error)) error {
	if len(args) == 0 {
		input, err := io.ReadAll(c.stdin)
		if err != nil {
			return err
		}
		output, err := f(input)
		if err != nil {
			return err
		}
		return c.writeOutput(output)
	}

	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return err
		}
		input, err := c.baseSystem.ReadFile(argAbsPath)
		if err != nil {
			return err
		}
		output, err := f(input)
		if err != nil {
			return err
		}
		if err := c.writeOutput(output); err != nil {
			return err
		}
	}

	return nil
}

type configTemplate struct {
	sourceAbsPath chezmoi.AbsPath
	format        chezmoi.Format
	targetRelPath chezmoi.RelPath
	contents      []byte
}

// findConfigTemplate searches for a config template, returning the path,
// format, and contents. It returns an error if multiple config file templates
// are found.
func (c *Config) findConfigTemplate() (*configTemplate, error) {
	sourceDirAbsPath, err := c.getSourceDirAbsPath(nil)
	if err != nil {
		return nil, err
	}

	dirEntries, err := c.baseSystem.ReadDir(sourceDirAbsPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, nil
	case err != nil:
		return nil, err
	}

	dirEntryNames := make(map[chezmoi.RelPath]struct{}, len(dirEntries))
	for _, dirEntry := range dirEntries {
		dirEntryNames[chezmoi.NewRelPath(dirEntry.Name())] = struct{}{}
	}

	var configTemplates []*configTemplate //nolint:prealloc
	for _, extension := range chezmoi.FormatExtensions {
		relPath := chezmoi.NewRelPath(chezmoi.Prefix + "." + extension + chezmoi.TemplateSuffix)
		if _, ok := dirEntryNames[relPath]; !ok {
			continue
		}
		absPath := sourceDirAbsPath.Join(relPath)
		contents, err := c.baseSystem.ReadFile(absPath)
		if err != nil {
			return nil, err
		}
		configTemplate := &configTemplate{
			targetRelPath: chezmoi.NewRelPath("chezmoi." + extension),
			format:        chezmoi.FormatsByExtension[extension],
			sourceAbsPath: absPath,
			contents:      contents,
		}
		configTemplates = append(configTemplates, configTemplate)
	}

	switch len(configTemplates) {
	case 0:
		return nil, nil
	case 1:
		return configTemplates[0], nil
	default:
		sourceAbsPathStrs := make([]string, 0, len(configTemplates))
		for _, configTemplate := range configTemplates {
			sourceAbsPathStr := configTemplate.sourceAbsPath.String()
			sourceAbsPathStrs = append(sourceAbsPathStrs, sourceAbsPathStr)
		}
		return nil, fmt.Errorf("multiple config file templates: %s ", englishList(sourceAbsPathStrs))
	}
}

func (c *Config) getHTTPClient() (*http.Client, error) {
	if c.httpClient != nil {
		return c.httpClient, nil
	}

	httpCacheBasePath, err := c.baseSystem.RawPath(c.CacheDirAbsPath.Join(httpCacheDirRelPath))
	if err != nil {
		return nil, err
	}
	httpCache := diskcache.New(httpCacheBasePath.String())
	httpTransport := httpcache.NewTransport(httpCache)
	c.httpClient = httpTransport.Client()

	return c.httpClient, nil
}

type getSourceDirAbsPathOptions struct {
	refresh bool
}

// getSourceDirAbsPath returns the source directory, using .chezmoiroot if it
// exists.
func (c *Config) getSourceDirAbsPath(options *getSourceDirAbsPathOptions) (chezmoi.AbsPath, error) {
	if options == nil || !options.refresh {
		if !c.sourceDirAbsPath.Empty() || c.sourceDirAbsPathErr != nil {
			return c.sourceDirAbsPath, c.sourceDirAbsPathErr
		}
	}

	switch data, err := c.sourceSystem.ReadFile(c.SourceDirAbsPath.JoinString(chezmoi.RootName)); {
	case errors.Is(err, fs.ErrNotExist):
		c.sourceDirAbsPath = c.SourceDirAbsPath
	case err != nil:
		c.sourceDirAbsPathErr = err
	default:
		c.sourceDirAbsPath = c.SourceDirAbsPath.JoinString(string(bytes.TrimSpace(data)))
	}

	return c.sourceDirAbsPath, c.sourceDirAbsPathErr
}

func (c *Config) getSourceState(ctx context.Context, cmd *cobra.Command) (*chezmoi.SourceState, error) {
	if c.sourceState != nil || c.sourceStateErr != nil {
		return c.sourceState, c.sourceStateErr
	}
	c.sourceState, c.sourceStateErr = c.newSourceState(ctx, cmd)
	return c.sourceState, c.sourceStateErr
}

// getTemplateData returns the default template data.
func (c *Config) getTemplateData(cmd *cobra.Command) *templateData {
	if c.templateData == nil {
		c.templateData = c.newTemplateData(cmd)
	}
	return c.templateData
}

// getTemplateDataMap returns the template data as a map.
func (c *Config) getTemplateDataMap(cmd *cobra.Command) map[string]any {
	templateData := c.getTemplateData(cmd)

	return map[string]any{
		"chezmoi": map[string]any{
			"arch":           templateData.Arch,
			"args":           templateData.Args,
			"cacheDir":       templateData.CacheDir.String(),
			"command":        templateData.Command,
			"configFile":     templateData.ConfigFile.String(),
			"executable":     templateData.Executable.String(),
			"fqdnHostname":   templateData.FQDNHostname,
			"gid":            templateData.GID,
			"group":          templateData.Group,
			"homeDir":        templateData.HomeDir.String(),
			"hostname":       templateData.Hostname,
			"kernel":         templateData.Kernel,
			"os":             templateData.OS,
			"osRelease":      templateData.OSRelease,
			"sourceDir":      templateData.SourceDir.String(),
			"uid":            templateData.UID,
			"username":       templateData.Username,
			"version":        templateData.Version,
			"windowsVersion": templateData.WindowsVersion,
			"workingTree":    templateData.WorkingTree.String(),
		},
	}
}

// gitAutoAdd adds all changes to the git index and returns the new git status.
func (c *Config) gitAutoAdd() (*git.Status, error) {
	if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"add", "."}); err != nil {
		return nil, err
	}
	output, err := c.cmdOutput(c.WorkingTreeAbsPath, c.Git.Command, []string{"status", "--porcelain=v2"})
	if err != nil {
		return nil, err
	}
	return git.ParseStatusPorcelainV2(output)
}

// gitAutoCommit commits all changes in the git index, including generating a
// commit message from status.
func (c *Config) gitAutoCommit(status *git.Status) error {
	if status.Empty() {
		return nil
	}
	commitMessageTemplate, err := templates.FS.ReadFile("COMMIT_MESSAGE.tmpl")
	if err != nil {
		return err
	}
	funcMap := maps.Clone(sprig.TxtFuncMap())
	funcMap["targetRelPath"] = func(source string) string {
		return chezmoi.NewSourceRelPath(source).TargetRelPath(c.encryption.EncryptedSuffix()).String()
	}
	templateOptions := chezmoi.TemplateOptions{
		Options: append([]string(nil), c.Template.Options...),
	}
	commitMessageTmpl, err := chezmoi.ParseTemplate("commit_message", commitMessageTemplate, funcMap, templateOptions)
	if err != nil {
		return err
	}
	commitMessage, err := commitMessageTmpl.Execute(status)
	if err != nil {
		return err
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"commit", "--message", string(commitMessage)})
}

// gitAutoPush pushes all changes to the remote if status is not empty.
func (c *Config) gitAutoPush(status *git.Status) error {
	if status.Empty() {
		return nil
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"push"})
}

// makeRunEWithSourceState returns a function for
// github.com/spf13/cobra.Command.RunE that includes reading the source state.
func (c *Config) makeRunEWithSourceState(
	runE func(*cobra.Command, []string, *chezmoi.SourceState) error,
) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sourceState, err := c.getSourceState(cmd.Context(), cmd)
		if err != nil {
			return err
		}
		return runE(cmd, args, sourceState)
	}
}

// marshal formats data in dataFormat and writes it to the standard output.
func (c *Config) marshal(dataFormat writeDataFormat, data any) error {
	marshaledData, err := dataFormat.Format().Marshal(data)
	if err != nil {
		return err
	}
	return c.writeOutput(marshaledData)
}

// newRootCmd returns a new root github.com/spf13/cobra.Command.
func (c *Config) newRootCmd() (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:                "chezmoi",
		Short:              "Manage your dotfiles across multiple diverse machines, securely",
		Version:            c.versionStr,
		PersistentPreRunE:  c.persistentPreRunRootE,
		PersistentPostRunE: c.persistentPostRunRootE,
		SilenceErrors:      true,
		SilenceUsage:       true,
	}

	persistentFlags := rootCmd.PersistentFlags()

	persistentFlags.Var(&c.CacheDirAbsPath, "cache", "Set cache directory")
	persistentFlags.Var(&c.Color, "color", "Colorize output")
	persistentFlags.VarP(&c.DestDirAbsPath, "destination", "D", "Set destination directory")
	persistentFlags.Var(&c.Mode, "mode", "Mode")
	persistentFlags.Var(&c.persistentStateAbsPath, "persistent-state", "Set persistent state file")
	persistentFlags.Var(&c.Progress, "progress", "Display progress bars")
	persistentFlags.BoolVar(&c.Safe, "safe", c.Safe, "Safely replace files and symlinks")
	persistentFlags.VarP(&c.SourceDirAbsPath, "source", "S", "Set source directory")
	persistentFlags.Var(&c.UseBuiltinAge, "use-builtin-age", "Use builtin age")
	persistentFlags.Var(&c.UseBuiltinGit, "use-builtin-git", "Use builtin git")
	persistentFlags.BoolVarP(&c.Verbose, "verbose", "v", c.Verbose, "Make output more verbose")
	persistentFlags.VarP(&c.WorkingTreeAbsPath, "working-tree", "W", "Set working tree directory")

	persistentFlags.VarP(&c.configFileAbsPath, "config", "c", "Set config file")
	persistentFlags.Var(&c.configFormat, "config-format", "Set config file format")
	persistentFlags.Var(&c.cpuProfile, "cpu-profile", "Write a CPU profile to path")
	persistentFlags.BoolVar(&c.debug, "debug", c.debug, "Include debug information in output")
	persistentFlags.BoolVarP(&c.dryRun, "dry-run", "n", c.dryRun, "Do not make any modifications to the destination directory") //nolint:lll
	persistentFlags.BoolVar(&c.force, "force", c.force, "Make all changes without prompting")
	persistentFlags.BoolVar(&c.interactive, "interactive", c.interactive, "Prompt for all changes")
	persistentFlags.BoolVarP(&c.keepGoing, "keep-going", "k", c.keepGoing, "Keep going as far as possible after an error")
	persistentFlags.BoolVar(&c.noPager, "no-pager", c.noPager, "Do not use the pager")
	persistentFlags.BoolVar(&c.noTTY, "no-tty", c.noTTY, "Do not attempt to get a TTY for prompts")
	persistentFlags.VarP(&c.outputAbsPath, "output", "o", "Write output to path instead of stdout")
	persistentFlags.VarP(&c.refreshExternals, "refresh-externals", "R", "Refresh external cache")
	persistentFlags.Lookup("refresh-externals").NoOptDefVal = chezmoi.RefreshExternalsAlways.String()
	persistentFlags.BoolVar(&c.sourcePath, "source-path", c.sourcePath, "Specify targets by source path")

	if err := multierr.Combine(
		rootCmd.MarkPersistentFlagFilename("config"),
		rootCmd.MarkPersistentFlagFilename("cpu-profile"),
		persistentFlags.MarkHidden("cpu-profile"),
		rootCmd.MarkPersistentFlagDirname("destination"),
		rootCmd.MarkPersistentFlagFilename("output"),
		persistentFlags.MarkHidden("safe"),
		rootCmd.MarkPersistentFlagDirname("source"),
		rootCmd.RegisterFlagCompletionFunc("color", autoBoolFlagCompletionFunc),
		rootCmd.RegisterFlagCompletionFunc("config-format", readDataFormatFlagCompletionFunc),
		rootCmd.RegisterFlagCompletionFunc("mode", chezmoi.ModeFlagCompletionFunc),
		rootCmd.RegisterFlagCompletionFunc("refresh-externals", chezmoi.RefreshExternalsFlagCompletionFunc),
		rootCmd.RegisterFlagCompletionFunc("use-builtin-age", autoBoolFlagCompletionFunc),
		rootCmd.RegisterFlagCompletionFunc("use-builtin-git", autoBoolFlagCompletionFunc),
	); err != nil {
		return nil, err
	}

	rootCmd.SetHelpCommand(c.newHelpCmd())
	for _, cmd := range []*cobra.Command{
		c.newAddCmd(),
		c.newApplyCmd(),
		c.newArchiveCmd(),
		c.newCatCmd(),
		c.newCatConfigCmd(),
		c.newCDCmd(),
		c.newChattrCmd(),
		c.newCompletionCmd(),
		c.newDataCmd(),
		c.newDecryptCommand(),
		c.newDiffCmd(),
		c.newDoctorCmd(),
		c.newDumpCmd(),
		c.newDumpConfigCmd(),
		c.newEditCmd(),
		c.newEditConfigCmd(),
		c.newEncryptCommand(),
		c.newExecuteTemplateCmd(),
		c.newForgetCmd(),
		c.newGenerateCmd(),
		c.newGitCmd(),
		c.newIgnoredCmd(),
		c.newImportCmd(),
		c.newInitCmd(),
		c.newInternalTestCmd(),
		c.newLicenseCmd(),
		c.newManagedCmd(),
		c.newMergeCmd(),
		c.newMergeAllCmd(),
		c.newPurgeCmd(),
		c.newReAddCmd(),
		c.newRemoveCmd(),
		c.newSecretCmd(),
		c.newSourcePathCmd(),
		c.newStateCmd(),
		c.newStatusCmd(),
		c.newTargetPathCmd(),
		c.newUnmanagedCmd(),
		c.newUpdateCmd(),
		c.newUpgradeCmd(),
		c.newVerifyCmd(),
		cobracompletefig.CreateCompletionSpecCommand(),
	} {
		if cmd != nil {
			rootCmd.AddCommand(cmd)
		}
	}

	return rootCmd, nil
}

// newDiffSystem returns a system that logs all changes to s to w using
// diff.command if set or the builtin git diff otherwise.
func (c *Config) newDiffSystem(s chezmoi.System, w io.Writer, dirAbsPath chezmoi.AbsPath) chezmoi.System {
	if c.Diff.useBuiltinDiff || c.Diff.Command == "" {
		options := &chezmoi.GitDiffSystemOptions{
			Color:          c.Color.Value(c.colorAutoFunc),
			Filter:         chezmoi.NewEntryTypeFilter(c.Diff.include.Bits(), c.Diff.Exclude.Bits()),
			Reverse:        c.Diff.Reverse,
			ScriptContents: c.Diff.ScriptContents,
			TextConvFunc:   c.TextConv.convert,
		}
		return chezmoi.NewGitDiffSystem(s, w, dirAbsPath, options)
	}
	options := &chezmoi.ExternalDiffSystemOptions{
		Filter:         chezmoi.NewEntryTypeFilter(c.Diff.include.Bits(), c.Diff.Exclude.Bits()),
		Reverse:        c.Diff.Reverse,
		ScriptContents: c.Diff.ScriptContents,
	}
	return chezmoi.NewExternalDiffSystem(s, c.Diff.Command, c.Diff.Args, c.DestDirAbsPath, options)
}

// newSourceState returns a new SourceState with options.
func (c *Config) newSourceState(
	ctx context.Context, cmd *cobra.Command, options ...chezmoi.SourceStateOption,
) (*chezmoi.SourceState, error) {
	if err := c.checkVersion(); err != nil {
		return nil, err
	}

	httpClient, err := c.getHTTPClient()
	if err != nil {
		return nil, err
	}

	sourceStateLogger := c.logger.With().Str(logComponentKey, logComponentValueSourceState).Logger()

	c.SourceDirAbsPath, err = c.getSourceDirAbsPath(nil)
	if err != nil {
		return nil, err
	}

	sourceState := chezmoi.NewSourceState(append([]chezmoi.SourceStateOption{
		chezmoi.WithBaseSystem(c.baseSystem),
		chezmoi.WithCacheDir(c.CacheDirAbsPath),
		chezmoi.WithDefaultTemplateDataFunc(func() map[string]any {
			return c.getTemplateDataMap(cmd)
		}),
		chezmoi.WithDestDir(c.DestDirAbsPath),
		chezmoi.WithEncryption(c.encryption),
		chezmoi.WithHTTPClient(httpClient),
		chezmoi.WithInterpreters(c.Interpreters),
		chezmoi.WithLogger(&sourceStateLogger),
		chezmoi.WithMode(c.Mode),
		chezmoi.WithPriorityTemplateData(c.Data),
		chezmoi.WithSourceDir(c.SourceDirAbsPath),
		chezmoi.WithSystem(c.sourceSystem),
		chezmoi.WithTemplateFuncs(c.templateFuncs),
		chezmoi.WithTemplateOptions(c.Template.Options),
		chezmoi.WithUmask(c.Umask),
		chezmoi.WithVersion(c.version),
	}, options...)...)

	if err := sourceState.Read(ctx, &chezmoi.ReadOptions{
		RefreshExternals: c.refreshExternals,
		ReadHTTPResponse: c.readHTTPResponse,
	}); err != nil {
		return nil, err
	}

	return sourceState, nil
}

// persistentPostRunRootE performs post-run actions for the root command.
func (c *Config) persistentPostRunRootE(cmd *cobra.Command, args []string) error {
	annotations := getAnnotations(cmd)

	if err := c.persistentState.Close(); err != nil {
		return err
	}

	if annotations.hasTag(modifiesConfigFile) {
		configFileContents, err := c.baseSystem.ReadFile(c.configFileAbsPath)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			err = nil
		case err != nil:
			// err is already set, do nothing.
		default:
			var format chezmoi.Format
			if format, err = chezmoi.FormatFromAbsPath(c.configFileAbsPath); err == nil {
				var config map[string]any
				if err = format.Unmarshal(configFileContents, &config); err != nil {
					// err is already set, do nothing.
				} else {
					err = c.decodeConfigMap(config, &ConfigFile{})
				}
			}
		}
		if err != nil {
			c.errorf("warning: %s: %v\n", c.configFileAbsPath, err)
		}
	}

	if annotations.hasTag(modifiesSourceDirectory) {
		var status *git.Status
		if c.Git.AutoAdd || c.Git.AutoCommit || c.Git.AutoPush {
			var err error
			status, err = c.gitAutoAdd()
			if err != nil {
				return err
			}
		}
		if c.Git.AutoCommit || c.Git.AutoPush {
			if err := c.gitAutoCommit(status); err != nil {
				return err
			}
		}
		if c.Git.AutoPush {
			if err := c.gitAutoPush(status); err != nil {
				return err
			}
		}
	}

	if c.restoreWindowsConsole != nil {
		if err := c.restoreWindowsConsole(); err != nil {
			return err
		}
	}

	if command := c.Hooks[cmd.Name()].Post; command.Command != "" {
		if err := c.run(c.homeDirAbsPath, command.Command, command.Args); err != nil {
			return err
		}
	}

	return nil
}

// pageOutputString writes output using cmdPager as the pager command.
func (c *Config) pageOutputString(output, cmdPager string) error {
	pager := firstNonEmptyString(cmdPager, c.Pager)
	if c.noPager || pager == "" {
		return c.writeOutputString(output)
	}

	// If the pager command contains any spaces, assume that it is a full
	// shell command to be executed via the user's shell. Otherwise, execute
	// it directly.
	var pagerCmd *exec.Cmd
	if strings.IndexFunc(pager, unicode.IsSpace) != -1 {
		shellCommand, _ := shell.CurrentUserShell()
		shellCommand, shellArgs, err := parseCommand(shellCommand, []string{"-c", pager})
		if err != nil {
			return err
		}
		pagerCmd = exec.Command(shellCommand, shellArgs...)
	} else {
		pagerCmd = exec.Command(pager)
	}
	pagerCmd.Stdin = bytes.NewBufferString(output)
	pagerCmd.Stdout = c.stdout
	pagerCmd.Stderr = c.stderr
	return chezmoilog.LogCmdRun(pagerCmd)
}

// persistentPreRunRootE performs pre-run actions for the root command.
func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	annotations := getAnnotations(cmd)

	// Add the completion template function. This needs cmd, so we can't do it
	// in newConfig.
	c.addTemplateFunc("completion", func(shell string) string {
		completion, err := completion(cmd, shell)
		if err != nil {
			panic(err)
		}
		return completion
	})

	// Enable CPU profiling if configured.
	if !c.cpuProfile.Empty() {
		f, err := os.Create(c.cpuProfile.String())
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
	}

	if runtime.GOOS == "windows" {
		var err error
		c.restoreWindowsConsole, err = termenv.EnableVirtualTerminalProcessing(termenv.DefaultOutput())
		if err != nil {
			return err
		}
	}

	// Save flags that were set on the command line. Skip some types as
	// spf13/pflag does not round trip them correctly.
	changedFlags := make(map[pflag.Value]string)
	brokenFlagTypes := map[string]bool{
		"stringToInt":    true,
		"stringToInt64":  true,
		"stringToString": true,
	}
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed && !brokenFlagTypes[flag.Value.Type()] {
			changedFlags[flag.Value] = flag.Value.String()
		}
	})

	// Read the config file.
	if annotations.hasTag(runsWithInvalidConfig) {
		if c.configFileAbsPathErr == nil {
			_ = c.readConfig()
		}
	} else {
		if c.configFileAbsPathErr != nil {
			return c.configFileAbsPathErr
		}
		if err := c.readConfig(); err != nil {
			return fmt.Errorf("invalid config: %s: %w", c.configFileAbsPath, err)
		}
	}

	// Restore flags that were set on the command line.
	for value, original := range changedFlags {
		if err := value.Set(original); err != nil {
			return err
		}
	}

	if c.force && c.interactive {
		return errors.New("the --force and --interactive flags are mutually exclusive")
	}

	// Configure the logger.
	log.Logger = log.Output(zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.Out = c.stderr
			w.NoColor = !c.Color.Value(c.colorAutoFunc)
			w.TimeFormat = time.RFC3339
		},
	))
	if c.debug {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
	c.logger = &log.Logger

	// Log basic information.
	c.logger.Info().
		Object("version", c.versionInfo).
		Strs("args", os.Args).
		Str("goVersion", runtime.Version()).
		Msg("persistentPreRunRootE")
	realSystem := chezmoi.NewRealSystem(c.fileSystem,
		chezmoi.RealSystemWithSafe(c.Safe),
		chezmoi.RealSystemWithScriptTempDir(c.ScriptTempDir),
	)
	c.baseSystem = realSystem
	if c.debug {
		systemLogger := c.logger.With().Str(logComponentKey, logComponentValueSystem).Logger()
		c.baseSystem = chezmoi.NewDebugSystem(c.baseSystem, &systemLogger)
	}

	// Set up the persistent state.
	switch persistentStateMode := annotations.persistentStateMode(); {
	case persistentStateMode == persistentStateModeEmpty:
		c.persistentState = chezmoi.NewMockPersistentState()
	case persistentStateMode == persistentStateModeReadOnly:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		c.persistentState, err = chezmoi.NewBoltPersistentState(
			c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadOnly,
		)
		if err != nil {
			return err
		}
	case persistentStateMode == persistentStateModeReadMockWrite:
		fallthrough
	case persistentStateMode == persistentStateModeReadWrite && c.dryRun:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		persistentState, err := chezmoi.NewBoltPersistentState(
			c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadOnly,
		)
		if err != nil {
			return err
		}
		dryRunPersistentState := chezmoi.NewMockPersistentState()
		if err := persistentState.CopyTo(dryRunPersistentState); err != nil {
			return err
		}
		if err := persistentState.Close(); err != nil {
			return err
		}
		c.persistentState = dryRunPersistentState
	case persistentStateMode == persistentStateModeReadWrite:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		c.persistentState, err = chezmoi.NewBoltPersistentState(
			c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadWrite,
		)
		if err != nil {
			return err
		}
	default:
		c.persistentState = chezmoi.NullPersistentState{}
	}
	if c.debug && c.persistentState != nil {
		persistentStateLogger := c.logger.With().Str(logComponentKey, logComponentValuePersistentState).Logger()
		c.persistentState = chezmoi.NewDebugPersistentState(c.persistentState, &persistentStateLogger)
	}

	// Set up the source and destination systems.
	c.sourceSystem = c.baseSystem
	c.destSystem = c.baseSystem
	if !annotations.hasTag(modifiesDestinationDirectory) {
		c.destSystem = chezmoi.NewReadOnlySystem(c.destSystem)
	}
	if !annotations.hasTag(modifiesSourceDirectory) {
		c.sourceSystem = chezmoi.NewReadOnlySystem(c.sourceSystem)
	}
	if c.dryRun {
		c.sourceSystem = chezmoi.NewDryRunSystem(c.sourceSystem)
		c.destSystem = chezmoi.NewDryRunSystem(c.destSystem)
	}
	if c.Verbose {
		c.sourceSystem = c.newDiffSystem(c.sourceSystem, c.stdout, c.SourceDirAbsPath)
		c.destSystem = c.newDiffSystem(c.destSystem, c.stdout, c.DestDirAbsPath)
	}

	if err := c.setEncryption(); err != nil {
		return err
	}

	// Create the config directory if needed.
	if annotations.hasTag(requiresConfigDirectory) {
		if err := chezmoi.MkdirAll(c.baseSystem, c.configFileAbsPath.Dir(), fs.ModePerm); err != nil {
			return err
		}
	}

	// Create the source directory if needed.
	if annotations.hasTag(createSourceDirectoryIfNeeded) {
		if err := chezmoi.MkdirAll(c.baseSystem, c.SourceDirAbsPath, fs.ModePerm); err != nil {
			return err
		}
	}

	// Verify that the source directory exists and is a directory, if needed.
	if annotations.hasTag(requiresSourceDirectory) {
		switch fileInfo, err := c.baseSystem.Stat(c.SourceDirAbsPath); {
		case err == nil && fileInfo.IsDir():
			// Do nothing.
		case err == nil:
			return fmt.Errorf("%s: not a directory", c.SourceDirAbsPath)
		default:
			return err
		}
	}

	// Create the runtime directory if needed.
	if annotations.hasTag(runsCommands) {
		if runtime.GOOS == "linux" && c.bds.RuntimeDir != "" {
			// Snap sets the $XDG_RUNTIME_DIR environment variable to
			// /run/user/$uid/snap.$snap_name, but does not create this
			// directory. Consequently, any spawned processes that need
			// $XDG_DATA_DIR will fail. As a work-around, create the directory
			// if it does not exist. See
			// https://forum.snapcraft.io/t/wayland-dconf-and-xdg-runtime-dir/186/13.
			if err := chezmoi.MkdirAll(c.baseSystem, chezmoi.NewAbsPath(c.bds.RuntimeDir), 0o700); err != nil {
				return err
			}
		}
	}

	// Determine the working tree directory if it is not configured.
	if c.WorkingTreeAbsPath.Empty() {
		workingTreeAbsPath := c.SourceDirAbsPath
	FOR:
		for {
			gitDirAbsPath := workingTreeAbsPath.JoinString(gogit.GitDirName)
			if fileInfo, err := c.baseSystem.Stat(gitDirAbsPath); err == nil && fileInfo.IsDir() {
				c.WorkingTreeAbsPath = workingTreeAbsPath
				break FOR
			}
			prevWorkingTreeDirAbsPath := workingTreeAbsPath
			workingTreeAbsPath = workingTreeAbsPath.Dir()
			if workingTreeAbsPath == c.homeDirAbsPath || workingTreeAbsPath.Len() >= prevWorkingTreeDirAbsPath.Len() {
				c.WorkingTreeAbsPath = c.SourceDirAbsPath
				break FOR
			}
		}
	}

	// Create the working tree directory if needed.
	if annotations.hasTag(requiresWorkingTree) {
		if _, err := c.SourceDirAbsPath.TrimDirPrefix(c.WorkingTreeAbsPath); err != nil {
			return err
		}
		if err := chezmoi.MkdirAll(c.baseSystem, c.WorkingTreeAbsPath, fs.ModePerm); err != nil {
			return err
		}
	}

	scriptEnv := os.Environ()
	templateData := c.getTemplateData(cmd)
	scriptEnv = append(scriptEnv, "CHEZMOI=1")
	for key, value := range map[string]string{
		"ARCH":          templateData.Arch,
		"ARGS":          strings.Join(templateData.Args, " "),
		"CACHE_DIR":     templateData.CacheDir.String(),
		"COMMAND":       templateData.Command,
		"CONFIG_FILE":   templateData.ConfigFile.String(),
		"EXECUTABLE":    templateData.Executable.String(),
		"FQDN_HOSTNAME": templateData.FQDNHostname,
		"GID":           templateData.GID,
		"GROUP":         templateData.Group,
		"HOME_DIR":      templateData.HomeDir.String(),
		"HOSTNAME":      templateData.Hostname,
		"OS":            templateData.OS,
		"SOURCE_DIR":    templateData.SourceDir.String(),
		"UID":           templateData.UID,
		"USERNAME":      templateData.Username,
		"WORKING_TREE":  templateData.WorkingTree.String(),
	} {
		scriptEnv = append(scriptEnv, "CHEZMOI_"+key+"="+value)
	}
	for groupKey, group := range map[string]map[string]any{
		"KERNEL":          templateData.Kernel,
		"OS_RELEASE":      templateData.OSRelease,
		"VERSION":         templateData.Version,
		"WINDOWS_VERSION": templateData.WindowsVersion,
	} {
		for key, value := range group {
			upperSnakeCaseKey := camelCaseToUpperSnakeCase(key)
			valueStr := fmt.Sprintf("%s", value)
			scriptEnv = append(scriptEnv, "CHEZMOI_"+groupKey+"_"+upperSnakeCaseKey+"="+valueStr)
		}
	}
	for key, value := range c.ScriptEnv {
		scriptEnv = append(scriptEnv, key+"="+value)
	}
	c.runEnv = scriptEnv
	realSystem.SetScriptEnv(scriptEnv)

	if command := c.Hooks[cmd.Name()].Pre; command.Command != "" {
		if err := c.run(c.homeDirAbsPath, command.Command, command.Args); err != nil {
			return err
		}
	}

	return nil
}

// persistentStateFile returns the absolute path to the persistent state file,
// returning the first persistent file found, and returning the default path if
// none are found.
func (c *Config) persistentStateFile() (chezmoi.AbsPath, error) {
	if !c.persistentStateAbsPath.Empty() {
		return c.persistentStateAbsPath, nil
	}
	if !c.configFileAbsPath.Empty() {
		return c.configFileAbsPath.Dir().Join(persistentStateFileRelPath), nil
	}
	for _, configDir := range c.bds.ConfigDirs {
		configDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(configDir, c.homeDirAbsPath)
		if err != nil {
			return chezmoi.EmptyAbsPath, err
		}
		persistentStateFile := configDirAbsPath.Join(chezmoiRelPath, persistentStateFileRelPath)
		if _, err := os.Stat(persistentStateFile.String()); err == nil {
			return persistentStateFile, nil
		}
	}
	defaultConfigFileAbsPath, err := c.defaultConfigFile(c.fileSystem, c.bds)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return defaultConfigFileAbsPath.Dir().Join(persistentStateFileRelPath), nil
}

// progressAutoFunc detects whether progress bars should be displayed.
func (c *Config) progressAutoFunc() bool {
	if stdout, ok := c.stdout.(*os.File); ok {
		return term.IsTerminal(int(stdout.Fd()))
	}
	return false
}

func (c *Config) newTemplateData(cmd *cobra.Command) *templateData {
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
	// On Windows, the user's group ID returned by user.Current() is an SID and
	// no further useful lookup is possible with Go's standard library.
	//
	// Since neither the username nor the group are likely widely used in
	// templates, leave these variables unset if their values cannot be
	// determined. Unset variables will trigger template errors if used,
	// alerting the user to the problem and allowing them to find alternative
	// solutions.
	var gid, group, uid, username string
	if currentUser, err := user.Current(); err == nil {
		gid = currentUser.Gid
		uid = currentUser.Uid
		username = currentUser.Username
		if runtime.GOOS != "windows" {
			if rawGroup, err := user.LookupGroupId(currentUser.Gid); err == nil {
				group = rawGroup.Name
			} else {
				c.logger.Info().
					Str("gid", currentUser.Gid).
					Err(err).
					Msg("user.LookupGroupId")
			}
		}
	} else {
		c.logger.Info().
			Err(err).
			Msg("user.Current")
		var ok bool
		username, ok = os.LookupEnv("USER")
		if !ok {
			c.logger.Info().
				Str("key", "USER").
				Bool("ok", ok).
				Msg("os.LookupEnv")
		}
	}

	fqdnHostname, err := chezmoi.FQDNHostname(c.fileSystem)
	if err != nil {
		c.logger.Info().
			Err(err).
			Msg("chezmoi.FQDNHostname")
	}
	hostname, _, _ := strings.Cut(fqdnHostname, ".")

	kernel, err := chezmoi.Kernel(c.fileSystem)
	if err != nil {
		c.logger.Info().
			Err(err).
			Msg("chezmoi.Kernel")
	}

	var osRelease map[string]any
	switch runtime.GOOS {
	case "openbsd", "windows":
		// Don't populate osRelease on OSes where /etc/os-release does not
		// exist.
	default:
		if rawOSRelease, err := chezmoi.OSRelease(c.fileSystem); err == nil {
			osRelease = upperSnakeCaseToCamelCaseMap(rawOSRelease)
		} else {
			c.logger.Info().
				Err(err).
				Msg("chezmoi.OSRelease")
		}
	}

	executable, _ := os.Executable()
	windowsVersion, _ := windowsVersion()
	sourceDirAbsPath, _ := c.getSourceDirAbsPath(nil)

	return &templateData{
		Arch:         runtime.GOARCH,
		Args:         os.Args,
		CacheDir:     c.CacheDirAbsPath,
		Command:      cmd.Name(),
		ConfigFile:   c.configFileAbsPath,
		Executable:   chezmoi.NewAbsPath(executable),
		FQDNHostname: fqdnHostname,
		GID:          gid,
		Group:        group,
		HomeDir:      c.homeDirAbsPath,
		Hostname:     hostname,
		Kernel:       kernel,
		OS:           runtime.GOOS,
		OSRelease:    osRelease,
		SourceDir:    sourceDirAbsPath,
		UID:          uid,
		Username:     username,
		Version: map[string]any{
			"builtBy": c.versionInfo.BuiltBy,
			"commit":  c.versionInfo.Commit,
			"date":    c.versionInfo.Date,
			"version": c.versionInfo.Version,
		},
		WindowsVersion: windowsVersion,
		WorkingTree:    c.WorkingTreeAbsPath,
	}
}

// readConfig reads the config file, if it exists.
func (c *Config) readConfig() error {
	switch err := c.decodeConfigFile(c.configFileAbsPath, &c.ConfigFile); {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	default:
		return err
	}
}

// resetSourceState clears the cached source state, if any.
func (c *Config) resetSourceState() {
	c.sourceState = nil
	c.sourceStateErr = nil
}

// run runs name with args in dir.
func (c *Config) run(dir chezmoi.AbsPath, name string, args []string) error {
	cmd := exec.Command(name, args...)
	if !dir.Empty() {
		dirRawAbsPath, err := c.baseSystem.RawPath(dir)
		if err != nil {
			return err
		}
		cmd.Dir = dirRawAbsPath.String()
	}
	cmd.Env = c.runEnv
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	return c.baseSystem.RunCmd(cmd)
}

// runEditor runs the configured editor with args.
func (c *Config) runEditor(args []string) error {
	if err := c.persistentState.Close(); err != nil {
		return err
	}
	editor, editorArgs, err := c.editor(args)
	if err != nil {
		return err
	}
	start := time.Now()
	err = c.run(chezmoi.EmptyAbsPath, editor, editorArgs)
	if runtime.GOOS != "windows" && c.Edit.MinDuration != 0 {
		if duration := time.Since(start); duration < c.Edit.MinDuration {
			c.errorf("warning: %s: returned in less than %s\n", shellQuoteCommand(editor, editorArgs), c.Edit.MinDuration)
		}
	}
	return err
}

// setEncryption configures c's encryption.
func (c *Config) setEncryption() error {
	switch c.Encryption {
	case "age":
		// Only use builtin age encryption if age encryption is explicitly
		// specified. Otherwise, chezmoi would fall back to using age encryption
		// (rather than no encryption) if age is not in $PATH, which leads to
		// error messages from the builtin age instead of error messages about
		// encryption not being configured.
		c.Age.UseBuiltin = c.UseBuiltinAge.Value(c.useBuiltinAgeAutoFunc)
		c.encryption = &c.Age
	case "gpg":
		c.encryption = &c.GPG
	case "":
		// Detect encryption if any non-default configuration is set, preferring
		// gpg for backwards compatibility.
		switch {
		case !reflect.DeepEqual(c.GPG, defaultGPGEncryptionConfig):
			c.encryption = &c.GPG
		case !reflect.DeepEqual(c.Age, defaultAgeEncryptionConfig):
			c.encryption = &c.Age
		default:
			c.encryption = chezmoi.NoEncryption{}
		}
	default:
		return fmt.Errorf("%s: unknown encryption", c.Encryption)
	}

	if c.debug {
		encryptionLogger := c.logger.With().Str(logComponentKey, logComponentValueEncryption).Logger()
		c.encryption = chezmoi.NewDebugEncryption(c.encryption, &encryptionLogger)
	}

	return nil
}

// sourceAbsPaths returns the source absolute paths for each target path in
// args.
func (c *Config) sourceAbsPaths(sourceState *chezmoi.SourceState, args []string) ([]chezmoi.AbsPath, error) {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return nil, err
	}
	sourceAbsPaths := make([]chezmoi.AbsPath, 0, len(targetRelPaths))
	for _, targetRelPath := range targetRelPaths {
		sourceAbsPath := c.SourceDirAbsPath.Join(sourceState.MustEntry(targetRelPath).SourceRelPath().RelPath())
		sourceAbsPaths = append(sourceAbsPaths, sourceAbsPath)
	}
	return sourceAbsPaths, nil
}

func (c *Config) targetRelPath(absPath chezmoi.AbsPath) (chezmoi.RelPath, error) {
	relPath, err := absPath.TrimDirPrefix(c.DestDirAbsPath)
	var notInAbsDirError *chezmoi.NotInAbsDirError
	if errors.As(err, &notInAbsDirError) {
		return chezmoi.EmptyRelPath, fmt.Errorf("%s: not in destination directory (%s)", absPath, c.DestDirAbsPath)
	}
	return relPath, err
}

type targetRelPathsOptions struct {
	mustBeInSourceState bool
	recursive           bool
}

// targetRelPaths returns the target relative paths for each target path in
// args. The returned paths are sorted and de-duplicated.
func (c *Config) targetRelPaths(
	sourceState *chezmoi.SourceState, args []string, options targetRelPathsOptions,
) (chezmoi.RelPaths, error) {
	targetRelPaths := make(chezmoi.RelPaths, 0, len(args))
	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		targetRelPath, err := c.targetRelPath(argAbsPath)
		if err != nil {
			return nil, err
		}
		if options.mustBeInSourceState {
			if sourceState.Get(targetRelPath) == nil {
				return nil, fmt.Errorf("%s: not in source state", arg)
			}
		}
		targetRelPaths = append(targetRelPaths, targetRelPath)
		if options.recursive {
			parentRelPath := targetRelPath
			// FIXME we should not call s.TargetRelPaths() here - risk of
			// accidentally quadratic
			for _, targetRelPath := range sourceState.TargetRelPaths() {
				if _, err := targetRelPath.TrimDirPrefix(parentRelPath); err == nil {
					targetRelPaths = append(targetRelPaths, targetRelPath)
				}
			}
		}
	}

	if len(targetRelPaths) == 0 {
		return nil, nil
	}

	// Sort and de-duplicate targetRelPaths in place.
	sort.Sort(targetRelPaths)
	n := 1
	for i := 1; i < len(targetRelPaths); i++ {
		if targetRelPaths[i] != targetRelPaths[i-1] {
			targetRelPaths[n] = targetRelPaths[i]
			n++
		}
	}
	return targetRelPaths[:n], nil
}

// targetRelPathsBySourcePath returns the target relative paths for each arg in
// args.
func (c *Config) targetRelPathsBySourcePath(
	sourceState *chezmoi.SourceState, args []string,
) ([]chezmoi.RelPath, error) {
	targetRelPaths := make([]chezmoi.RelPath, 0, len(args))
	targetRelPathsBySourceRelPath := make(map[chezmoi.RelPath]chezmoi.RelPath)
	_ = sourceState.ForEach(func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
		sourceRelPath := sourceStateEntry.SourceRelPath().RelPath()
		targetRelPathsBySourceRelPath[sourceRelPath] = targetRelPath
		return nil
	})
	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		sourceRelPath, err := argAbsPath.TrimDirPrefix(c.SourceDirAbsPath)
		if err != nil {
			return nil, err
		}
		targetRelPath, ok := targetRelPathsBySourceRelPath[sourceRelPath]
		if !ok {
			return nil, fmt.Errorf("%s: not in source state", arg)
		}
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	return targetRelPaths, nil
}

// targetValidArgs returns target completions for toComplete given args.
func (c *Config) targetValidArgs(
	cmd *cobra.Command, args []string, toComplete string,
) ([]string, cobra.ShellCompDirective) {
	if !c.Completion.Custom {
		return nil, cobra.ShellCompDirectiveDefault
	}

	toCompleteAbsPath, err := chezmoi.NewAbsPathFromExtPath(toComplete, c.homeDirAbsPath)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	sourceState, err := c.getSourceState(cmd.Context(), cmd)
	if err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	var completions []string
	if err := sourceState.ForEach(func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
		completion := c.DestDirAbsPath.Join(targetRelPath).String()
		if _, ok := sourceStateEntry.(*chezmoi.SourceStateDir); ok {
			completion += "/"
		}
		if strings.HasPrefix(completion, toCompleteAbsPath.String()) {
			completions = append(completions, completion)
		}
		return nil
	}); err != nil {
		cobra.CompErrorln(err.Error())
		return nil, cobra.ShellCompDirectiveError
	}

	if !filepath.IsAbs(toComplete) {
		wd, err := os.Getwd()
		if err != nil {
			cobra.CompErrorln(err.Error())
			return nil, cobra.ShellCompDirectiveError
		}
		for i, completion := range completions {
			completions[i] = strings.TrimPrefix(completion, wd+"/")
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// tempDir returns the temporary directory for the given key, creating it if
// needed.
func (c *Config) tempDir(key string) (chezmoi.AbsPath, error) {
	if tempDirAbsPath, ok := c.tempDirs[key]; ok {
		return tempDirAbsPath, nil
	}
	tempDir, err := os.MkdirTemp("", key)
	c.logger.Err(err).
		Str("tempDir", tempDir).
		Msg("MkdirTemp")
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	tempDirAbsPath := chezmoi.NewAbsPath(tempDir)
	c.tempDirs[key] = tempDirAbsPath
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempDir, 0o700); err != nil {
			return chezmoi.EmptyAbsPath, err
		}
	}
	return tempDirAbsPath, nil
}

// useBuiltinAgeAutoFunc detects whether the builtin age should be used.
func (c *Config) useBuiltinAgeAutoFunc() bool {
	if _, err := chezmoi.LookPath(c.Age.Command); err == nil {
		return false
	}
	return true
}

// useBuiltinGitAutoFunc detects whether the builtin git should be used.
func (c *Config) useBuiltinGitAutoFunc() bool {
	// useBuiltinGit is false by default on Solaris as it uses the unavailable
	// flock function.
	if runtime.GOOS == "solaris" {
		return false
	}
	if _, err := chezmoi.LookPath(c.Git.Command); err == nil {
		return false
	}
	return true
}

// writeOutput writes data to the configured output.
func (c *Config) writeOutput(data []byte) error {
	if c.outputAbsPath.Empty() || c.outputAbsPath == chezmoi.NewAbsPath("-") {
		_, err := c.stdout.Write(data)
		return err
	}
	return os.WriteFile(c.outputAbsPath.String(), data, 0o666) //nolint:gosec
}

// writeOutputString writes data to the configured output.
func (c *Config) writeOutputString(data string) error {
	return c.writeOutput([]byte(data))
}

func newConfigFile(bds *xdg.BaseDirectorySpecification) ConfigFile {
	return ConfigFile{
		// Global configuration.
		CacheDirAbsPath: chezmoi.NewAbsPath(bds.CacheHome).Join(chezmoiRelPath),
		Color: autoBool{
			auto: true,
		},
		Interpreters: defaultInterpreters,
		Pager:        os.Getenv("PAGER"),
		Progress: autoBool{
			auto: true,
		},
		PINEntry: pinEntryConfig{
			Options: pinEntryDefaultOptions,
		},
		Safe: true,
		Template: templateConfig{
			Options: chezmoi.DefaultTemplateOptions,
		},
		Umask: chezmoi.Umask,
		UseBuiltinAge: autoBool{
			auto: true,
		},
		UseBuiltinGit: autoBool{
			auto: true,
		},
		Warnings: warningsConfig{
			ConfigFileTemplateHasChanged: true,
		},

		// Password manager configurations.
		Bitwarden: bitwardenConfig{
			Command: "bw",
		},
		Dashlane: dashlaneConfig{
			Command: "dcli",
		},
		Ejson: ejsonConfig{
			KeyDir: firstNonEmptyString(os.Getenv("EJSON_KEYDIR"), "/opt/ejson/keys"),
		},
		Gopass: gopassConfig{
			Command: "gopass",
		},
		Keepassxc: keepassxcConfig{
			Command: "keepassxc-cli",
			Prompt:  true,
		},
		Keeper: keeperConfig{
			Command: "keeper",
		},
		Lastpass: lastpassConfig{
			Command: "lpass",
		},
		Onepassword: onepasswordConfig{
			Command: "op",
			Prompt:  true,
		},
		Pass: passConfig{
			Command: "pass",
		},
		Passhole: passholeConfig{
			Command: "ph",
			Prompt:  true,
		},
		RBW: rbwConfig{
			Command: "rbw",
		},
		Vault: vaultConfig{
			Command: "vault",
		},

		// Encryption configurations.
		Age: defaultAgeEncryptionConfig,
		GPG: defaultGPGEncryptionConfig,

		// Command configurations.
		Add: addCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		Diff: diffCmdConfig{
			Exclude:        chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			ScriptContents: true,
			include:        chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
		},
		Edit: editCmdConfig{
			Hardlink:    true,
			MinDuration: 1 * time.Second,
			filter:      chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
		},
		Format: writeDataFormatJSON,
		Git: gitCmdConfig{
			Command: "git",
		},
		GitHub: gitHubConfig{
			RefreshPeriod: 1 * time.Minute,
		},
		Merge: mergeCmdConfig{
			Command: "vimdiff",
		},
		Status: statusCmdConfig{
			Exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		Update: updateCmdConfig{
			RecurseSubmodules: true,
			apply:             true,
			filter:            chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive:         true,
		},
		Verify: verifyCmdConfig{
			Exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
	}
}

func parseCommand(command string, args []string) (string, []string, error) {
	// If command is found, then return it.
	if path, err := chezmoi.LookPath(command); err == nil {
		return path, args, nil
	}

	// Otherwise, if the command contains spaces, parse it as a shell command.
	if whitespaceRx.MatchString(command) {
		var words []*syntax.Word
		if err := syntax.NewParser().Words(strings.NewReader(command), func(word *syntax.Word) bool {
			words = append(words, word)
			return true
		}); err != nil {
			return "", nil, err
		}
		fields, err := expand.Fields(&expand.Config{
			Env: expand.FuncEnviron(os.Getenv),
		}, words...)
		if err != nil {
			return "", nil, err
		}
		return fields[0], append(fields[1:], args...), nil
	}

	// Fallback to the command only.
	return command, args, nil
}

// withVersionInfo sets the version information.
func withVersionInfo(versionInfo VersionInfo) configOption {
	return func(c *Config) error {
		var version *semver.Version
		var versionElems []string
		if versionInfo.Version != "" {
			var err error
			version, err = semver.NewVersion(strings.TrimPrefix(versionInfo.Version, "v"))
			if err != nil {
				return err
			}
			versionElems = append(versionElems, "v"+version.String())
		} else {
			versionElems = append(versionElems, "dev")
		}
		if versionInfo.Commit != "" {
			versionElems = append(versionElems, "commit "+versionInfo.Commit)
		}
		if versionInfo.Date != "" {
			date := versionInfo.Date
			if sec, err := strconv.ParseInt(date, 10, 64); err == nil {
				date = time.Unix(sec, 0).UTC().Format(time.RFC3339)
			}
			versionElems = append(versionElems, "built at "+date)
		}
		if versionInfo.BuiltBy != "" {
			versionElems = append(versionElems, "built by "+versionInfo.BuiltBy)
		}
		if version != nil {
			c.version = *version
		}
		c.versionInfo = versionInfo
		c.versionStr = strings.Join(versionElems, ", ")
		return nil
	}
}
