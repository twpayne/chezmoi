package cmd

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/pprof"
	"slices"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"github.com/mitchellh/mapstructure"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-xdg/v6"
	"github.com/zricethezav/gitleaks/v8/detect"
	"golang.org/x/term"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"

	"github.com/twpayne/chezmoi/v2/assets/templates"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoierrors"
	"github.com/twpayne/chezmoi/v2/internal/chezmoigit"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

// defaultSentinel is a string value used to indicate that the default value
// should be used. It is a string unlikely to be an actual value set by the
// user.
const defaultSentinel = "\x00"

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
	Args    []string `json:"args"    mapstructure:"args"    yaml:"args"`
}

type hookConfig struct {
	Pre  commandConfig `json:"pre"  mapstructure:"pre"  yaml:"pre"`
	Post commandConfig `json:"post" mapstructure:"post" yaml:"post"`
}

type templateConfig struct {
	Options []string `json:"options" mapstructure:"options" yaml:"options"`
}

type warningsConfig struct {
	ConfigFileTemplateHasChanged bool `json:"configFileTemplateHasChanged" mapstructure:"configFileTemplateHasChanged" yaml:"configFileTemplateHasChanged"`
}

// ConfigFile contains all data settable in the config file.
type ConfigFile struct {
	// Global configuration.
	CacheDirAbsPath        chezmoi.AbsPath                `json:"cacheDir"        mapstructure:"cacheDir"        yaml:"cacheDir"`
	Color                  autoBool                       `json:"color"           mapstructure:"color"           yaml:"color"`
	Data                   map[string]any                 `json:"data"            mapstructure:"data"            yaml:"data"`
	Env                    map[string]string              `json:"env"             mapstructure:"env"             yaml:"env"`
	Format                 writeDataFormat                `json:"format"          mapstructure:"format"          yaml:"format"`
	DestDirAbsPath         chezmoi.AbsPath                `json:"destDir"         mapstructure:"destDir"         yaml:"destDir"`
	GitHub                 gitHubConfig                   `json:"gitHub"          mapstructure:"gitHub"          yaml:"gitHub"`
	Hooks                  map[string]hookConfig          `json:"hooks"           mapstructure:"hooks"           yaml:"hooks"`
	Interpreters           map[string]chezmoi.Interpreter `json:"interpreters"    mapstructure:"interpreters"    yaml:"interpreters"`
	Mode                   chezmoi.Mode                   `json:"mode"            mapstructure:"mode"            yaml:"mode"`
	Pager                  string                         `json:"pager"           mapstructure:"pager"           yaml:"pager"`
	PersistentStateAbsPath chezmoi.AbsPath                `json:"persistentState" mapstructure:"persistentState" yaml:"persistentState"`
	PINEntry               pinEntryConfig                 `json:"pinentry"        mapstructure:"pinentry"        yaml:"pinentry"`
	Progress               autoBool                       `json:"progress"        mapstructure:"progress"        yaml:"progress"`
	Safe                   bool                           `json:"safe"            mapstructure:"safe"            yaml:"safe"`
	ScriptEnv              map[string]string              `json:"scriptEnv"       mapstructure:"scriptEnv"       yaml:"scriptEnv"`
	ScriptTempDir          chezmoi.AbsPath                `json:"scriptTempDir"   mapstructure:"scriptTempDir"   yaml:"scriptTempDir"`
	SourceDirAbsPath       chezmoi.AbsPath                `json:"sourceDir"       mapstructure:"sourceDir"       yaml:"sourceDir"`
	TempDir                chezmoi.AbsPath                `json:"tempDir"         mapstructure:"tempDir"         yaml:"tempDir"`
	Template               templateConfig                 `json:"template"        mapstructure:"template"        yaml:"template"`
	TextConv               textConv                       `json:"textConv"        mapstructure:"textConv"        yaml:"textConv"`
	Umask                  fs.FileMode                    `json:"umask"           mapstructure:"umask"           yaml:"umask"`
	UseBuiltinAge          autoBool                       `json:"useBuiltinAge"   mapstructure:"useBuiltinAge"   yaml:"useBuiltinAge"`
	UseBuiltinGit          autoBool                       `json:"useBuiltinGit"   mapstructure:"useBuiltinGit"   yaml:"useBuiltinGit"`
	Verbose                bool                           `json:"verbose"         mapstructure:"verbose"         yaml:"verbose"`
	Warnings               warningsConfig                 `json:"warnings"        mapstructure:"warnings"        yaml:"warnings"`
	WorkingTreeAbsPath     chezmoi.AbsPath                `json:"workingTree"     mapstructure:"workingTree"     yaml:"workingTree"`

	// Password manager configurations.
	AWSSecretsManager awsSecretsManagerConfig `json:"awsSecretsManager" mapstructure:"awsSecretsManager" yaml:"awsSecretsManager"`
	AzureKeyVault     azureKeyVaultConfig     `json:"azureKeyVault"     mapstructure:"azureKeyVault"     yaml:"azureKeyVault"`
	Bitwarden         bitwardenConfig         `json:"bitwarden"         mapstructure:"bitwarden"         yaml:"bitwarden"`
	BitwardenSecrets  bitwardenSecretsConfig  `json:"bitwardenSecrets"  mapstructure:"bitwardenSecrets"  yaml:"bitwardenSecrets"`
	Dashlane          dashlaneConfig          `json:"dashlane"          mapstructure:"dashlane"          yaml:"dashlane"`
	Doppler           dopplerConfig           `json:"doppler"           mapstructure:"doppler"           yaml:"doppler"`
	Ejson             ejsonConfig             `json:"ejson"             mapstructure:"ejson"             yaml:"ejson"`
	Gopass            gopassConfig            `json:"gopass"            mapstructure:"gopass"            yaml:"gopass"`
	HCPVaultSecrets   hcpVaultSecretConfig    `json:"hcpVaultSecrets"   mapstructure:"hcpVaultSecrets"   yaml:"hcpVaultSecrets"`
	Keepassxc         keepassxcConfig         `json:"keepassxc"         mapstructure:"keepassxc"         yaml:"keepassxc"`
	Keeper            keeperConfig            `json:"keeper"            mapstructure:"keeper"            yaml:"keeper"`
	Lastpass          lastpassConfig          `json:"lastpass"          mapstructure:"lastpass"          yaml:"lastpass"`
	Onepassword       onepasswordConfig       `json:"onepassword"       mapstructure:"onepassword"       yaml:"onepassword"`
	OnepasswordSDK    onepasswordSDKConfig    `json:"onepasswordSDK"    mapstructure:"onepasswordSDK"    yaml:"onepasswordSDK"` //nolint:tagliatelle
	Pass              passConfig              `json:"pass"              mapstructure:"pass"              yaml:"pass"`
	Passhole          passholeConfig          `json:"passhole"          mapstructure:"passhole"          yaml:"passhole"`
	RBW               rbwConfig               `json:"rbw"               mapstructure:"rbw"               yaml:"rbw"`
	Secret            secretConfig            `json:"secret"            mapstructure:"secret"            yaml:"secret"`
	Vault             vaultConfig             `json:"vault"             mapstructure:"vault"             yaml:"vault"`

	// Encryption configurations.
	Encryption string                `json:"encryption" mapstructure:"encryption" yaml:"encryption"`
	Age        chezmoi.AgeEncryption `json:"age"        mapstructure:"age"        yaml:"age"`
	GPG        chezmoi.GPGEncryption `json:"gpg"        mapstructure:"gpg"        yaml:"gpg"`

	// Command configurations.
	Add        addCmdConfig        `json:"add"        mapstructure:"add"        yaml:"add"`
	CD         cdCmdConfig         `json:"cd"         mapstructure:"cd"         yaml:"cd"`
	Completion completionCmdConfig `json:"completion" mapstructure:"completion" yaml:"completion"`
	Diff       diffCmdConfig       `json:"diff"       mapstructure:"diff"       yaml:"diff"`
	Edit       editCmdConfig       `json:"edit"       mapstructure:"edit"       yaml:"edit"`
	Git        gitCmdConfig        `json:"git"        mapstructure:"git"        yaml:"git"`
	Merge      mergeCmdConfig      `json:"merge"      mapstructure:"merge"      yaml:"merge"`
	Status     statusCmdConfig     `json:"status"     mapstructure:"status"     yaml:"status"`
	Update     updateCmdConfig     `json:"update"     mapstructure:"update"     yaml:"update"`
	Verify     verifyCmdConfig     `json:"verify"     mapstructure:"verify"     yaml:"verify"`
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
	useBuiltinDiff   bool

	// Password manager data.
	gitHub  gitHubData
	keyring keyringData

	// Command configurations, not settable in the config file.
	age             ageCmdConfig
	apply           applyCmdConfig
	archive         archiveCmdConfig
	chattr          chattrCmdConfig
	destroy         destroyCmdConfig
	doctor          doctorCmdConfig
	dump            dumpCmdConfig
	executeTemplate executeTemplateCmdConfig
	ignored         ignoredCmdConfig
	_import         importCmdConfig
	init            initCmdConfig
	managed         managedCmdConfig
	mergeAll        mergeAllCmdConfig
	purge           purgeCmdConfig
	reAdd           reAddCmdConfig
	secret          secretCmdConfig
	state           stateCmdConfig
	unmanaged       unmanagedCmdConfig
	upgrade         upgradeCmdConfig

	// Common configuration.
	interactiveTemplateFuncs interactiveTemplateFuncsConfig

	// Version information.
	version     semver.Version
	versionInfo VersionInfo
	versionStr  string

	// Configuration.
	fileSystem                  vfs.FS
	bds                         *xdg.BaseDirectorySpecification
	defaultConfigFileAbsPath    chezmoi.AbsPath
	defaultConfigFileAbsPathErr error
	customConfigFileAbsPath     chezmoi.AbsPath
	baseSystem                  chezmoi.System
	sourceSystem                chezmoi.System
	destSystem                  chezmoi.System
	persistentState             chezmoi.PersistentState
	httpClient                  *http.Client
	logger                      *slog.Logger

	// Computed configuration.
	commandDirAbsPath   chezmoi.AbsPath
	homeDirAbsPath      chezmoi.AbsPath
	encryption          chezmoi.Encryption
	sourceDirAbsPath    chezmoi.AbsPath
	sourceDirAbsPathErr error
	sourceState         *chezmoi.SourceState
	sourceStateErr      error
	templateData        *templateData
	gitleaksDetector    *detect.Detector
	gitleaksDetectorErr error

	stdin             io.Reader
	stdout            io.Writer
	stderr            io.Writer
	bufioReader       *bufio.Reader
	diffPagerCmdStdin io.WriteCloser
	diffPagerCmd      *exec.Cmd

	tempDirs map[string]chezmoi.AbsPath

	ioregData ioregData

	restoreWindowsConsole func() error
}

