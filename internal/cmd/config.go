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
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Masterminds/sprig/v3"
	"github.com/coreos/go-semver/semver"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/google/gops/agent"
	"github.com/mitchellh/mapstructure"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-vfs/v4"
	"github.com/twpayne/go-xdg/v6"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/v2/assets/templates"
	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/git"
)

type purgeOptions struct {
	binary bool
}

type templateConfig struct {
	Options []string `mapstructure:"options"`
}

// A Config represents a configuration.
type Config struct {
	// Global configuration, settable in the config file.
	CacheDirAbsPath    chezmoi.AbsPath                 `mapstructure:"cacheDir"`
	Color              autoBool                        `mapstructure:"color"`
	Data               map[string]interface{}          `mapstructure:"data"`
	DestDirAbsPath     chezmoi.AbsPath                 `mapstructure:"destDir"`
	Interpreters       map[string]*chezmoi.Interpreter `mapstructure:"interpreters"`
	Mode               chezmoi.Mode                    `mapstructure:"mode"`
	Pager              string                          `mapstructure:"pager"`
	Safe               bool                            `mapstructure:"safe"`
	SourceDirAbsPath   chezmoi.AbsPath                 `mapstructure:"sourceDir"`
	Template           templateConfig                  `mapstructure:"template"`
	Umask              fs.FileMode                     `mapstructure:"umask"`
	UseBuiltinAge      autoBool                        `mapstructure:"useBuiltinAge"`
	UseBuiltinGit      autoBool                        `mapstructure:"useBuiltinGit"`
	WorkingTreeAbsPath chezmoi.AbsPath                 `mapstructure:"workingTree"`

	// Global configuration, not settable in the config file.
	configFormat     readDataFormat
	cpuProfile       chezmoi.AbsPath
	debug            bool
	dryRun           bool
	force            bool
	gops             bool
	homeDir          string
	keepGoing        bool
	noPager          bool
	noTTY            bool
	outputAbsPath    chezmoi.AbsPath
	refreshExternals bool
	sourcePath       bool
	verbose          bool
	templateFuncs    template.FuncMap

	// Password manager configurations, settable in the config file.
	Bitwarden   bitwardenConfig   `mapstructure:"bitwarden"`
	Gopass      gopassConfig      `mapstructure:"gopass"`
	Keepassxc   keepassxcConfig   `mapstructure:"keepassxc"`
	Lastpass    lastpassConfig    `mapstructure:"lastpass"`
	Onepassword onepasswordConfig `mapstructure:"onepassword"`
	Pass        passConfig        `mapstructure:"pass"`
	Secret      secretConfig      `mapstructure:"secret"`
	Vault       vaultConfig       `mapstructure:"vault"`

	// Encryption configurations, settable in the config file.
	Encryption string                `mapstructure:"encryption"`
	Age        chezmoi.AgeEncryption `mapstructure:"age"`
	GPG        chezmoi.GPGEncryption `mapstructure:"gpg"`

	// Password manager data.
	gitHub  gitHubData
	keyring keyringData

	// Command configurations, settable in the config file.
	Add   addCmdConfig   `mapstructure:"add"`
	CD    cdCmdConfig    `mapstructure:"cd"`
	Diff  diffCmdConfig  `mapstructure:"diff"`
	Docs  docsCmdConfig  `mapstructure:"docs"`
	Edit  editCmdConfig  `mapstructure:"edit"`
	Git   gitCmdConfig   `mapstructure:"git"`
	Merge mergeCmdConfig `mapstructure:"merge"`

	// Command configurations, not settable in the config file.
	apply           applyCmdConfig
	archive         archiveCmdConfig
	data            dataCmdConfig
	dump            dumpCmdConfig
	executeTemplate executeTemplateCmdConfig
	_import         importCmdConfig
	init            initCmdConfig
	managed         managedCmdConfig
	purge           purgeCmdConfig
	reAdd           reAddCmdConfig
	remove          removeCmdConfig
	secretKeyring   secretKeyringCmdConfig
	state           stateCmdConfig
	status          statusCmdConfig
	update          updateCmdConfig
	upgrade         upgradeCmdConfig
	verify          verifyCmdConfig

	// Version information.
	version     *semver.Version
	versionInfo VersionInfo
	versionStr  string

	// Configuration.
	fileSystem        vfs.FS
	bds               *xdg.BaseDirectorySpecification
	configFileAbsPath chezmoi.AbsPath
	baseSystem        chezmoi.System
	sourceSystem      chezmoi.System
	destSystem        chezmoi.System
	persistentState   chezmoi.PersistentState

	// Computed configuration.
	homeDirAbsPath chezmoi.AbsPath
	encryption     chezmoi.Encryption

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	ioregData ioregData
}

// A configOption sets and option on a Config.
type configOption func(*Config) error

type configState struct {
	ConfigTemplateContentsSHA256 chezmoi.HexBytes `json:"configTemplateContentsSHA256" yaml:"configTemplateContentsSHA256"` //nolint:tagliatelle
}