type templateData struct {
	arch              string
	args              []string
	cacheDir          chezmoi.AbsPath
	command           string
	commandDir        chezmoi.AbsPath
	config            map[string]any
	configFile        chezmoi.AbsPath
	executable        chezmoi.AbsPath
	fqdnHostname      string
	gid               string
	group             string
	homeDir           chezmoi.AbsPath
	hostname          string
	kernel            map[string]any
	os                string
	osRelease         map[string]any
	pathListSeparator string
	pathSeparator     string
	sourceDir         chezmoi.AbsPath
	uid               string
	username          string
	version           map[string]any
	windowsVersion    map[string]any
	workingTree       chezmoi.AbsPath
}

// A configOption sets and option on a Config.
type configOption func(*Config) error

type configState struct {
	ConfigTemplateContentsSHA256 chezmoi.HexBytes `json:"configTemplateContentsSHA256" yaml:"configTemplateContentsSHA256"` //nolint:tagliatelle
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

	commonFlagCompletionFuncs = map[string]func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective){
		"exclude":    chezmoi.EntryTypeSetFlagCompletionFunc,
		"format":     writeDataFormatFlagCompletionFunc,
		"include":    chezmoi.EntryTypeSetFlagCompletionFunc,
		"path-style": chezmoi.PathStyleFlagCompletionFunc,
		"secrets":    severityFlagCompletionFunc,
	}
)