var (
	persistentStateFilename = chezmoi.RelPath("chezmoistate.boltdb")
	configStateKey          = []byte("configState")

	defaultAgeEncryptionConfig = chezmoi.AgeEncryption{
		Command: "age",
		Suffix:  ".age",
	}
	defaultGPGEncryptionConfig = chezmoi.GPGEncryption{
		Command: "gpg",
		Suffix:  ".asc",
	}

	identifierRx = regexp.MustCompile(`\A[\pL_][\pL\p{Nd}_]*\z`)
	whitespaceRx = regexp.MustCompile(`\s+`)

	viperDecodeConfigOptions = []viper.DecoderConfigOption{
		viper.DecodeHook(
			mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
				chezmoi.StringSliceToEntryTypeSetHookFunc(),
				chezmoi.StringToAbsPathHookFunc(),
				StringOrBoolToAutoBoolHookFunc(),
			),
		),
	}
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
		// Global configuration, settable in the config file.
		CacheDirAbsPath: chezmoi.NewAbsPath(bds.CacheHome).Join("chezmoi"),
		Color: autoBool{
			auto: true,
		},
		Interpreters: defaultInterpreters,
		Pager:        os.Getenv("PAGER"),
		Safe:         true,
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

		// Global configuration, not settable in the config file.
		homeDir:       userHomeDir,
		templateFuncs: sprig.TxtFuncMap(),

		// Password manager configurations, settable in the config file.
		Bitwarden: bitwardenConfig{
			Command: "bw",
		},
		Gopass: gopassConfig{
			Command: "gopass",
		},
		Keepassxc: keepassxcConfig{
			Command: "keepassxc-cli",
		},
		Lastpass: lastpassConfig{
			Command: "lpass",
		},
		Onepassword: onepasswordConfig{
			Command: "op",
		},
		Pass: passConfig{
			Command: "pass",
		},
		Vault: vaultConfig{
			Command: "vault",
		},

		// Encryption configurations, settable in the config file.
		Age: defaultAgeEncryptionConfig,
		GPG: defaultGPGEncryptionConfig,

		// Password manager data.

		// Command configurations, settable in the config file.
		Add: addCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		Diff: diffCmdConfig{
			Exclude: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
		},
		Docs: docsCmdConfig{
			MaxWidth: 80,
		},
		Edit: editCmdConfig{
			exclude: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypeDirs | chezmoi.EntryTypeFiles | chezmoi.EntryTypeSymlinks | chezmoi.EntryTypeEncrypted),
		},
		Git: gitCmdConfig{
			Command: "git",
		},
		Merge: mergeCmdConfig{
			Command: "vimdiff",
		},

		// Command configurations, not settable in the config file.
		apply: applyCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		archive: archiveCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		data: dataCmdConfig{
			format: defaultWriteDataFormat,
		},
		dump: dumpCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			format:    defaultWriteDataFormat,
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		executeTemplate: executeTemplateCmdConfig{
			stdinIsATTY: true,
		},
		_import: importCmdConfig{
			destination: homeDirAbsPath,
			exclude:     chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:     chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
		},
		init: initCmdConfig{
			data:    true,
			exclude: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
		},
		managed: managedCmdConfig{
			exclude: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypeDirs | chezmoi.EntryTypeFiles | chezmoi.EntryTypeSymlinks | chezmoi.EntryTypeEncrypted),
		},
		reAdd: reAddCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		state: stateCmdConfig{
			data: stateDataCmdConfig{
				format: defaultWriteDataFormat,
			},
			dump: stateDumpCmdConfig{
				format: defaultWriteDataFormat,
			},
		},
		status: statusCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		update: updateCmdConfig{
			apply:     true,
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		verify: verifyCmdConfig{
			exclude:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll &^ chezmoi.EntryTypeScripts),
			recursive: true,
		},

		// Configuration.
		fileSystem: vfs.OSFS,
		bds:        bds,

		// Computed configuration.
		homeDirAbsPath: homeDirAbsPath,

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	for key, value := range map[string]interface{}{
		"bitwarden":                c.bitwardenTemplateFunc,
		"bitwardenAttachment":      c.bitwardenAttachmentTemplateFunc,
		"bitwardenFields":          c.bitwardenFieldsTemplateFunc,
		"gitHubKeys":               c.gitHubKeysTemplateFunc,
		"gopass":                   c.gopassTemplateFunc,
		"gopassRaw":                c.gopassRawTemplateFunc,
		"include":                  c.includeTemplateFunc,
		"ioreg":                    c.ioregTemplateFunc,
		"joinPath":                 c.joinPathTemplateFunc,
		"keepassxc":                c.keepassxcTemplateFunc,
		"keepassxcAttribute":       c.keepassxcAttributeTemplateFunc,
		"keyring":                  c.keyringTemplateFunc,
		"lastpass":                 c.lastpassTemplateFunc,
		"lastpassRaw":              c.lastpassRawTemplateFunc,
		"lookPath":                 c.lookPathTemplateFunc,
		"mozillaInstallHash":       c.mozillaInstallHashTemplateFunc,
		"onepassword":              c.onepasswordTemplateFunc,
		"onepasswordDetailsFields": c.onepasswordDetailsFieldsTemplateFunc,
		"onepasswordDocument":      c.onepasswordDocumentTemplateFunc,
		"onepasswordItemFields":    c.onepasswordItemFieldsTemplateFunc,
		"output":                   c.outputTemplateFunc,
		"pass":                     c.passTemplateFunc,
		"passRaw":                  c.passRawTemplateFunc,
		"secret":                   c.secretTemplateFunc,
		"secretJSON":               c.secretJSONTemplateFunc,
		"stat":                     c.statTemplateFunc,
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
	c.configFileAbsPath, err = c.defaultConfigFile(c.fileSystem, c.bds)
	if err != nil {
		return nil, err
	}
	c.SourceDirAbsPath, err = c.defaultSourceDir(c.fileSystem, c.bds)
	if err != nil {
		return nil, err
	}
	c.DestDirAbsPath = c.homeDirAbsPath
	c._import.destination = c.homeDirAbsPath

	return c, nil
}

func (c *Config) addTemplateFunc(key string, value interface{}) {
	if _, ok := c.templateFuncs[key]; ok {
		panic(fmt.Sprintf("%s: already defined", key))
	}
	c.templateFuncs[key] = value
}

type applyArgsOptions struct {
	include      *chezmoi.EntryTypeSet
	exclude      *chezmoi.EntryTypeSet
	recursive    bool
	umask        fs.FileMode
	preApplyFunc chezmoi.PreApplyFunc
}

func (c *Config) applyArgs(ctx context.Context, targetSystem chezmoi.System, targetDirAbsPath chezmoi.AbsPath, args []string, options applyArgsOptions) error {
	sourceState, err := c.newSourceState(ctx)
	if err != nil {
		return err
	}

	var currentConfigTemplateContentsSHA256 []byte
	configTemplateRelPath, _, configTemplateContents, err := c.findConfigTemplate()
	if err != nil {
		return err
	}
	if configTemplateRelPath != "" {
		currentConfigTemplateContentsSHA256 = chezmoi.SHA256Sum(configTemplateContents)
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
	configTemplateContentsUnchanged := (currentConfigTemplateContentsSHA256 == nil && previousConfigTemplateContentsSHA256 == nil) ||
		bytes.Equal(currentConfigTemplateContentsSHA256, previousConfigTemplateContentsSHA256)
	if !configTemplateContentsUnchanged {
		if c.force {
			if configTemplateRelPath == "" {
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
		} else {
			c.errorf("warning: config file template has changed, run chezmoi init to regenerate config file\n")
		}
	}

	applyOptions := chezmoi.ApplyOptions{
		Include:      options.include.Sub(options.exclude),
		PreApplyFunc: options.preApplyFunc,
		Umask:        options.umask,
	}

	var targetRelPaths []chezmoi.RelPath
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

	//nolint:ifshort
	keptGoingAfterErr := false
	for _, targetRelPath := range targetRelPaths {
		switch err := sourceState.Apply(targetSystem, c.destSystem, c.persistentState, targetDirAbsPath, targetRelPath, applyOptions); {
		case errors.Is(err, chezmoi.Skip):
			continue
		case err != nil && c.keepGoing:
			c.errorf("%v\n", err)
			keptGoingAfterErr = true
		case err != nil:
			return err
		}
	}
	if keptGoingAfterErr {
		return ExitCodeError(1)
	}

	return nil
}

func (c *Config) cmdOutput(dirAbsPath chezmoi.AbsPath, name string, args []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if !dirAbsPath.Empty() {
		dirRawAbsPath, err := c.baseSystem.RawPath(dirAbsPath)
		if err != nil {
			return nil, err
		}
		cmd.Dir = dirRawAbsPath.String()
	}
	return c.baseSystem.IdempotentCmdOutput(cmd)
}

func (c *Config) colorAutoFunc() bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	if stdout, ok := c.stdout.(*os.File); ok {
		return term.IsTerminal(int(stdout.Fd()))
	}
	return false
}

// defaultConfigFile returns the default config file according to the XDG Base
// Directory Specification.
func (c *Config) defaultConfigFile(fileSystem vfs.Stater, bds *xdg.BaseDirectorySpecification) (chezmoi.AbsPath, error) {
	// Search XDG Base Directory Specification config directories first.
	for _, configDir := range bds.ConfigDirs {
		configDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(configDir, c.homeDirAbsPath)
		if err != nil {
			return chezmoi.EmptyAbsPath, err
		}
		for _, extension := range viper.SupportedExts {
			configFileAbsPath := configDirAbsPath.Join("chezmoi", chezmoi.RelPath("chezmoi."+extension))
			if _, err := fileSystem.Stat(configFileAbsPath.String()); err == nil {
				return configFileAbsPath, nil
			}
		}
	}
	// Fallback to XDG Base Directory Specification default.
	configHomeAbsPath, err := chezmoi.NewAbsPathFromExtPath(bds.ConfigHome, c.homeDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return configHomeAbsPath.Join("chezmoi", "chezmoi.toml"), nil
}

func (c *Config) defaultPreApplyFunc(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
	log.Debug().
		Stringer("targetRelPath", targetRelPath).
		Object("targetEntryState", targetEntryState).
		Object("lastWrittenEntryState", lastWrittenEntryState).
		Object("actualEntryState", actualEntryState).
		Msg("defaultPreApplyFunc")

	switch {
	case targetEntryState.Overwrite():
		return nil
	case targetEntryState.Type == chezmoi.EntryStateTypeScript:
		return nil
	case c.force:
		return nil
	case lastWrittenEntryState == nil:
		return nil
	case lastWrittenEntryState.Equivalent(actualEntryState):
		return nil
	}

	// LATER add merge option
	var choices []string
	actualContents := actualEntryState.Contents()
	targetContents := targetEntryState.Contents()
	if actualContents != nil || targetContents != nil {
		choices = append(choices, "diff")
	}
	choices = append(choices, "overwrite", "all-overwrite", "skip", "quit")
	for {
		switch choice, err := c.promptChoice(fmt.Sprintf("%s has changed since chezmoi last wrote it", targetRelPath), choices); {
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
			return chezmoi.Skip
		case choice == "quit":
			return ExitCodeError(1)
		default:
			return nil
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
		sourceDirAbsPath := dataDirAbsPath.Join("chezmoi")
		if _, err := fileSystem.Stat(sourceDirAbsPath.String()); err == nil {
			return sourceDirAbsPath, nil
		}
	}
	// Fallback to XDG Base Directory Specification default.
	dataHomeAbsPath, err := chezmoi.NewAbsPathFromExtPath(bds.DataHome, c.homeDirAbsPath)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return dataHomeAbsPath.Join("chezmoi"), nil
}

func (c *Config) defaultTemplateData() map[string]interface{} {
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
	var username, group string
	if currentUser, err := user.Current(); err == nil {
		username = currentUser.Username
		if runtime.GOOS != "windows" {
			if rawGroup, err := user.LookupGroupId(currentUser.Gid); err == nil {
				group = rawGroup.Name
			} else {
				log.Debug().
					Str("gid", currentUser.Gid).
					Err(err).
					Msg("user.LookupGroupId")
			}
		}
	} else {
		log.Debug().
			Err(err).
			Msg("user.Current")
		var ok bool
		username, ok = os.LookupEnv("USER")
		if !ok {
			log.Debug().
				Str("key", "USER").
				Bool("ok", ok).
				Msg("os.LookupEnv")
		}
	}

	fqdnHostname := chezmoi.FQDNHostname(c.fileSystem)

	var hostname string
	if rawHostname, err := os.Hostname(); err == nil {
		hostname = strings.SplitN(rawHostname, ".", 2)[0]
	} else {
		log.Debug().
			Err(err).
			Msg("os.Hostname")
	}

	kernel, err := chezmoi.Kernel(c.fileSystem)
	if err != nil {
		log.Debug().
			Err(err).
			Msg("chezmoi.Kernel")
	}

	var osRelease map[string]interface{}
	if rawOSRelease, err := chezmoi.OSRelease(c.baseSystem); err == nil {
		osRelease = upperSnakeCaseToCamelCaseMap(rawOSRelease)
	} else {
		log.Debug().
			Err(err).
			Msg("chezmoi.OSRelease")
	}

	return map[string]interface{}{
		"chezmoi": map[string]interface{}{
			"arch":         runtime.GOARCH,
			"fqdnHostname": fqdnHostname,
			"group":        group,
			"homeDir":      c.homeDir,
			"hostname":     hostname,
			"kernel":       kernel,
			"os":           runtime.GOOS,
			"osRelease":    osRelease,
			"sourceDir":    c.SourceDirAbsPath.String(),
			"username":     username,
			"version": map[string]interface{}{
				"builtBy": c.versionInfo.BuiltBy,
				"commit":  c.versionInfo.Commit,
				"date":    c.versionInfo.Date,
				"version": c.versionInfo.Version,
			},
		},
	}
}

type destAbsPathInfosOptions struct {
	follow         bool
	ignoreNotExist bool
	recursive      bool
}

func (c *Config) destAbsPathInfos(sourceState *chezmoi.SourceState, args []string, options destAbsPathInfosOptions) (map[chezmoi.AbsPath]fs.FileInfo, error) {
	destAbsPathInfos := make(map[chezmoi.AbsPath]fs.FileInfo)
	for _, arg := range args {
		arg = filepath.Clean(arg)
		destAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		if _, err := destAbsPath.TrimDirPrefix(c.DestDirAbsPath); err != nil {
			return nil, err
		}
		if options.recursive {
			if err := chezmoi.Walk(c.destSystem, destAbsPath, func(destAbsPath chezmoi.AbsPath, info fs.FileInfo, err error) error {
				switch {
				case options.ignoreNotExist && errors.Is(err, fs.ErrNotExist):
					return nil
				case err != nil:
					return err
				}
				if options.follow && info.Mode().Type() == fs.ModeSymlink {
					info, err = c.destSystem.Stat(destAbsPath)
					if err != nil {
						return err
					}
				}
				return sourceState.AddDestAbsPathInfos(destAbsPathInfos, c.destSystem, destAbsPath, info)
			}); err != nil {
				return nil, err
			}
		} else {
			var info fs.FileInfo
			if options.follow {
				info, err = c.destSystem.Stat(destAbsPath)
			} else {
				info, err = c.destSystem.Lstat(destAbsPath)
			}
			switch {
			case options.ignoreNotExist && errors.Is(err, fs.ErrNotExist):
				continue
			case err != nil:
				return nil, err
			}
			if err := sourceState.AddDestAbsPathInfos(destAbsPathInfos, c.destSystem, destAbsPath, info); err != nil {
				return nil, err
			}
		}
	}
	return destAbsPathInfos, nil
}

func (c *Config) diffFile(path chezmoi.RelPath, fromData []byte, fromMode fs.FileMode, toData []byte, toMode fs.FileMode) error {
	var sb strings.Builder
	unifiedEncoder := diff.NewUnifiedEncoder(&sb, diff.DefaultContextLines)
	color := c.Color.Value(c.colorAutoFunc)
	if color {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	diffPatch, err := chezmoi.DiffPatch(path, fromData, fromMode, toData, toMode)
	if err != nil {
		return err
	}
	if err := unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return c.pageOutputString(sb.String(), c.Diff.Pager)
}

func (c *Config) doPurge(purgeOptions *purgeOptions) error {
	if c.persistentState != nil {
		if err := c.persistentState.Close(); err != nil {
			return err
		}
	}

	persistentStateFileAbsPath, err := c.persistentStateFile()
	if err != nil {
		return err
	}
	absPaths := []chezmoi.AbsPath{
		c.CacheDirAbsPath,
		c.configFileAbsPath.Dir(),
		c.configFileAbsPath,
		persistentStateFileAbsPath,
		c.WorkingTreeAbsPath,
		c.SourceDirAbsPath,
	}
	if purgeOptions != nil && purgeOptions.binary {
		executable, err := os.Executable()
		// Special case: do not purge the binary if it is a test binary created
		// by go test as this would break later tests.
		if err == nil && !strings.Contains(executable, "test") {
			absPaths = append(absPaths, chezmoi.NewAbsPath(executable))
		}
	}

	// Remove all paths that exist.
	for _, absPath := range absPaths {
		switch _, err := c.destSystem.Stat(absPath); {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return err
		}

		if !c.force {
			switch choice, err := c.promptChoice(fmt.Sprintf("Remove %s", absPath), yesNoAllQuit); {
			case err != nil:
				return err
			case choice == "yes":
			case choice == "no":
				continue
			case choice == "all":
				c.force = true
			case choice == "quit":
				return nil
			}
		}

		switch err := c.destSystem.RemoveAll(absPath); {
		case errors.Is(err, fs.ErrPermission):
			continue
		case err != nil:
			return err
		}
	}

	return nil
}

// editor returns the path to the user's editor and any extra arguments.
func (c *Config) editor() (string, []string) {
	// If the user has set and edit command then use it.
	if c.Edit.Command != "" {
		return c.Edit.Command, c.Edit.Args
	}

	// Prefer $VISUAL over $EDITOR and fallback to the OS's default editor.
	editor := firstNonEmptyString(
		os.Getenv("VISUAL"),
		os.Getenv("EDITOR"),
		defaultEditor,
	)

	// If editor is found, return it.
	if path, err := exec.LookPath(editor); err == nil {
		return path, nil
	}

	// Otherwise, if editor contains spaces, then assume that the first word is
	// the editor and the rest are arguments.
	components := whitespaceRx.Split(editor, -1)
	if len(components) > 1 {
		if path, err := exec.LookPath(components[0]); err == nil {
			return path, components[1:]
		}
	}

	// Fallback to editor only.
	return editor, nil
}

func (c *Config) errorf(format string, args ...interface{}) {
	fmt.Fprintf(c.stderr, "chezmoi: "+format, args...)
}

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

func (c *Config) findConfigTemplate() (chezmoi.RelPath, string, []byte, error) {
	for _, ext := range viper.SupportedExts {
		filename := chezmoi.RelPath(chezmoi.Prefix + "." + ext + chezmoi.TemplateSuffix)
		contents, err := c.baseSystem.ReadFile(c.SourceDirAbsPath.Join(filename))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			continue
		case err != nil:
			return "", "", nil, err
		}
		return chezmoi.RelPath("chezmoi." + ext), ext, contents, nil
	}
	return "", "", nil, nil
}

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

func (c *Config) gitAutoCommit(status *git.Status) error {
	if status.Empty() {
		return nil
	}
	commitMessageTemplate, err := templates.FS.ReadFile("COMMIT_MESSAGE.tmpl")
	if err != nil {
		return err
	}
	commitMessageTmpl, err := template.New("commit_message").Funcs(c.templateFuncs).Parse(string(commitMessageTemplate))
	if err != nil {
		return err
	}
	commitMessage := strings.Builder{}
	if err := commitMessageTmpl.Execute(&commitMessage, status); err != nil {
		return err
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"commit", "--message", commitMessage.String()})
}

func (c *Config) gitAutoPush(status *git.Status) error {
	if status.Empty() {
		return nil
	}
	return c.run(c.WorkingTreeAbsPath, c.Git.Command, []string{"push"})
}

func (c *Config) makeRunEWithSourceState(runE func(*cobra.Command, []string, *chezmoi.SourceState) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sourceState, err := c.newSourceState(cmd.Context())
		if err != nil {
			return err
		}
		return runE(cmd, args, sourceState)
	}
}

func (c *Config) marshal(dataFormat writeDataFormat, data interface{}) error {
	var format chezmoi.Format
	switch dataFormat {
	case writeDataFormatJSON:
		format = chezmoi.FormatJSON
	case writeDataFormatYAML:
		format = chezmoi.FormatYAML
	default:
		return fmt.Errorf("%s: unknown format", dataFormat)
	}
	marshaledData, err := format.Marshal(data)
	if err != nil {
		return err
	}
	return c.writeOutput(marshaledData)
}

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

	persistentFlags.Var(&c.Color, "color", "Colorize output")
	persistentFlags.VarP(&c.DestDirAbsPath, "destination", "D", "Set destination directory")
	persistentFlags.Var(&c.Mode, "mode", "Mode")
	persistentFlags.BoolVar(&c.Safe, "safe", c.Safe, "Safely replace files and symlinks")
	persistentFlags.VarP(&c.SourceDirAbsPath, "source", "S", "Set source directory")
	persistentFlags.Var(&c.UseBuiltinAge, "use-builtin-age", "Use builtin age")
	persistentFlags.Var(&c.UseBuiltinGit, "use-builtin-git", "Use builtin git")
	persistentFlags.VarP(&c.WorkingTreeAbsPath, "working-tree", "W", "Set working tree directory")
	for viperKey, key := range map[string]string{
		"color":         "color",
		"destDir":       "destination",
		"mode":          "mode",
		"safe":          "safe",
		"sourceDir":     "source",
		"useBuiltinAge": "use-builtin-age",
		"useBuiltinGit": "use-builtin-git",
		"workingTree":   "working-tree",
	} {
		if err := viper.BindPFlag(viperKey, persistentFlags.Lookup(key)); err != nil {
			return nil, err
		}
	}

	persistentFlags.VarP(&c.configFileAbsPath, "config", "c", "Set config file")
	persistentFlags.Var(&c.configFormat, "config-format", "Set config file format")
	persistentFlags.Var(&c.cpuProfile, "cpu-profile", "Write a CPU profile to path")
	persistentFlags.BoolVar(&c.debug, "debug", c.debug, "Include debug information in output")
	persistentFlags.BoolVarP(&c.dryRun, "dry-run", "n", c.dryRun, "Do not make any modifications to the destination directory")
	persistentFlags.BoolVar(&c.force, "force", c.force, "Make all changes without prompting")
	persistentFlags.BoolVar(&c.gops, "gops", c.gops, "Enable gops agent")
	persistentFlags.BoolVarP(&c.keepGoing, "keep-going", "k", c.keepGoing, "Keep going as far as possible after an error")
	persistentFlags.BoolVar(&c.noPager, "no-pager", c.noPager, "Do not use the pager")
	persistentFlags.BoolVar(&c.noTTY, "no-tty", c.noTTY, "Do not attempt to get a TTY for reading passwords")
	persistentFlags.VarP(&c.outputAbsPath, "output", "o", "Write output to path instead of stdout")
	persistentFlags.BoolVarP(&c.refreshExternals, "refresh-externals", "R", c.refreshExternals, "Refresh external cache")
	persistentFlags.BoolVar(&c.sourcePath, "source-path", c.sourcePath, "Specify targets by source path")
	persistentFlags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "Make output more verbose")

	for _, err := range []error{
		rootCmd.MarkPersistentFlagFilename("config"),
		rootCmd.MarkPersistentFlagFilename("cpu-profile"),
		persistentFlags.MarkHidden("cpu-profile"),
		rootCmd.MarkPersistentFlagDirname("destination"),
		persistentFlags.MarkHidden("gops"),
		rootCmd.MarkPersistentFlagFilename("output"),
		persistentFlags.MarkHidden("safe"),
		rootCmd.MarkPersistentFlagDirname("source"),
	} {
		if err != nil {
			return nil, err
		}
	}

	rootCmd.SetHelpCommand(c.newHelpCmd())
	for _, cmd := range []*cobra.Command{
		c.newAddCmd(),
		c.newApplyCmd(),
		c.newArchiveCmd(),
		c.newCatCmd(),
		c.newCDCmd(),
		c.newChattrCmd(),
		c.newCompletionCmd(),
		c.newDataCmd(),
		c.newDecryptCommand(),
		c.newDiffCmd(),
		c.newDocsCmd(),
		c.newDoctorCmd(),
		c.newDumpCmd(),
		c.newEditCmd(),
		c.newEditConfigCmd(),
		c.newEncryptCommand(),
		c.newExecuteTemplateCmd(),
		c.newForgetCmd(),
		c.newGitCmd(),
		c.newImportCmd(),
		c.newInitCmd(),
		c.newInternalTestCmd(),
		c.newManagedCmd(),
		c.newMergeCmd(),
		c.newPurgeCmd(),
		c.newReAddCmd(),
		c.newRemoveCmd(),
		c.newSecretCmd(),
		c.newSourcePathCmd(),
		c.newStateCmd(),
		c.newStatusCmd(),
		c.newUnmanagedCmd(),
		c.newUpdateCmd(),
		c.newUpgradeCmd(),
		c.newVerifyCmd(),
	} {
		if cmd != nil {
			rootCmd.AddCommand(cmd)
		}
	}

	return rootCmd, nil
}