// newConfig creates a new Config with the given options.
func newConfig(options ...configOption) (*Config, error) {
	userHomeDir, err := chezmoi.UserHomeDir()
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

	logger := slog.Default()

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
			pathStyle: chezmoi.PathStyleRelative,
		},
		mergeAll: mergeAllCmdConfig{
			recursive: true,
		},
		reAdd: reAddCmdConfig{
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		unmanaged: unmanagedCmdConfig{
			pathStyle: chezmoi.PathStyleSimple(chezmoi.PathStyleRelative),
		},

		// Configuration.
		fileSystem: vfs.OSFS,
		bds:        bds,
		logger:     logger,

		// Computed configuration.
		homeDirAbsPath: homeDirAbsPath,

		tempDirs: make(map[string]chezmoi.AbsPath),

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	// Override sprig template functions. Delete them from the template function
	// map first to avoid a duplicate function panic.
	delete(c.templateFuncs, "fromJson")
	delete(c.templateFuncs, "quote")
	delete(c.templateFuncs, "splitList")
	delete(c.templateFuncs, "squote")
	delete(c.templateFuncs, "toPrettyJson")

	// The completion template function is added in persistentPreRunRootE as
	// it needs a *cobra.Command, which we don't yet have.
	for key, value := range map[string]any{
		"awsSecretsManager":            c.awsSecretsManagerTemplateFunc,
		"awsSecretsManagerRaw":         c.awsSecretsManagerRawTemplateFunc,
		"azureKeyVault":                c.azureKeyVaultTemplateFunc,
		"bitwarden":                    c.bitwardenTemplateFunc,
		"bitwardenAttachment":          c.bitwardenAttachmentTemplateFunc,
		"bitwardenAttachmentByRef":     c.bitwardenAttachmentByRefTemplateFunc,
		"bitwardenFields":              c.bitwardenFieldsTemplateFunc,
		"bitwardenSecrets":             c.bitwardenSecretsTemplateFunc,
		"comment":                      c.commentTemplateFunc,
		"dashlaneNote":                 c.dashlaneNoteTemplateFunc,
		"dashlanePassword":             c.dashlanePasswordTemplateFunc,
		"decrypt":                      c.decryptTemplateFunc,
		"deleteValueAtPath":            c.deleteValueAtPathTemplateFunc,
		"doppler":                      c.dopplerTemplateFunc,
		"dopplerProjectJson":           c.dopplerProjectJSONTemplateFunc,
		"ejsonDecrypt":                 c.ejsonDecryptTemplateFunc,
		"ejsonDecryptWithKey":          c.ejsonDecryptWithKeyTemplateFunc,
		"encrypt":                      c.encryptTemplateFunc,
		"eqFold":                       c.eqFoldTemplateFunc,
		"findExecutable":               c.findExecutableTemplateFunc,
		"findOneExecutable":            c.findOneExecutableTemplateFunc,
		"fromIni":                      c.fromIniTemplateFunc,
		"fromJson":                     c.fromJsonTemplateFunc,
		"fromJsonc":                    c.fromJsoncTemplateFunc,
		"fromToml":                     c.fromTomlTemplateFunc,
		"fromYaml":                     c.fromYamlTemplateFunc,
		"gitHubKeys":                   c.gitHubKeysTemplateFunc,
		"gitHubLatestRelease":          c.gitHubLatestReleaseTemplateFunc,
		"gitHubLatestReleaseAssetURL":  c.gitHubLatestReleaseAssetURLTemplateFunc,
		"gitHubLatestTag":              c.gitHubLatestTagTemplateFunc,
		"gitHubRelease":                c.gitHubReleaseTemplateFunc,
		"gitHubReleaseAssetURL":        c.gitHubReleaseAssetURLTemplateFunc,
		"gitHubReleases":               c.gitHubReleasesTemplateFunc,
		"gitHubTags":                   c.gitHubTagsTemplateFunc,
		"glob":                         c.globTemplateFunc,
		"gopass":                       c.gopassTemplateFunc,
		"gopassRaw":                    c.gopassRawTemplateFunc,
		"hcpVaultSecret":               c.hcpVaultSecretTemplateFunc,
		"hcpVaultSecretJson":           c.hcpVaultSecretJSONTemplateFunc,
		"hexDecode":                    c.hexDecodeTemplateFunc,
		"hexEncode":                    c.hexEncodeTemplateFunc,
		"include":                      c.includeTemplateFunc,
		"includeTemplate":              c.includeTemplateTemplateFunc,
		"ioreg":                        c.ioregTemplateFunc,
		"isExecutable":                 c.isExecutableTemplateFunc,
		"joinPath":                     c.joinPathTemplateFunc,
		"jq":                           c.jqTemplateFunc,
		"keepassxc":                    c.keepassxcTemplateFunc,
		"keepassxcAttachment":          c.keepassxcAttachmentTemplateFunc,
		"keepassxcAttribute":           c.keepassxcAttributeTemplateFunc,
		"keeper":                       c.keeperTemplateFunc,
		"keeperDataFields":             c.keeperDataFieldsTemplateFunc,
		"keeperFindPassword":           c.keeperFindPasswordTemplateFunc,
		"keyring":                      c.keyringTemplateFunc,
		"lastpass":                     c.lastpassTemplateFunc,
		"lastpassRaw":                  c.lastpassRawTemplateFunc,
		"lookPath":                     c.lookPathTemplateFunc,
		"lstat":                        c.lstatTemplateFunc,
		"mozillaInstallHash":           c.mozillaInstallHashTemplateFunc,
		"onepassword":                  c.onepasswordTemplateFunc,
		"onepasswordDetailsFields":     c.onepasswordDetailsFieldsTemplateFunc,
		"onepasswordDocument":          c.onepasswordDocumentTemplateFunc,
		"onepasswordItemFields":        c.onepasswordItemFieldsTemplateFunc,
		"onepasswordRead":              c.onepasswordReadTemplateFunc,
		"onepasswordSDKItemsGet":       c.onepasswordSDKItemsGet,
		"onepasswordSDKSecretsResolve": c.onepasswordSDKSecretsResolve,
		"output":                       c.outputTemplateFunc,
		"pass":                         c.passTemplateFunc,
		"passFields":                   c.passFieldsTemplateFunc,
		"passRaw":                      c.passRawTemplateFunc,
		"passhole":                     c.passholeTemplateFunc,
		"pruneEmptyDicts":              c.pruneEmptyDictsTemplateFunc,
		"quote":                        c.quoteTemplateFunc,
		"quoteList":                    c.quoteListTemplateFunc,
		"rbw":                          c.rbwTemplateFunc,
		"rbwFields":                    c.rbwFieldsTemplateFunc,
		"replaceAllRegex":              c.replaceAllRegexTemplateFunc,
		"secret":                       c.secretTemplateFunc,
		"secretJSON":                   c.secretJSONTemplateFunc,
		"setValueAtPath":               c.setValueAtPathTemplateFunc,
		"splitList":                    c.splitListTemplateFunc,
		"squote":                       c.squoteTemplateFunc,
		"stat":                         c.statTemplateFunc,
		"toIni":                        c.toIniTemplateFunc,
		"toPrettyJson":                 c.toPrettyJsonTemplateFunc,
		"toToml":                       c.toTomlTemplateFunc,
		"toYaml":                       c.toYamlTemplateFunc,
		"vault":                        c.vaultTemplateFunc,
	} {
		c.addTemplateFunc(key, value)
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	c.commandDirAbsPath, err = chezmoi.NormalizePath(wd)
	if err != nil {
		return nil, err
	}
	c.homeDirAbsPath, err = chezmoi.NormalizePath(c.homeDir)
	if err != nil {
		return nil, err
	}
	c.defaultConfigFileAbsPath, c.defaultConfigFileAbsPathErr = c.defaultConfigFile(c.fileSystem, c.bds)
	c.SourceDirAbsPath, err = c.defaultSourceDir(c.fileSystem, c.bds)
	if err != nil {
		return nil, err
	}
	c.DestDirAbsPath = c.homeDirAbsPath
	c._import.destination = c.homeDirAbsPath

	return c, nil
}

func (c *Config) getConfigFileAbsPath() chezmoi.AbsPath {
	if c.customConfigFileAbsPath.Empty() {
		return c.defaultConfigFileAbsPath
	}
	return c.customConfigFileAbsPath
}

// Close closes resources associated with c.
func (c *Config) Close() error {
	errs := make([]error, 0, len(c.tempDirs))
	for _, tempDirAbsPath := range c.tempDirs {
		err := os.RemoveAll(tempDirAbsPath.String())
		chezmoilog.InfoOrError(c.logger, "RemoveAll", err,
			chezmoilog.Stringer("tempDir", tempDirAbsPath),
		)
		errs = append(errs, err)
	}
	pprof.StopCPUProfile()
	return chezmoierrors.Combine(errs...)
}

// addTemplateFunc adds the template function with the given key and value
// to c. It panics if there is already an existing template function with the
// same key.
func (c *Config) addTemplateFunc(key string, value any) {
	if _, ok := c.templateFuncs[key]; ok {
		panic(key + ": already defined")
	}
	c.templateFuncs[key] = value
}

type applyArgsOptions struct {
	cmd          *cobra.Command
	filter       *chezmoi.EntryTypeFilter
	init         bool
	parentDirs   bool
	recursive    bool
	umask        fs.FileMode
	preApplyFunc chezmoi.PreApplyFunc
}

// applyArgs is the core of all commands that make changes to a target system.
// It checks config file freshness, reads the source state, and then applies the
// source state for each target entry in args. If args is empty then the source
// state is applied to all target entries.
func (c *Config) applyArgs(
	ctx context.Context,
	targetSystem chezmoi.System,
	targetDirAbsPath chezmoi.AbsPath,
	args []string,
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
		currentConfigTemplateContentsSHA256 = sha256Sum(configTemplate.contents)
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
				configStateValue, err := chezmoi.FormatJSON.Marshal(configState{
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
			recursive: options.recursive,
		})
		if err != nil {
			return err
		}
	}

	if options.parentDirs {
		targetRelPaths = prependParentRelPaths(targetRelPaths)
	}

	applyOptions := chezmoi.ApplyOptions{
		Filter:       options.filter,
		PreApplyFunc: options.preApplyFunc,
		Umask:        options.umask,
	}

	keptGoingAfterErr := false
	for _, targetRelPath := range targetRelPaths {
		switch err := sourceState.Apply(targetSystem, c.destSystem, c.persistentState, targetDirAbsPath, targetRelPath, applyOptions); {
		case errors.Is(err, fs.SkipDir):
			continue
		case err != nil:
			err = fmt.Errorf("%s: %w", targetRelPath, err)
			if c.keepGoing {
				c.errorf("%v\n", err)
				keptGoingAfterErr = true
			} else {
				return err
			}
		}
	}

	switch err := sourceState.PostApply(targetSystem, c.persistentState, targetDirAbsPath, targetRelPaths); {
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
	return chezmoilog.LogCmdOutput(slog.Default(), cmd)
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
	// Refresh the source directory, as there might be a .chezmoiroot file and
	// the template data is set before .chezmoiroot is read.
	sourceDirAbsPath, err := c.getSourceDirAbsPath(&getSourceDirAbsPathOptions{
		refresh: true,
	})
	if err != nil {
		return err
	}
	c.templateData.sourceDir = sourceDirAbsPath
	os.Setenv("CHEZMOI_SOURCE_DIR", sourceDirAbsPath.String())

	// Find config template, execute it, and create config file.
	configTemplate, err := c.findConfigTemplate()
	if err != nil {
		return err
	}

	if configTemplate == nil {
		return c.persistentState.Delete(chezmoi.ConfigStateBucket, configStateKey)
	}

	configFileContents, err := c.createConfigFile(configTemplate.targetRelPath, configTemplate.contents, cmd)
	if err != nil {
		return err
	}

	// Validate the config file.
	var configFile ConfigFile
	if err := c.decodeConfigBytes(configTemplate.format, configFileContents, &configFile); err != nil {
		return fmt.Errorf("%s: %w", configTemplate.sourceAbsPath, err)
	}

	// Write the config.
	configPath := c.init.configPath
	if c.init.configPath.Empty() {
		if c.customConfigFileAbsPath.Empty() {
			configPath = chezmoi.NewAbsPath(c.bds.ConfigHome).Join(chezmoiRelPath, configTemplate.targetRelPath)
		} else {
			configPath = c.customConfigFileAbsPath
		}
	}
	if err := chezmoi.MkdirAll(c.baseSystem, configPath.Dir(), fs.ModePerm); err != nil {
		return err
	}
	if err := c.baseSystem.WriteFile(configPath, configFileContents, 0o600); err != nil {
		return err
	}

	configStateValue, err := chezmoi.FormatJSON.Marshal(configState{
		ConfigTemplateContentsSHA256: chezmoi.HexBytes(sha256Sum(configTemplate.contents)),
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

	if err := c.setEnvironmentVariables(); err != nil {
		return err
	}

	return nil
}

// createConfigFile creates a config file using a template and returns its
// contents.
func (c *Config) createConfigFile(filename chezmoi.RelPath, data []byte, cmd *cobra.Command) ([]byte, error) {
	// Clone funcMap and restore it after creating the config.
	// This ensures that the init template functions
	// are removed before "normal" template parsing.
	funcMap := make(template.FuncMap)
	chezmoi.RecursiveMerge(funcMap, c.templateFuncs)
	defer func() {
		c.templateFuncs = funcMap
	}()

	initTemplateFuncs := map[string]any{
		"exit":             c.exitInitTemplateFunc,
		"promptBool":       c.promptBoolInteractiveTemplateFunc,
		"promptBoolOnce":   c.promptBoolOnceInteractiveTemplateFunc,
		"promptChoice":     c.promptChoiceInteractiveTemplateFunc,
		"promptChoiceOnce": c.promptChoiceOnceInteractiveTemplateFunc,
		"promptInt":        c.promptIntInteractiveTemplateFunc,
		"promptIntOnce":    c.promptIntOnceInteractiveTemplateFunc,
		"promptString":     c.promptStringInteractiveTemplateFunc,
		"promptStringOnce": c.promptStringOnceInteractiveTemplateFunc,
		"stdinIsATTY":      c.stdinIsATTYInitTemplateFunc,
		"writeToStdout":    c.writeToStdout,
	}
	chezmoi.RecursiveMerge(c.templateFuncs, initTemplateFuncs)

	tmpl, err := chezmoi.ParseTemplate(filename.String(), data, c.templateFuncs, chezmoi.TemplateOptions{
		Options: slices.Clone(c.Template.Options),
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

		dirEntryNames := chezmoiset.NewWithCapacity[string](len(dirEntries))
		for _, dirEntry := range dirEntries {
			dirEntryNames.Add(dirEntry.Name())
		}

		var names []string
		for _, extension := range chezmoi.FormatExtensions {
			name := "chezmoi." + extension
			if dirEntryNames.Contains(name) {
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

	if configFile.Git.CommitMessageTemplate != "" && configFile.Git.CommitMessageTemplateFile != "" {
		return fmt.Errorf(
			"%s: cannot specify both git.commitMessageTemplate and git.commitMessageTemplateFile",
			configFileAbsPath,
		)
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
	return decoder.Decode(configMap)
}

// defaultPreApplyFunc is the default pre-apply function. If the target entry
// has changed since chezmoi last wrote it then it prompts the user for the
// action to take.
func (c *Config) defaultPreApplyFunc(
	targetRelPath chezmoi.RelPath,
	targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState,
) error {
	c.logger.Info("defaultPreApplyFunc",
		chezmoilog.Stringer("targetRelPath", targetRelPath),
		slog.Any("targetEntryState", targetEntryState),
		slog.Any("lastWrittenEntryState", lastWrittenEntryState),
		slog.Any("actualEntryState", actualEntryState),
	)

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
				err := c.diffFile(targetRelPath, actualContents, actualEntryState.Mode, targetContents, targetEntryState.Mode)
				if err != nil {
					return err
				}
			case choice == "yes":
				return nil
			case choice == "no":
				return fs.SkipDir
			case choice == "all":
				c.interactive = false
				return nil
			case choice == "quit":
				return chezmoi.ExitCodeError(0)
			default:
				panic(choice + ": unexpected choice")
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
			if err := c.diffFile(targetRelPath, actualContents, actualEntryState.Mode, targetContents, targetEntryState.Mode); err != nil {
				return err
			}
		case choice == "overwrite":
			return nil
		case choice == "all-overwrite":
			c.force = true
			return nil
		case choice == "skip":
			return fs.SkipDir
		case choice == "quit":
			return chezmoi.ExitCodeError(0)
		default:
			panic(choice + ": unexpected choice")
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
	onIgnoreFunc   func(chezmoi.RelPath)
	recursive      bool
}

// destAbsPathInfos returns the os/fs.FileInfos for each destination entry in
// args, recursing into subdirectories and following symlinks if configured in
// options.
func (c *Config) destAbsPathInfos(
	sourceState *chezmoi.SourceState,
	args []string,
	options destAbsPathInfosOptions,
) (map[chezmoi.AbsPath]fs.FileInfo, error) {
	destAbsPathInfos := make(map[chezmoi.AbsPath]fs.FileInfo)
	for _, arg := range args {
		arg = filepath.Clean(arg)
		destAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		targetRelPath, err := c.targetRelPath(destAbsPath)
		if err != nil {
			return nil, err
		}
		if sourceState.Ignore(targetRelPath) {
			options.onIgnoreFunc(targetRelPath)
			continue
		}
		if options.recursive {
			walkFunc := func(destAbsPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
				switch {
				case options.ignoreNotExist && errors.Is(err, fs.ErrNotExist):
					return nil
				case err != nil:
					return err
				}

				targetRelPath, err := c.targetRelPath(destAbsPath)
				if err != nil {
					return err
				}
				if sourceState.Ignore(targetRelPath) {
					options.onIgnoreFunc(targetRelPath)
					if fileInfo.IsDir() {
						return fs.SkipDir
					}
					return nil
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
	fromData []byte,
	fromMode fs.FileMode,
	toData []byte,
	toMode fs.FileMode,
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
	return c.pageDiffOutput(builder.String())
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
	editCommand = cmp.Or(os.Getenv("VISUAL"), os.Getenv("EDITOR"), defaultEditor)

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

	dirEntryNames := chezmoiset.NewWithCapacity[chezmoi.RelPath](len(dirEntries))
	for _, dirEntry := range dirEntries {
		dirEntryNames.Add(chezmoi.NewRelPath(dirEntry.Name()))
	}

	var configTemplates []*configTemplate //nolint:prealloc
	for _, extension := range chezmoi.FormatExtensions {
		relPath := chezmoi.NewRelPath(chezmoi.Prefix + "." + extension + chezmoi.TemplateSuffix)
		if !dirEntryNames.Contains(relPath) {
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
		sourceAbsPathStrs := make([]string, len(configTemplates))
		for i, configTemplate := range configTemplates {
			sourceAbsPathStr := configTemplate.sourceAbsPath.String()
			sourceAbsPathStrs[i] = sourceAbsPathStr
		}
		return nil, fmt.Errorf("multiple config file templates: %s ", englishList(sourceAbsPathStrs))
	}
}

// getDiffPager returns the pager for diff output.
func (c *Config) getDiffPager() string {
	switch {
	case c.noPager:
		return ""
	case c.Diff.Pager != defaultSentinel:
		return c.Diff.Pager
	default:
		return c.Pager
	}
}

// getDiffPagerCmd returns a command to run the diff pager, or nil if there is
// no diff pager configured.
func (c *Config) getDiffPagerCmd() (*exec.Cmd, error) {
	pager := c.getDiffPager()
	if pager == "" {
		return nil, nil
	}

	// If the pager command contains any spaces, assume that it is a full
	// shell command to be executed via the user's shell. Otherwise, execute
	// it directly.
	var pagerCmd *exec.Cmd
	if strings.IndexFunc(pager, unicode.IsSpace) != -1 {
		shellCommand, _ := shell.CurrentUserShell()
		shellCommand, shellArgs, err := parseCommand(shellCommand, []string{"-c", pager})
		if err != nil {
			return nil, err
		}
		pagerCmd = exec.Command(shellCommand, shellArgs...)
	} else {
		pagerCmd = exec.Command(pager)
	}
	pagerCmd.Stdout = c.stdout
	pagerCmd.Stderr = c.stderr
	return pagerCmd, nil
}

func (c *Config) getGitleaksDetector() (*detect.Detector, error) {
	if c.gitleaksDetector == nil && c.gitleaksDetectorErr == nil {
		c.gitleaksDetector, c.gitleaksDetectorErr = detect.NewDetectorDefaultConfig()
	}
	return c.gitleaksDetector, c.gitleaksDetectorErr
}

// A modifyHTTPRequestFunc is a function that modifies a net/http.Request before
// it is sent.
type modifyHTTPRequestFunc func(*http.Request) (*http.Request, error)

// A modifyHTTPRequestRoundTripper is a net/http.Transport that modifies the
// request before it is sent.
type modifyHTTPRequestRoundTripper struct {
	modifyHTTPRequestFunc modifyHTTPRequestFunc
	httpRoundTripper      http.RoundTripper
}

func newModifyHTTPRequestRoundTripper(f modifyHTTPRequestFunc, t http.RoundTripper) modifyHTTPRequestRoundTripper {
	return modifyHTTPRequestRoundTripper{
		modifyHTTPRequestFunc: f,
		httpRoundTripper:      t,
	}
}

// RoundTrip implements net/http.Transport.RoundTrip.
func (m modifyHTTPRequestRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	modifiedRequest, err := m.modifyHTTPRequestFunc(req)
	if err != nil {
		return nil, err
	}
	return m.httpRoundTripper.RoundTrip(modifiedRequest)
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

	c.httpClient.Transport = newModifyHTTPRequestRoundTripper(
		func(req *http.Request) (*http.Request, error) {
			req = req.Clone(req.Context())
			req.Header.Add("User-Agent", "chezmoi.io/"+c.version.String())
			return req, nil
		},
		c.httpClient.Transport,
	)

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
			"arch":              templateData.arch,
			"args":              templateData.args,
			"cacheDir":          templateData.cacheDir.String(),
			"command":           templateData.command,
			"commandDir":        templateData.commandDir.String(),
			"config":            templateData.config,
			"configFile":        templateData.configFile.String(),
			"executable":        templateData.executable.String(),
			"fqdnHostname":      templateData.fqdnHostname,
			"gid":               templateData.gid,
			"group":             templateData.group,
			"homeDir":           templateData.homeDir.String(),
			"hostname":          templateData.hostname,
			"kernel":            templateData.kernel,
			"os":                templateData.os,
			"osRelease":         templateData.osRelease,
			"pathListSeparator": templateData.pathListSeparator,
			"pathSeparator":     templateData.pathSeparator,
			"sourceDir":         templateData.sourceDir.String(),
			"uid":               templateData.uid,
			"username":          templateData.username,
			"version":           templateData.version,
			"windowsVersion":    templateData.windowsVersion,
			"workingTree":       templateData.workingTree.String(),
		},
	}
}

// gitAutoAdd adds all changes to the git index and returns the new git status.
func (c *Config) gitAutoAdd() (*chezmoigit.Status, error) {
	if err := c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"add", "."}); err != nil {
		return nil, err
	}
	output, err := c.cmdOutput(c.WorkingTreeAbsPath, c.Git.Command, []string{"status", "--porcelain=v2"})
	if err != nil {
		return nil, err
	}
	return chezmoigit.ParseStatusPorcelainV2(output)
}

// gitAutoCommit commits all changes in the git index, including generating a
// commit message from status.
func (c *Config) gitAutoCommit(cmd *cobra.Command, status *chezmoigit.Status) error {
	if status.Empty() {
		return nil
	}
	commitMessage, err := c.gitCommitMessage(cmd, status)
	if err != nil {
		return err
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"commit", "--message", string(commitMessage)})
}

// gitAutoPush pushes all changes to the remote if status is not empty.
func (c *Config) gitAutoPush(status *chezmoigit.Status) error {
	if status.Empty() {
		return nil
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"push"})
}

// gitCommitMessage returns the git commit message for the given status.
func (c *Config) gitCommitMessage(cmd *cobra.Command, status *chezmoigit.Status) ([]byte, error) {
	funcMap := maps.Clone(c.templateFuncs)
	maps.Copy(funcMap, map[string]any{
		"promptBool":   c.promptBoolInteractiveTemplateFunc,
		"promptChoice": c.promptChoiceInteractiveTemplateFunc,
		"promptInt":    c.promptIntInteractiveTemplateFunc,
		"promptString": c.promptStringInteractiveTemplateFunc,
		"targetRelPath": func(source string) string {
			return chezmoi.NewSourceRelPath(source).TargetRelPath(c.encryption.EncryptedSuffix()).String()
		},
	})
	var name string
	var commitMessageTemplateData []byte
	switch {
	case c.Git.CommitMessageTemplate != "":
		name = "git.commitMessageTemplate"
		commitMessageTemplateData = []byte(c.Git.CommitMessageTemplate)
	case c.Git.CommitMessageTemplateFile != "":
		if c.sourceDirAbsPathErr != nil {
			return nil, c.sourceDirAbsPathErr
		}
		commitMessageTemplateFileAbsPath := c.sourceDirAbsPath.JoinString(c.Git.CommitMessageTemplateFile)
		name = c.sourceDirAbsPath.String()
		var err error
		commitMessageTemplateData, err = c.baseSystem.ReadFile(commitMessageTemplateFileAbsPath)
		if err != nil {
			return nil, err
		}
	default:
		name = "COMMIT_MESSAGE"
		commitMessageTemplateData = []byte(templates.CommitMessageTmpl)
	}
	commitMessageTmpl, err := chezmoi.ParseTemplate(name, commitMessageTemplateData, funcMap, chezmoi.TemplateOptions{
		Options: slices.Clone(c.Template.Options),
	})
	if err != nil {
		return nil, err
	}
	sourceState, err := c.getSourceState(cmd.Context(), cmd)
	if err != nil {
		return nil, err
	}
	templateDataMap := sourceState.TemplateData()
	templateDataMap["chezmoi"].(map[string]any)["status"] = status //nolint:forcetypeassert
	return commitMessageTmpl.Execute(templateDataMap)
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
	persistentFlags.Var(&c.PersistentStateAbsPath, "persistent-state", "Set persistent state file")
	persistentFlags.Var(&c.Progress, "progress", "Display progress bars")
	persistentFlags.BoolVar(&c.Safe, "safe", c.Safe, "Safely replace files and symlinks")
	persistentFlags.VarP(&c.SourceDirAbsPath, "source", "S", "Set source directory")
	persistentFlags.Var(&c.UseBuiltinAge, "use-builtin-age", "Use builtin age")
	persistentFlags.Var(&c.UseBuiltinGit, "use-builtin-git", "Use builtin git")
	persistentFlags.BoolVarP(&c.Verbose, "verbose", "v", c.Verbose, "Make output more verbose")
	persistentFlags.VarP(&c.WorkingTreeAbsPath, "working-tree", "W", "Set working tree directory")

	persistentFlags.VarP(&c.customConfigFileAbsPath, "config", "c", "Set config file")
	persistentFlags.Var(&c.configFormat, "config-format", "Set config file format")
	persistentFlags.Var(&c.cpuProfile, "cpu-profile", "Write a CPU profile to path")
	persistentFlags.BoolVar(&c.debug, "debug", c.debug, "Include debug information in output")
	persistentFlags.BoolVarP(&c.dryRun, "dry-run", "n", c.dryRun, "Do not make any modifications to the destination directory")
	persistentFlags.BoolVar(&c.force, "force", c.force, "Make all changes without prompting")
	persistentFlags.BoolVar(&c.interactive, "interactive", c.interactive, "Prompt for all changes")
	persistentFlags.BoolVarP(&c.keepGoing, "keep-going", "k", c.keepGoing, "Keep going as far as possible after an error")
	persistentFlags.BoolVar(&c.noPager, "no-pager", c.noPager, "Do not use the pager")
	persistentFlags.BoolVar(&c.noTTY, "no-tty", c.noTTY, "Do not attempt to get a TTY for prompts")
	persistentFlags.VarP(&c.outputAbsPath, "output", "o", "Write output to path instead of stdout")
	persistentFlags.VarP(&c.refreshExternals, "refresh-externals", "R", "Refresh external cache")
	persistentFlags.Lookup("refresh-externals").NoOptDefVal = chezmoi.RefreshExternalsAlways.String()
	persistentFlags.BoolVar(&c.sourcePath, "source-path", c.sourcePath, "Specify targets by source path")
	persistentFlags.BoolVarP(&c.useBuiltinDiff, "use-builtin-diff", "", c.useBuiltinDiff, "Use builtin diff")

	if err := chezmoierrors.Combine(
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
		c.newAgeCmd(),
		c.newApplyCmd(),
		c.newArchiveCmd(),
		c.newCatCmd(),
		c.newCatConfigCmd(),
		c.newCDCmd(),
		c.newChattrCmd(),
		c.newCompletionCmd(),
		c.newDataCmd(),
		c.newDecryptCommand(),
		c.newDestroyCmd(),
		c.newDiffCmd(),
		c.newDoctorCmd(),
		c.newDumpCmd(),
		c.newDumpConfigCmd(),
		c.newEditCmd(),
		c.newEditConfigCmd(),
		c.newEditConfigTemplateCmd(),
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
		c.newMackupCmd(),
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
	} {
		if cmd != nil {
			ensureAllFlagsDocumented(cmd, persistentFlags)
			registerCommonFlagCompletionFuncs(cmd)
			rootCmd.AddCommand(cmd)
		}
	}

	return rootCmd, nil
}

// newDiffSystem returns a system that logs all changes to s to w using
// diff.command if set or the builtin git diff otherwise.
func (c *Config) newDiffSystem(s chezmoi.System, w io.Writer, dirAbsPath chezmoi.AbsPath) chezmoi.System {
	if c.useBuiltinDiff || c.Diff.Command == "" {
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
	ctx context.Context,
	cmd *cobra.Command,
	options ...chezmoi.SourceStateOption,
) (*chezmoi.SourceState, error) {
	if err := c.checkVersion(); err != nil {
		return nil, err
	}

	httpClient, err := c.getHTTPClient()
	if err != nil {
		return nil, err
	}

	sourceStateLogger := c.logger.With(logComponentKey, logComponentValueSourceState)

	c.SourceDirAbsPath, err = c.getSourceDirAbsPath(nil)
	if err != nil {
		return nil, err
	}

	if err := c.runHookPre(readSourceStateHookName); err != nil {
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
		chezmoi.WithLogger(sourceStateLogger),
		chezmoi.WithMode(c.Mode),
		chezmoi.WithPriorityTemplateData(c.Data),
		chezmoi.WithScriptTempDir(c.ScriptTempDir),
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

	if err := c.runHookPost(readSourceStateHookName); err != nil {
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

	// Close any connection to keepassxc-cli.
	if err := c.keepassxcClose(); err != nil {
		return err
	}

	// Wait for any diff pager process to terminate.
	if c.diffPagerCmd != nil {
		if err := c.diffPagerCmdStdin.Close(); err != nil {
			return err
		}
		if c.diffPagerCmd.Process != nil {
			if err := chezmoilog.LogCmdWait(c.logger, c.diffPagerCmd); err != nil {
				return err
			}
		}
	}

	if annotations.hasTag(modifiesConfigFile) {
		configFileContents, err := c.baseSystem.ReadFile(c.getConfigFileAbsPath())
		switch {
		case errors.Is(err, fs.ErrNotExist):
			err = nil
		case err != nil:
			// err is already set, do nothing.
		default:
			var format chezmoi.Format
			if format, err = chezmoi.FormatFromAbsPath(c.getConfigFileAbsPath()); err == nil {
				var config map[string]any
				if err = format.Unmarshal(configFileContents, &config); err != nil { //nolint:revive
					// err is already set, do nothing.
				} else {
					err = c.decodeConfigMap(config, &ConfigFile{})
				}
			}
		}
		if err != nil {
			c.errorf("warning: %s: %v\n", c.getConfigFileAbsPath(), err)
		}
	}

	if annotations.hasTag(modifiesSourceDirectory) {
		var status *chezmoigit.Status
		if c.Git.AutoAdd || c.Git.AutoCommit || c.Git.AutoPush {
			var err error
			status, err = c.gitAutoAdd()
			if err != nil {
				return err
			}
		}
		if c.Git.AutoCommit || c.Git.AutoPush {
			if err := c.gitAutoCommit(cmd, status); err != nil {
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

	if err := c.runHookPost(cmd.Name()); err != nil {
		return err
	}

	return nil
}

// pageDiffOutput pages the diff output to stdout.
func (c *Config) pageDiffOutput(output string) error {
	switch pagerCmd, err := c.getDiffPagerCmd(); {
	case err != nil:
		return err
	case pagerCmd == nil:
		return c.writeOutputString(output)
	default:
		pagerCmd.Stdin = bytes.NewBufferString(output)
		return chezmoilog.LogCmdRun(c.logger, pagerCmd)
	}
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
	if annotations.hasTag(doesNotRequireValidConfig) {
		if c.defaultConfigFileAbsPathErr == nil {
			_ = c.readConfig()
		}
	} else {
		if c.defaultConfigFileAbsPathErr != nil {
			return c.defaultConfigFileAbsPathErr
		}
		if err := c.readConfig(); err != nil {
			return fmt.Errorf("invalid config: %s: %w", c.getConfigFileAbsPath(), err)
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
	var handler slog.Handler
	if c.debug {
		handler = slog.NewTextHandler(c.stderr, nil)
	} else {
		handler = chezmoilog.NullHandler{}
	}
	c.logger = slog.New(handler)
	slog.SetDefault(c.logger)

	// Log basic information.
	c.logger.Info("persistentPreRunRootE",
		slog.Any("version", c.versionInfo),
		slog.Any("args", os.Args),
		slog.String("goVersion", runtime.Version()),
	)
	realSystem := chezmoi.NewRealSystem(c.fileSystem,
		chezmoi.RealSystemWithSafe(c.Safe),
		chezmoi.RealSystemWithScriptTempDir(c.ScriptTempDir),
	)
	c.baseSystem = realSystem
	if c.debug {
		systemLogger := c.logger.With(slog.String(logComponentKey, logComponentValueSystem))
		c.baseSystem = chezmoi.NewDebugSystem(c.baseSystem, systemLogger)
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
			c.baseSystem,
			persistentStateFileAbsPath,
			chezmoi.BoltPersistentStateReadOnly,
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
			c.baseSystem,
			persistentStateFileAbsPath,
			chezmoi.BoltPersistentStateReadOnly,
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
			c.baseSystem,
			persistentStateFileAbsPath,
			chezmoi.BoltPersistentStateReadWrite,
		)
		if err != nil {
			return err
		}
	default:
		c.persistentState = chezmoi.NullPersistentState{}
	}
	if c.debug && c.persistentState != nil {
		persistentStateLogger := c.logger.With(slog.String(logComponentKey, logComponentValuePersistentState))
		c.persistentState = chezmoi.NewDebugPersistentState(c.persistentState, persistentStateLogger)
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
	if c.dryRun || annotations.hasTag(dryRun) {
		c.sourceSystem = chezmoi.NewDryRunSystem(c.sourceSystem)
		c.destSystem = chezmoi.NewDryRunSystem(c.destSystem)
	}
	if annotations.hasTag(outputsDiff) ||
		c.Verbose && (annotations.hasTag(modifiesDestinationDirectory) || annotations.hasTag(modifiesSourceDirectory)) {
		// If the user has configured a diff pager, then start it as a process.
		// Otherwise, write the diff output directly to stdout.
		var writer io.Writer
		switch pagerCmd, err := c.getDiffPagerCmd(); {
		case err != nil:
			return err
		case pagerCmd == nil:
			writer = c.stdout
		default:
			pipeReader, pipeWriter := io.Pipe()
			pagerCmd.Stdin = pipeReader
			lazyWriter := newLazyWriter(func() (io.WriteCloser, error) {
				if err := chezmoilog.LogCmdStart(c.logger, pagerCmd); err != nil {
					return nil, err
				}
				return pipeWriter, nil
			})
			writer = lazyWriter
			c.diffPagerCmd = pagerCmd
			c.diffPagerCmdStdin = lazyWriter
		}
		c.sourceSystem = c.newDiffSystem(c.sourceSystem, writer, c.SourceDirAbsPath)
		c.destSystem = c.newDiffSystem(c.destSystem, writer, c.DestDirAbsPath)
	}

	if err := c.setEncryption(); err != nil {
		return err
	}

	// Create the config directory if needed.
	if annotations.hasTag(requiresConfigDirectory) {
		if err := chezmoi.MkdirAll(c.baseSystem, c.getConfigFileAbsPath().Dir(), fs.ModePerm); err != nil {
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
			gitDirAbsPath := workingTreeAbsPath.JoinString(git.GitDirName)
			if _, err := c.baseSystem.Stat(gitDirAbsPath); err == nil {
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

	templateData := c.getTemplateData(cmd)
	os.Setenv("CHEZMOI", "1")
	for key, value := range map[string]string{
		"ARCH":          templateData.arch,
		"ARGS":          strings.Join(templateData.args, " "),
		"CACHE_DIR":     templateData.cacheDir.String(),
		"COMMAND":       templateData.command,
		"COMMAND_DIR":   templateData.commandDir.String(),
		"CONFIG_FILE":   templateData.configFile.String(),
		"EXECUTABLE":    templateData.executable.String(),
		"FQDN_HOSTNAME": templateData.fqdnHostname,
		"GID":           templateData.gid,
		"GROUP":         templateData.group,
		"HOME_DIR":      templateData.homeDir.String(),
		"HOSTNAME":      templateData.hostname,
		"OS":            templateData.os,
		"SOURCE_DIR":    templateData.sourceDir.String(),
		"UID":           templateData.uid,
		"USERNAME":      templateData.username,
		"WORKING_TREE":  templateData.workingTree.String(),
	} {
		os.Setenv("CHEZMOI_"+key, value)
	}
	if c.Verbose {
		os.Setenv("CHEZMOI_VERBOSE", "1")
	}
	for groupKey, group := range map[string]map[string]any{
		"KERNEL":          templateData.kernel,
		"OS_RELEASE":      templateData.osRelease,
		"VERSION":         templateData.version,
		"WINDOWS_VERSION": templateData.windowsVersion,
	} {
		for key, value := range group {
			key := "CHEZMOI_" + groupKey + "_" + camelCaseToUpperSnakeCase(key)
			valueStr := fmt.Sprintf("%s", value)
			os.Setenv(key, valueStr)
		}
	}

	if err := c.setEnvironmentVariables(); err != nil {
		return err
	}

	if err := c.runHookPre(cmd.Name()); err != nil {
		return err
	}

	return nil
}

// persistentStateFile returns the absolute path to the persistent state file,
// returning the first persistent file found, and returning the default path if
// none are found.
func (c *Config) persistentStateFile() (chezmoi.AbsPath, error) {
	if !c.PersistentStateAbsPath.Empty() {
		return c.PersistentStateAbsPath, nil
	}
	if !c.getConfigFileAbsPath().Empty() {
		return c.getConfigFileAbsPath().Dir().Join(persistentStateFileRelPath), nil
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
	// os/user.Current and os/user.LookupGroupId in Go's standard library are
	// generally unreliable, so work around errors if possible, or ignore them.
	//
	// On Android, user.Current always fails. Instead, use $LOGNAME (as this is
	// set by Termux), or $USER if $LOGNAME is not set.
	//
	// If CGO is disabled, then the Go standard library falls back to parsing
	// /etc/passwd and /etc/group, which will return incorrect results without
	// error if the system uses an alternative password database such as NIS or
	// LDAP.
	//
	// If CGO is enabled then os/user.Current and os/user.LookupGroupId will use
	// the underlying libc functions, namely getpwuid_r and getgrnam_r. If
	// linked with glibc this will return the correct result. If linked with
	// musl then they will use musl's implementation which, like Go's non-CGO
	// implementation, also only parses /etc/passwd and /etc/group and so also
	// returns incorrect results without error if NIS or LDAP are being used.
	//
	// On Windows, the user's group ID returned by os/user.Current() is an SID
	// and no further useful lookup is possible with Go's standard library.
	//
	// If os/user.Current fails, then fallback to $USER.
	//
	// Since neither the username nor the group are likely widely used in
	// templates, leave these variables unset if their values cannot be
	// determined. Unset variables will trigger template errors if used,
	// alerting the user to the problem and allowing them to find alternative
	// solutions.
	var gid, group, uid, username string
	if runtime.GOOS == "android" {
		username = cmp.Or(os.Getenv("LOGNAME"), os.Getenv("USER"))
	} else if currentUser, err := user.Current(); err == nil {
		gid = currentUser.Gid
		uid = currentUser.Uid
		username = currentUser.Username
		if runtime.GOOS != "windows" {
			if rawGroup, err := user.LookupGroupId(currentUser.Gid); err == nil {
				group = rawGroup.Name
			} else {
				c.logger.Info("user.LookupGroupId", slog.Any("err", err), slog.String("gid", currentUser.Gid))
			}
		}
	} else {
		c.logger.Error("user.Current", slog.Any("err", err))
		var ok bool
		username, ok = os.LookupEnv("USER")
		if !ok {
			c.logger.Info("os.LookupEnv", slog.String("key", "USER"), slog.Bool("ok", ok))
		}
	}

	fqdnHostname, err := chezmoi.FQDNHostname(c.fileSystem)
	if err != nil {
		c.logger.Info("chezmoi.FQDNHostname", slog.Any("err", err))
	}
	hostname, _, _ := strings.Cut(fqdnHostname, ".")

	kernel, err := chezmoi.Kernel(c.fileSystem)
	if err != nil {
		c.logger.Info("chezmoi.Kernel", slog.Any("err", err))
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
			c.logger.Info("chezmoi.OSRelease", slog.Any("err", err))
		}
	}

	executable, _ := os.Executable()
	windowsVersion, _ := windowsVersion()
	sourceDirAbsPath, _ := c.getSourceDirAbsPath(nil)

	return &templateData{
		arch:              runtime.GOARCH,
		args:              os.Args,
		cacheDir:          c.CacheDirAbsPath,
		command:           cmd.Name(),
		commandDir:        c.commandDirAbsPath,
		config:            c.ConfigFile.toMap(),
		configFile:        c.getConfigFileAbsPath(),
		executable:        chezmoi.NewAbsPath(executable),
		fqdnHostname:      fqdnHostname,
		gid:               gid,
		group:             group,
		homeDir:           c.homeDirAbsPath,
		hostname:          hostname,
		kernel:            kernel,
		os:                runtime.GOOS,
		osRelease:         osRelease,
		pathListSeparator: string(os.PathListSeparator),
		pathSeparator:     string(os.PathSeparator),
		sourceDir:         sourceDirAbsPath,
		uid:               uid,
		username:          username,
		version: map[string]any{
			"builtBy": c.versionInfo.BuiltBy,
			"commit":  c.versionInfo.Commit,
			"date":    c.versionInfo.Date,
			"version": c.versionInfo.Version,
		},
		windowsVersion: windowsVersion,
		workingTree:    c.WorkingTreeAbsPath,
	}
}

// readConfig reads the config file, if it exists.
func (c *Config) readConfig() error {
	switch err := c.decodeConfigFile(c.getConfigFileAbsPath(), &c.ConfigFile); {
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
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	if err := chezmoilog.LogCmdRun(c.logger, cmd); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
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

// runHookPost runs the hook's post command, if it is set.
func (c *Config) runHookPost(hook string) error {
	command := c.Hooks[hook].Post
	if command.Command == "" {
		return nil
	}
	return c.run(c.homeDirAbsPath, command.Command, command.Args)
}

// runHookPre runs the hook's pre command, if it is set.
func (c *Config) runHookPre(hook string) error {
	command := c.Hooks[hook].Pre
	if command.Command == "" {
		return nil
	}
	return c.run(c.homeDirAbsPath, command.Command, command.Args)
}

// setEncryption configures c's encryption.
func (c *Config) setEncryption() error {
	switch c.Encryption {
	case "age":
		c.Age.UseBuiltin = c.UseBuiltinAge.Value(c.useBuiltinAgeAutoFunc)
		c.encryption = &c.Age
	case "gpg":
		c.encryption = &c.GPG
	case "":
		// Detect encryption if any non-default configuration is set, preferring
		// gpg for backwards compatibility.
		switch {
		case !reflect.DeepEqual(c.GPG, defaultGPGEncryptionConfig):
			c.errorf(
				"warning: 'encryption' not set, using gpg configuration. " +
					"Check if 'encryption' is correctly set as the top-level key.\n",
			)
			c.encryption = &c.GPG
		case !reflect.DeepEqual(c.Age, defaultAgeEncryptionConfig):
			c.errorf(
				"warning: 'encryption' not set, using age configuration. " +
					"Check if 'encryption' is correctly set as the top-level key.\n",
			)
			c.Age.UseBuiltin = c.UseBuiltinAge.Value(c.useBuiltinAgeAutoFunc)
			c.encryption = &c.Age
		default:
			c.encryption = chezmoi.NoEncryption{}
		}
	default:
		return fmt.Errorf("%s: unknown encryption", c.Encryption)
	}

	if c.debug {
		encryptionLogger := c.logger.With(logComponentKey, logComponentValueEncryption)
		c.encryption = chezmoi.NewDebugEncryption(c.encryption, encryptionLogger)
	}

	return nil
}

// setEnvironmentVariables sets all environment variables defined in c.
func (c *Config) setEnvironmentVariables() error {
	var env map[string]string
	switch {
	case len(c.Env) != 0 && len(c.ScriptEnv) != 0:
		return errors.New("only one of env or scriptEnv may be set")
	case len(c.Env) != 0:
		env = c.Env
	case len(c.ScriptEnv) != 0:
		env = c.ScriptEnv
	}
	for key, value := range env {
		if strings.HasPrefix(key, "CHEZMOI_") {
			c.errorf("warning: %s: overriding reserved environment variable\n", key)
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
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
	if notInAbsDirError := (&chezmoi.NotInAbsDirError{}); errors.As(err, &notInAbsDirError) {
		return chezmoi.EmptyRelPath, fmt.Errorf("%s: not in destination directory (%s)", absPath, c.DestDirAbsPath)
	}
	return relPath, err
}

type targetRelPathsOptions struct {
	mustBeInSourceState bool
	mustNotBeExternal   bool
	recursive           bool
}

// targetRelPaths returns the target relative paths for each target path in
// args. The returned paths are sorted and de-duplicated.
func (c *Config) targetRelPaths(
	sourceState *chezmoi.SourceState,
	args []string,
	options targetRelPathsOptions,
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
		sourceStateEntry := sourceState.Get(targetRelPath)
		if sourceStateEntry == nil {
			return nil, fmt.Errorf("%s: not managed", arg)
		}
		if options.mustBeInSourceState {
			if _, ok := sourceStateEntry.(*chezmoi.SourceStateRemove); ok {
				return nil, fmt.Errorf("%s: not in source state", arg)
			}
		}
		if options.mustNotBeExternal {
			targetStateEntry, err := sourceStateEntry.TargetStateEntry(c.destSystem, c.DestDirAbsPath.Join(targetRelPath))
			if err != nil {
				return nil, err
			}
			if targetStateEntry.SourceAttr().External {
				return nil, fmt.Errorf("%s: is an external", arg)
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
func (c *Config) targetRelPathsBySourcePath(sourceState *chezmoi.SourceState, args []string) ([]chezmoi.RelPath, error) {
	targetRelPaths := make([]chezmoi.RelPath, len(args))
	targetRelPathsBySourceRelPath := make(map[chezmoi.RelPath]chezmoi.RelPath)
	_ = sourceState.ForEach(
		func(targetRelPath chezmoi.RelPath, sourceStateEntry chezmoi.SourceStateEntry) error {
			sourceRelPath := sourceStateEntry.SourceRelPath().RelPath()
			targetRelPathsBySourceRelPath[sourceRelPath] = targetRelPath
			return nil
		},
	)
	for i, arg := range args {
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
		targetRelPaths[i] = targetRelPath
	}
	return targetRelPaths, nil
}

// targetValidArgs returns target completions for toComplete given args.
func (c *Config) targetValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
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
		for i, completion := range completions {
			completions[i] = strings.TrimPrefix(completion, c.commandDirAbsPath.String()+"/")
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
	tempDir, err := os.MkdirTemp(c.TempDir.String(), key)
	chezmoilog.InfoOrError(c.logger, "MkdirTemp", err, slog.String("tempDir", tempDir))
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
	return os.WriteFile(c.outputAbsPath.String(), data, 0o666)
}

type writePathsOptions struct {
	tree bool
}

func (c *Config) writePaths(paths []string, options writePathsOptions) error {
	builder := strings.Builder{}
	if options.tree {
		newPathListTreeFromPathsSlice(paths).writeChildren(&builder, "", "  ")
	} else {
		sort.Strings(paths)
		for _, path := range paths {
			fmt.Fprintln(&builder, path)
		}
	}
	return c.writeOutputString(builder.String())
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
		Mode:         chezmoi.ModeFile,
		Pager:        os.Getenv("PAGER"),
		Progress: autoBool{
			auto: true,
		},
		PINEntry: pinEntryConfig{
			Options: pinEntryDefaultOptions,
		},
		Safe:    true,
		TempDir: chezmoi.NewAbsPath(os.TempDir()),
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
		BitwardenSecrets: bitwardenSecretsConfig{
			Command: "bws",
		},
		Dashlane: dashlaneConfig{
			Command: "dcli",
		},
		Doppler: dopplerConfig{
			Command: "doppler",
		},
		Ejson: ejsonConfig{
			KeyDir: cmp.Or(os.Getenv("EJSON_KEYDIR"), "/opt/ejson/keys"),
		},
		Gopass: gopassConfig{
			Command: "gopass",
		},
		HCPVaultSecrets: hcpVaultSecretConfig{
			Command: "vlt",
		},
		Keepassxc: keepassxcConfig{
			Command: "keepassxc-cli",
			Prompt:  true,
			Mode:    keepassxcModeCachePassword,
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
			Mode:    onepasswordModeAccount,
		},
		OnepasswordSDK: onepasswordSDKConfig{
			TokenEnvVar: "OP_SERVICE_ACCOUNT_TOKEN",
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
			Secrets:   severityWarning,
			filter:    chezmoi.NewEntryTypeFilter(chezmoi.EntryTypesAll, chezmoi.EntryTypesNone),
			recursive: true,
		},
		Diff: diffCmdConfig{
			Exclude:        chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			Pager:          defaultSentinel,
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
			PathStyle: chezmoi.PathStyleRelative.Copy(),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		Update: updateCmdConfig{
			Apply:             true,
			RecurseSubmodules: true,
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

func (f *ConfigFile) toMap() map[string]any {
	// Make a copy of f and replace any default sentinels with the empty string
	// to ensure that there are no default sentinels in the result.
	configFile := *f
	if configFile.Diff.Pager == defaultSentinel {
		configFile.Diff.Pager = ""
	}

	// This is a horrible hack. We want the returned map to contain only simple
	// types because they are used with masterminds/sprig template functions
	// which don't accept fmt.Stringers in place of strings. As a work-around,
	// round-trip via JSON.
	data, err := json.Marshal(configFile)
	if err != nil {
		return nil
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		panic(err)
	}
	return result
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
		switch fields, err := expand.Fields(&expand.Config{
			Env: expand.FuncEnviron(os.Getenv),
		}, words...); {
		case err != nil:
			return "", nil, err
		case len(fields) > 1:
			return fields[0], append(fields[1:], args...), nil
		case len(fields) == 1:
			return fields[0], args, nil
		}
	}

	// Fallback to the command only.
	return command, args, nil
}

// prependParentRelPaths returns a new slice of RelPaths where the parents of
// each RelPath appear before each RelPath.
func prependParentRelPaths(relPaths []chezmoi.RelPath) []chezmoi.RelPath {
	// For each target relative path, enumerate its parents from the root down
	// and insert any parents which have not yet been seen.
	result := make([]chezmoi.RelPath, 0, len(relPaths))
	seenRelPaths := make(map[chezmoi.RelPath]struct{}, len(relPaths))
	for _, relPath := range relPaths {
		components := relPath.SplitAll()
		for i := 1; i < len(components); i++ {
			parentRelPath := chezmoi.EmptyRelPath.Join(components[:i]...)
			if _, ok := seenRelPaths[parentRelPath]; !ok {
				result = append(result, parentRelPath)
				seenRelPaths[parentRelPath] = struct{}{}
			}
		}
		result = append(result, relPath)
		seenRelPaths[relPath] = struct{}{}
	}
	return result
}

// registerCommonFlagCompletionFuncs registers completion functions for cmd's
// common flags, recursively. It panics on any error.
func registerCommonFlagCompletionFuncs(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if _, exist := cmd.GetFlagCompletionFunc(flag.Name); exist {
			return
		}
		if flagCompletionFunc, ok := commonFlagCompletionFuncs[flag.Name]; ok {
			if err := cmd.RegisterFlagCompletionFunc(flag.Name, flagCompletionFunc); err != nil {
				panic(err)
			}
		}
	})
	for _, command := range cmd.Commands() {
		registerCommonFlagCompletionFuncs(command)
	}
}

// sha256Sum returns the SHA256 sum of data.
func sha256Sum(data []byte) []byte {
	sha256SumArr := sha256.Sum256(data)
	return sha256SumArr[:]
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