func (c *Config) newSourceState(ctx context.Context, options ...chezmoi.SourceStateOption) (*chezmoi.SourceState, error) {
	s := chezmoi.NewSourceState(append([]chezmoi.SourceStateOption{
		chezmoi.WithBaseSystem(c.baseSystem),
		chezmoi.WithCacheDir(c.CacheDirAbsPath),
		chezmoi.WithDefaultTemplateDataFunc(c.defaultTemplateData),
		chezmoi.WithDestDir(c.DestDirAbsPath),
		chezmoi.WithEncryption(c.encryption),
		chezmoi.WithInterpreters(c.Interpreters),
		chezmoi.WithMode(c.Mode),
		chezmoi.WithPriorityTemplateData(c.Data),
		chezmoi.WithSourceDir(c.SourceDirAbsPath),
		chezmoi.WithSystem(c.sourceSystem),
		chezmoi.WithTemplateFuncs(c.templateFuncs),
		chezmoi.WithTemplateOptions(c.Template.Options),
	}, options...)...)

	if err := s.Read(ctx, &chezmoi.ReadOptions{
		RefreshExternals: c.refreshExternals,
	}); err != nil {
		return nil, err
	}

	if minVersion := s.MinVersion(); c.version != nil && !isDevVersion(c.version) && c.version.LessThan(minVersion) {
		return nil, fmt.Errorf("source state requires version %s or later, chezmoi is version %s", minVersion, c.version)
	}

	return s, nil
}

func (c *Config) persistentPostRunRootE(cmd *cobra.Command, args []string) error {
	defer pprof.StopCPUProfile()

	if err := c.persistentState.Close(); err != nil {
		return err
	}

	if boolAnnotation(cmd, modifiesConfigFile) {
		// Warn the user of any errors reading the config file.
		v := viper.New()
		v.SetFs(afero.FromIOFS{FS: c.fileSystem})
		v.SetConfigFile(c.configFileAbsPath.String())
		err := v.ReadInConfig()
		if err == nil {
			err = v.Unmarshal(&Config{}, viperDecodeConfigOptions...)
		}
		if err != nil {
			c.errorf("warning: %s: %v\n", c.configFileAbsPath, err)
		}
	}

	if boolAnnotation(cmd, modifiesSourceDirectory) {
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

	if c.gops {
		agent.Close()
	}

	return nil
}

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
		shell, _ := shell.CurrentUserShell()
		pagerCmd = exec.Command(shell, "-c", pager)
	} else {
		pagerCmd = exec.Command(pager)
	}
	pagerCmd.Stdin = bytes.NewBufferString(output)
	pagerCmd.Stdout = c.stdout
	pagerCmd.Stderr = c.stderr
	return pagerCmd.Run()
}

func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	if !c.cpuProfile.Empty() {
		f, err := os.Create(c.cpuProfile.String())
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
	}

	if c.gops {
		if err := agent.Listen(agent.Options{}); err != nil {
			return err
		}
	}

	if err := c.readConfig(); err != nil {
		if !boolAnnotation(cmd, doesNotRequireValidConfig) {
			return fmt.Errorf("invalid config: %s: %w", c.configFileAbsPath, err)
		}
		c.errorf("warning: %s: %v\n", c.configFileAbsPath, err)
	}

	color := c.Color.Value(c.colorAutoFunc)
	if color {
		if err := enableVirtualTerminalProcessing(c.stdout); err != nil {
			return err
		}
	}

	log.Logger = log.Output(zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.Out = c.stderr
			w.NoColor = !color
			w.TimeFormat = time.RFC3339
		},
	))
	if c.debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	log.Debug().
		Object("version", c.versionInfo).
		Strs("args", args).
		Str("goVersion", runtime.Version()).
		Msg("persistentPreRunRootE")

	c.baseSystem = chezmoi.NewRealSystem(c.fileSystem,
		chezmoi.RealSystemWithSafe(c.Safe),
	)
	if c.debug {
		c.baseSystem = chezmoi.NewDebugSystem(c.baseSystem)
	}

	switch {
	case cmd.Annotations[persistentStateMode] == persistentStateModeEmpty:
		c.persistentState = chezmoi.NewMockPersistentState()
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadOnly:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		c.persistentState, err = chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadOnly)
		if err != nil {
			return err
		}
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadMockWrite:
		fallthrough
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadWrite && c.dryRun:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		persistentState, err := chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadOnly)
		if err != nil {
			return err
		}
		dryRunPeristentState := chezmoi.NewMockPersistentState()
		if err := persistentState.CopyTo(dryRunPeristentState); err != nil {
			return err
		}
		if err := persistentState.Close(); err != nil {
			return err
		}
		c.persistentState = dryRunPeristentState
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadWrite:
		persistentStateFileAbsPath, err := c.persistentStateFile()
		if err != nil {
			return err
		}
		c.persistentState, err = chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFileAbsPath, chezmoi.BoltPersistentStateReadWrite)
		if err != nil {
			return err
		}
	default:
		c.persistentState = chezmoi.NullPersistentState{}
	}
	if c.debug && c.persistentState != nil {
		c.persistentState = chezmoi.NewDebugPersistentState(c.persistentState)
	}

	c.sourceSystem = c.baseSystem
	c.destSystem = c.baseSystem
	if !boolAnnotation(cmd, modifiesDestinationDirectory) {
		c.destSystem = chezmoi.NewReadOnlySystem(c.destSystem)
	}
	if !boolAnnotation(cmd, modifiesSourceDirectory) {
		c.sourceSystem = chezmoi.NewReadOnlySystem(c.sourceSystem)
	}
	if c.dryRun {
		c.sourceSystem = chezmoi.NewDryRunSystem(c.sourceSystem)
		c.destSystem = chezmoi.NewDryRunSystem(c.destSystem)
	}
	if c.verbose {
		c.sourceSystem = chezmoi.NewGitDiffSystem(c.sourceSystem, c.stdout, c.SourceDirAbsPath, color)
		c.destSystem = chezmoi.NewGitDiffSystem(c.destSystem, c.stdout, c.DestDirAbsPath, color)
	}

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
		c.encryption = chezmoi.NewDebugEncryption(c.encryption)
	}

	if boolAnnotation(cmd, requiresConfigDirectory) {
		if err := chezmoi.MkdirAll(c.baseSystem, c.configFileAbsPath.Dir(), 0o777); err != nil {
			return err
		}
	}

	if boolAnnotation(cmd, requiresSourceDirectory) {
		if err := chezmoi.MkdirAll(c.baseSystem, c.SourceDirAbsPath, 0o777); err != nil {
			return err
		}
	}

	if boolAnnotation(cmd, runsCommands) {
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

	if c.WorkingTreeAbsPath.Empty() {
		workingTreeAbsPath := c.SourceDirAbsPath
	FOR:
		for {
			if info, err := c.baseSystem.Stat(workingTreeAbsPath.Join(".git")); err == nil && info.IsDir() {
				c.WorkingTreeAbsPath = workingTreeAbsPath
				break FOR
			}
			prevWorkingTreeDirAbsPath := workingTreeAbsPath
			workingTreeAbsPath = workingTreeAbsPath.Dir()
			if workingTreeAbsPath.Len() >= prevWorkingTreeDirAbsPath.Len() {
				c.WorkingTreeAbsPath = c.SourceDirAbsPath
				break FOR
			}
		}
	}

	if boolAnnotation(cmd, requiresWorkingTree) {
		if _, err := c.SourceDirAbsPath.TrimDirPrefix(c.WorkingTreeAbsPath); err != nil {
			return err
		}
		if err := chezmoi.MkdirAll(c.baseSystem, c.WorkingTreeAbsPath, 0o777); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) persistentStateFile() (chezmoi.AbsPath, error) {
	if !c.configFileAbsPath.Empty() {
		return c.configFileAbsPath.Dir().Join(persistentStateFilename), nil
	}
	for _, configDir := range c.bds.ConfigDirs {
		configDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(configDir, c.homeDirAbsPath)
		if err != nil {
			return chezmoi.EmptyAbsPath, err
		}
		persistentStateFile := configDirAbsPath.Join("chezmoi", persistentStateFilename)
		if _, err := os.Stat(persistentStateFile.String()); err == nil {
			return persistentStateFile, nil
		}
	}
	defaultConfigFileAbsPath, err := c.defaultConfigFile(c.fileSystem, c.bds)
	if err != nil {
		return chezmoi.EmptyAbsPath, err
	}
	return defaultConfigFileAbsPath.Dir().Join(persistentStateFilename), nil
}

func (c *Config) promptChoice(prompt string, choices []string) (string, error) {
	promptWithChoices := fmt.Sprintf("%s [%s]? ", prompt, strings.Join(choices, ","))
	abbreviations := uniqueAbbreviations(choices)
	for {
		line, err := c.readLine(promptWithChoices)
		if err != nil {
			return "", err
		}
		if value, ok := abbreviations[strings.TrimSpace(line)]; ok {
			return value, nil
		}
	}
}

func (c *Config) readConfig() error {
	viper.SetConfigFile(c.configFileAbsPath.String())
	if c.configFormat != "" {
		viper.SetConfigType(c.configFormat.String())
	}
	viper.SetFs(afero.FromIOFS{FS: c.fileSystem})
	switch err := viper.ReadInConfig(); {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	}
	if err := viper.Unmarshal(c, viperDecodeConfigOptions...); err != nil {
		return err
	}
	return c.validateData()
}

func (c *Config) readLine(prompt string) (string, error) {
	_, err := c.stdout.Write([]byte(prompt))
	if err != nil {
		return "", err
	}
	line, err := bufio.NewReader(c.stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

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
	return c.baseSystem.RunCmd(cmd)
}

func (c *Config) runEditor(args []string) error {
	if err := c.persistentState.Close(); err != nil {
		return err
	}
	editor, editorArgs := c.editor()
	return c.run(chezmoi.EmptyAbsPath, editor, append(editorArgs, args...))
}

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

type targetRelPathsOptions struct {
	mustBeInSourceState bool
	recursive           bool
}

func (c *Config) targetRelPaths(sourceState *chezmoi.SourceState, args []string, options targetRelPathsOptions) ([]chezmoi.RelPath, error) {
	targetRelPaths := make([]chezmoi.RelPath, 0, len(args))
	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		targetRelPath, err := argAbsPath.TrimDirPrefix(c.DestDirAbsPath)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		if options.mustBeInSourceState {
			if _, ok := sourceState.Entry(targetRelPath); !ok {
				return nil, fmt.Errorf("%s: not in source state", arg)
			}
		}
		targetRelPaths = append(targetRelPaths, targetRelPath)
		if options.recursive {
			parentRelPath := targetRelPath
			// FIXME we should not call s.TargetRelPaths() here - risk of accidentally quadratic
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
	sort.Slice(targetRelPaths, func(i, j int) bool {
		return targetRelPaths[i] < targetRelPaths[j]
	})
	n := 1
	for i := 1; i < len(targetRelPaths); i++ {
		if targetRelPaths[i] != targetRelPaths[i-1] {
			targetRelPaths[n] = targetRelPaths[i]
			n++
		}
	}
	return targetRelPaths[:n], nil
}

func (c *Config) targetRelPathsBySourcePath(sourceState *chezmoi.SourceState, args []string) ([]chezmoi.RelPath, error) {
	targetRelPaths := make([]chezmoi.RelPath, 0, len(args))
	targetRelPathsBySourceRelPath := make(map[chezmoi.RelPath]chezmoi.RelPath)
	for targetRelPath, sourceStateEntry := range sourceState.Entries() {
		sourceRelPath := sourceStateEntry.SourceRelPath().RelPath()
		targetRelPathsBySourceRelPath[sourceRelPath] = targetRelPath
	}
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

func (c *Config) useBuiltinAgeAutoFunc() bool {
	if _, err := exec.LookPath(c.Age.Command); err == nil {
		return false
	}
	return true
}

func (c *Config) useBuiltinGitAutoFunc() bool {
	if _, err := exec.LookPath(c.Git.Command); err == nil {
		return false
	}
	return true
}

func (c *Config) validateData() error {
	return validateKeys(c.Data, identifierRx)
}

func (c *Config) writeOutput(data []byte) error {
	if c.outputAbsPath.Empty() || c.outputAbsPath == chezmoi.NewAbsPath("-") {
		_, err := c.stdout.Write(data)
		return err
	}
	return c.baseSystem.WriteFile(c.outputAbsPath, data, 0o666)
}

func (c *Config) writeOutputString(data string) error {
	return c.writeOutput([]byte(data))
}

// isDevVersion returns true if version is a development version (i.e. that the
// major, minor, and patch version numbers are all zero).
func isDevVersion(v *semver.Version) bool {
	return v.Major == 0 && v.Minor == 0 && v.Patch == 0
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
			versionElems = append(versionElems, "built at "+versionInfo.Date)
		}
		if versionInfo.BuiltBy != "" {
			versionElems = append(versionElems, "built by "+versionInfo.BuiltBy)
		}
		c.version = version
		c.versionInfo = versionInfo
		c.versionStr = strings.Join(versionElems, ", ")
		return nil
	}
}
