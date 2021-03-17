package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-vfs/v2"
	vfsafero "github.com/twpayne/go-vfsafero/v2"
	"github.com/twpayne/go-xdg/v4"
	"golang.org/x/term"

	"github.com/twpayne/chezmoi/assets/templates"
	"github.com/twpayne/chezmoi/internal/chezmoi"
	"github.com/twpayne/chezmoi/internal/git"
)

var defaultFormat = "json"

type purgeOptions struct {
	binary bool
}

type templateConfig struct {
	Options []string `mapstructure:"options"`
}

// A Config represents a configuration.
type Config struct {
	version     *semver.Version
	versionInfo VersionInfo
	versionStr  string

	bds *xdg.BaseDirectorySpecification

	fs              vfs.FS
	configFile      string
	baseSystem      chezmoi.System
	sourceSystem    chezmoi.System
	destSystem      chezmoi.System
	persistentState chezmoi.PersistentState
	color           bool

	// Global configuration, settable in the config file.
	SourceDir     string                 `mapstructure:"sourceDir"`
	DestDir       string                 `mapstructure:"destDir"`
	Umask         os.FileMode            `mapstructure:"umask"`
	Remove        bool                   `mapstructure:"remove"`
	Color         string                 `mapstructure:"color"`
	Data          map[string]interface{} `mapstructure:"data"`
	Template      templateConfig         `mapstructure:"template"`
	UseBuiltinGit string                 `mapstructure:"useBuiltinGit"`

	// Global configuration, not settable in the config file.
	cpuProfile    string
	debug         bool
	dryRun        bool
	exclude       *chezmoi.EntryTypeSet
	force         bool
	homeDir       string
	keepGoing     bool
	noPager       bool
	noTTY         bool
	outputStr     string
	sourcePath    bool
	verbose       bool
	templateFuncs template.FuncMap

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
	AGE        chezmoi.AGEEncryption `mapstructure:"age"`
	GPG        chezmoi.GPGEncryption `mapstructure:"gpg"`

	// Password manager data.
	gitHub  gitHubData
	keyring keyringData

	// Command configurations, settable in the config file.
	CD    cdCmdConfig    `mapstructure:"cd"`
	Diff  diffCmdConfig  `mapstructure:"diff"`
	Edit  editCmdConfig  `mapstructure:"edit"`
	Git   gitCmdConfig   `mapstructure:"git"`
	Merge mergeCmdConfig `mapstructure:"merge"`

	// Command configurations, not settable in the config file.
	add             addCmdConfig
	apply           applyCmdConfig
	archive         archiveCmdConfig
	data            dataCmdConfig
	dump            dumpCmdConfig
	executeTemplate executeTemplateCmdConfig
	_import         importCmdConfig
	init            initCmdConfig
	managed         managedCmdConfig
	purge           purgeCmdConfig
	secretKeyring   secretKeyringCmdConfig
	state           stateCmdConfig
	status          statusCmdConfig
	update          updateCmdConfig
	verify          verifyCmdConfig

	// Computed configuration.
	configFileAbsPath chezmoi.AbsPath
	homeDirAbsPath    chezmoi.AbsPath
	sourceDirAbsPath  chezmoi.AbsPath
	destDirAbsPath    chezmoi.AbsPath
	encryption        chezmoi.Encryption

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	ioregData ioregData
}

// A configOption sets and option on a Config.
type configOption func(*Config) error

type configState struct {
	ConfigTemplateContentsSHA256 chezmoi.HexBytes `json:"configTemplateContentsSHA256" yaml:"configTemplateContentsSHA256"`
}

var (
	persistentStateFilename = chezmoi.RelPath("chezmoistate.boltdb")
	configStateKey          = []byte("configState")

	identifierRx = regexp.MustCompile(`\A[\pL_][\pL\p{Nd}_]*\z`)
	whitespaceRx = regexp.MustCompile(`\s+`)
)

// newConfig creates a new Config with the given options.
func newConfig(options ...configOption) (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	normalizedHomeDir, err := chezmoi.NormalizePath(homeDir)
	if err != nil {
		return nil, err
	}

	bds, err := xdg.NewBaseDirectorySpecification()
	if err != nil {
		return nil, err
	}

	c := &Config{
		bds:     bds,
		fs:      vfs.OSFS,
		homeDir: homeDir,
		DestDir: homeDir,
		Umask:   chezmoi.Umask,
		Color:   "auto",
		Diff: diffCmdConfig{
			Pager:   os.Getenv("PAGER"),
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll &^ chezmoi.EntryTypeScripts),
		},
		Edit: editCmdConfig{
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypeDirs | chezmoi.EntryTypeFiles | chezmoi.EntryTypeSymlinks | chezmoi.EntryTypeEncrypted),
		},
		Git: gitCmdConfig{
			Command: "git",
		},
		Merge: mergeCmdConfig{
			Command: "vimdiff",
		},
		Template: templateConfig{
			Options: chezmoi.DefaultTemplateOptions,
		},
		templateFuncs: sprig.TxtFuncMap(),
		exclude:       chezmoi.NewEntryTypeSet(chezmoi.EntryTypesNone),
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
		AGE: chezmoi.AGEEncryption{
			Command: "age",
			Suffix:  ".age",
		},
		GPG: chezmoi.GPGEncryption{
			Command: "gpg",
			Suffix:  ".asc",
		},
		add: addCmdConfig{
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		apply: applyCmdConfig{
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		archive: archiveCmdConfig{
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		data: dataCmdConfig{
			format: defaultFormat,
		},
		dump: dumpCmdConfig{
			format:    defaultFormat,
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		_import: importCmdConfig{
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
		},
		managed: managedCmdConfig{
			include: chezmoi.NewEntryTypeSet(chezmoi.EntryTypeDirs | chezmoi.EntryTypeFiles | chezmoi.EntryTypeSymlinks | chezmoi.EntryTypeEncrypted),
		},
		state: stateCmdConfig{
			dump: stateDumpCmdConfig{
				format: defaultFormat,
			},
		},
		status: statusCmdConfig{
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		update: updateCmdConfig{
			apply:     true,
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll),
			recursive: true,
		},
		verify: verifyCmdConfig{
			include:   chezmoi.NewEntryTypeSet(chezmoi.EntryTypesAll &^ chezmoi.EntryTypeScripts),
			recursive: true,
		},

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,

		homeDirAbsPath: normalizedHomeDir,
	}

	for key, value := range map[string]interface{}{
		"bitwarden":                c.bitwardenTemplateFunc,
		"bitwardenAttachment":      c.bitwardenAttachmentTemplateFunc,
		"bitwardenFields":          c.bitwardenFieldsTemplateFunc,
		"gitHubKeys":               c.gitHubKeysTemplateFunc,
		"gopass":                   c.gopassTemplateFunc,
		"include":                  c.includeTemplateFunc,
		"ioreg":                    c.ioregTemplateFunc,
		"joinPath":                 c.joinPathTemplateFunc,
		"keepassxc":                c.keepassxcTemplateFunc,
		"keepassxcAttribute":       c.keepassxcAttributeTemplateFunc,
		"keyring":                  c.keyringTemplateFunc,
		"lastpass":                 c.lastpassTemplateFunc,
		"lastpassRaw":              c.lastpassRawTemplateFunc,
		"lookPath":                 c.lookPathTemplateFunc,
		"onepassword":              c.onepasswordTemplateFunc,
		"onepasswordDetailsFields": c.onepasswordDetailsFieldsTemplateFunc,
		"onepasswordDocument":      c.onepasswordDocumentTemplateFunc,
		"output":                   c.outputTemplateFunc,
		"pass":                     c.passTemplateFunc,
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

	c.configFile = string(defaultConfigFile(c.fs, c.bds))
	c.SourceDir = string(defaultSourceDir(c.fs, c.bds))

	c.homeDirAbsPath, err = chezmoi.NormalizePath(c.homeDir)
	if err != nil {
		return nil, err
	}
	c._import.destination = string(c.homeDirAbsPath)

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
	recursive    bool
	umask        os.FileMode
	preApplyFunc chezmoi.PreApplyFunc
}

func (c *Config) applyArgs(targetSystem chezmoi.System, targetDirAbsPath chezmoi.AbsPath, args []string, options applyArgsOptions) error {
	sourceState, err := c.sourceState()
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
		Include:      options.include.Sub(c.exclude),
		PreApplyFunc: options.preApplyFunc,
		Umask:        options.umask,
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

	keptGoingAfterErr := false
	for _, targetRelPath := range targetRelPaths {
		switch err := sourceState.Apply(targetSystem, c.destSystem, c.persistentState, targetDirAbsPath, targetRelPath, applyOptions); {
		case errors.Is(err, chezmoi.Skip):
			continue
		case err != nil && c.keepGoing:
			c.errorf("%v", err)
			keptGoingAfterErr = true
		case err != nil:
			return err
		}
	}
	if keptGoingAfterErr {
		return ErrExitCode(1)
	}

	return nil
}

func (c *Config) cmdOutput(dirAbsPath chezmoi.AbsPath, name string, args []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	if dirAbsPath != "" {
		dirRawAbsPath, err := c.baseSystem.RawPath(dirAbsPath)
		if err != nil {
			return nil, err
		}
		cmd.Dir = string(dirRawAbsPath)
	}
	return c.baseSystem.IdempotentCmdOutput(cmd)
}

func (c *Config) defaultPreApplyFunc(targetRelPath chezmoi.RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *chezmoi.EntryState) error {
	switch {
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
			return ErrExitCode(1)
		default:
			return nil
		}
	}
}

func (c *Config) defaultTemplateData() map[string]interface{} {
	data := map[string]interface{}{
		"arch":      runtime.GOARCH,
		"homeDir":   c.homeDir,
		"homedir":   c.homeDir, // TODO Remove in version 2.1.
		"os":        runtime.GOOS,
		"sourceDir": c.sourceDirAbsPath,
		"version": map[string]interface{}{
			"builtBy": c.versionInfo.BuiltBy,
			"commit":  c.versionInfo.Commit,
			"date":    c.versionInfo.Date,
			"version": c.versionInfo.Version,
		},
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
	// On Windows, the user's group ID returned by user.Current() is an SID and
	// no further useful lookup is possible with Go's standard library.
	//
	// Since neither the username nor the group are likely widely used in
	// templates, leave these variables unset if their values cannot be
	// determined. Unset variables will trigger template errors if used,
	// alerting the user to the problem and allowing them to find alternative
	// solutions.
	if currentUser, err := user.Current(); err == nil {
		data["username"] = currentUser.Username
		if runtime.GOOS != "windows" {
			if group, err := user.LookupGroupId(currentUser.Gid); err == nil {
				data["group"] = group.Name
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
		user, ok := os.LookupEnv("USER")
		if ok {
			data["username"] = user
		} else {
			log.Debug().
				Str("key", "USER").
				Bool("ok", ok).
				Msg("os.LookupEnv")
		}
	}

	if fqdnHostname, err := chezmoi.FQDNHostname(c.fs); err == nil && fqdnHostname != "" {
		data["fqdnHostname"] = fqdnHostname
	} else {
		log.Debug().
			Err(err).
			Msg("chezmoi.EtcHostsFQDNHostname")
	}

	if hostname, err := os.Hostname(); err == nil {
		data["hostname"] = strings.SplitN(hostname, ".", 2)[0]
	} else {
		log.Debug().
			Err(err).
			Msg("os.Hostname")
	}

	if kernelInfo, err := chezmoi.KernelInfo(c.fs); err == nil {
		data["kernel"] = kernelInfo
	} else {
		log.Debug().
			Err(err).
			Msg("chezmoi.KernelInfo")
	}

	if osRelease, err := chezmoi.OSRelease(c.fs); err == nil {
		data["osRelease"] = upperSnakeCaseToCamelCaseMap(osRelease)
	} else {
		log.Debug().
			Err(err).
			Msg("chezmoi.OSRelease")
	}

	return map[string]interface{}{
		"chezmoi": data,
	}
}

func (c *Config) destAbsPathInfos(sourceState *chezmoi.SourceState, args []string, recursive, follow bool) (map[chezmoi.AbsPath]os.FileInfo, error) {
	destAbsPathInfos := make(map[chezmoi.AbsPath]os.FileInfo)
	for _, arg := range args {
		destAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		if _, err := destAbsPath.TrimDirPrefix(c.destDirAbsPath); err != nil {
			return nil, err
		}
		if recursive {
			if err := chezmoi.Walk(c.destSystem, destAbsPath, func(destAbsPath chezmoi.AbsPath, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if follow && info.Mode()&os.ModeType == os.ModeSymlink {
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
			var info os.FileInfo
			if follow {
				info, err = c.destSystem.Stat(destAbsPath)
			} else {
				info, err = c.destSystem.Lstat(destAbsPath)
			}
			if err != nil {
				return nil, err
			}
			if err := sourceState.AddDestAbsPathInfos(destAbsPathInfos, c.destSystem, destAbsPath, info); err != nil {
				return nil, err
			}
		}
	}
	return destAbsPathInfos, nil
}

func (c *Config) diffFile(path chezmoi.RelPath, fromData []byte, fromMode os.FileMode, toData []byte, toMode os.FileMode) error {
	var sb strings.Builder
	unifiedEncoder := diff.NewUnifiedEncoder(&sb, diff.DefaultContextLines)
	if c.color {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	diffPatch, err := chezmoi.DiffPatch(path, fromData, fromMode, toData, toMode)
	if err != nil {
		return err
	}
	if err := unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return c.diffPager(sb.String())
}

func (c *Config) diffPager(output string) error {
	if c.noPager || c.Diff.Pager == "" {
		return c.writeOutputString(output)
	}

	// If the pager command contains any spaces, assume that it is a full
	// shell command to be executed via the user's shell. Otherwise, execute
	// it directly.
	var pagerCmd *exec.Cmd
	if strings.IndexFunc(c.Diff.Pager, unicode.IsSpace) != -1 {
		shell, _ := shell.CurrentUserShell()
		pagerCmd = exec.Command(shell, "-c", c.Diff.Pager)
	} else {
		//nolint:gosec
		pagerCmd = exec.Command(c.Diff.Pager)
	}
	pagerCmd.Stdin = bytes.NewBufferString(output)
	pagerCmd.Stdout = c.stdout
	pagerCmd.Stderr = c.stderr
	return pagerCmd.Run()
}

func (c *Config) doPurge(purgeOptions *purgeOptions) error {
	if c.persistentState != nil {
		if err := c.persistentState.Close(); err != nil {
			return err
		}
	}

	absSlashPersistentStateFile := c.persistentStateFile()
	absPaths := chezmoi.AbsPaths{
		c.configFileAbsPath.Dir(),
		c.configFileAbsPath,
		absSlashPersistentStateFile,
		c.sourceDirAbsPath,
	}
	if purgeOptions != nil && purgeOptions.binary {
		executable, err := os.Executable()
		if err == nil {
			absPaths = append(absPaths, chezmoi.AbsPath(executable))
		}
	}

	// Remove all paths that exist.
	for _, absPath := range absPaths {
		switch _, err := c.destSystem.Stat(absPath); {
		case os.IsNotExist(err):
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
		case os.IsPermission(err):
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

func (c *Config) findConfigTemplate() (chezmoi.RelPath, string, []byte, error) {
	for _, ext := range viper.SupportedExts {
		filename := chezmoi.RelPath(chezmoi.Prefix + "." + ext + chezmoi.TemplateSuffix)
		contents, err := c.baseSystem.ReadFile(c.sourceDirAbsPath.Join(filename))
		switch {
		case os.IsNotExist(err):
			continue
		case err != nil:
			return "", "", nil, err
		}
		return chezmoi.RelPath("chezmoi." + ext), ext, contents, nil
	}
	return "", "", nil, nil
}

func (c *Config) gitAutoAdd() (*git.Status, error) {
	if err := c.run(c.sourceDirAbsPath, c.Git.Command, []string{"add", "."}); err != nil {
		return nil, err
	}
	output, err := c.cmdOutput(c.sourceDirAbsPath, c.Git.Command, []string{"status", "--porcelain=v2"})
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
	return c.run(c.sourceDirAbsPath, c.Git.Command, []string{"commit", "--message", commitMessage.String()})
}

func (c *Config) gitAutoPush(status *git.Status) error {
	if status.Empty() {
		return nil
	}
	return c.run(c.sourceDirAbsPath, c.Git.Command, []string{"push"})
}

func (c *Config) makeRunEWithSourceState(runE func(*cobra.Command, []string, *chezmoi.SourceState) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		sourceState, err := c.sourceState()
		if err != nil {
			return err
		}
		return runE(cmd, args, sourceState)
	}
}

func (c *Config) marshal(formatStr string, data interface{}) error {
	var format chezmoi.Format
	switch formatStr {
	case "json":
		format = chezmoi.JSONFormat
	case "yaml":
		format = chezmoi.YAMLFormat
	default:
		return fmt.Errorf("%s: unknown format", formatStr)
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

	persistentFlags.StringVar(&c.Color, "color", c.Color, "colorize diffs")
	persistentFlags.StringVarP(&c.DestDir, "destination", "D", c.DestDir, "destination directory")
	persistentFlags.BoolVar(&c.Remove, "remove", c.Remove, "remove targets")
	persistentFlags.StringVarP(&c.SourceDir, "source", "S", c.SourceDir, "source directory")
	persistentFlags.StringVar(&c.UseBuiltinGit, "use-builtin-git", c.UseBuiltinGit, "use builtin git")
	for _, key := range []string{
		"color",
		"destination",
		"remove",
		"source",
	} {
		if err := viper.BindPFlag(key, persistentFlags.Lookup(key)); err != nil {
			return nil, err
		}
	}

	persistentFlags.StringVarP(&c.configFile, "config", "c", c.configFile, "config file")
	persistentFlags.StringVar(&c.cpuProfile, "cpu-profile", c.cpuProfile, "write CPU profile to file")
	persistentFlags.BoolVarP(&c.dryRun, "dry-run", "n", c.dryRun, "dry run")
	persistentFlags.VarP(c.exclude, "exclude", "x", "exclude entry types")
	persistentFlags.BoolVar(&c.force, "force", c.force, "force")
	persistentFlags.BoolVarP(&c.keepGoing, "keep-going", "k", c.keepGoing, "keep going as far as possible after an error")
	persistentFlags.BoolVar(&c.noPager, "no-pager", c.noPager, "do not use the pager")
	persistentFlags.BoolVar(&c.noTTY, "no-tty", c.noTTY, "don't attempt to get a TTY for reading passwords")
	persistentFlags.BoolVar(&c.sourcePath, "source-path", c.sourcePath, "specify targets by source path")
	persistentFlags.BoolVarP(&c.verbose, "verbose", "v", c.verbose, "verbose")
	persistentFlags.StringVarP(&c.outputStr, "output", "o", c.outputStr, "output file")
	persistentFlags.BoolVar(&c.debug, "debug", c.debug, "write debug logs")

	for _, err := range []error{
		rootCmd.MarkPersistentFlagFilename("config"),
		rootCmd.MarkPersistentFlagFilename("cpu-profile"),
		rootCmd.MarkPersistentFlagDirname("destination"),
		rootCmd.MarkPersistentFlagFilename("output"),
		rootCmd.MarkPersistentFlagDirname("source"),
	} {
		if err != nil {
			return nil, err
		}
	}

	rootCmd.SetHelpCommand(c.newHelpCmd())
	for _, newCmdFunc := range []func() *cobra.Command{
		c.newAddCmd,
		c.newApplyCmd,
		c.newArchiveCmd,
		c.newCatCmd,
		c.newCDCmd,
		c.newChattrCmd,
		c.newCompletionCmd,
		c.newDataCmd,
		c.newDiffCmd,
		c.newDocsCmd,
		c.newDoctorCmd,
		c.newDumpCmd,
		c.newEditCmd,
		c.newEditConfigCmd,
		c.newExecuteTemplateCmd,
		c.newForgetCmd,
		c.newGitCmd,
		c.newImportCmd,
		c.newInitCmd,
		c.newManagedCmd,
		c.newMergeCmd,
		c.newPurgeCmd,
		c.newRemoveCmd,
		c.newSecretCmd,
		c.newSourcePathCmd,
		c.newStateCmd,
		c.newStatusCmd,
		c.newUnmanagedCmd,
		c.newUpdateCmd,
		c.newVerifyCmd,
	} {
		rootCmd.AddCommand(newCmdFunc())
	}

	return rootCmd, nil
}

func (c *Config) persistentPostRunRootE(cmd *cobra.Command, args []string) error {
	defer pprof.StopCPUProfile()

	if c.persistentState != nil {
		if err := c.persistentState.Close(); err != nil {
			return err
		}
	}

	if boolAnnotation(cmd, modifiesConfigFile) {
		// Warn the user of any errors reading the config file.
		v := viper.New()
		v.SetFs(vfsafero.NewAferoFS(c.fs))
		v.SetConfigFile(string(c.configFileAbsPath))
		err := v.ReadInConfig()
		if err == nil {
			err = v.Unmarshal(&Config{})
		}
		if err != nil {
			cmd.Printf("warning: %s: %v\n", c.configFileAbsPath, err)
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

	return nil
}

func (c *Config) persistentPreRunRootE(cmd *cobra.Command, args []string) error {
	if c.cpuProfile != "" {
		f, err := os.Create(c.cpuProfile)
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
	}

	var err error
	c.configFileAbsPath, err = chezmoi.NewAbsPathFromExtPath(c.configFile, c.homeDirAbsPath)
	if err != nil {
		return err
	}

	if err := c.readConfig(); err != nil {
		if !boolAnnotation(cmd, doesNotRequireValidConfig) {
			return fmt.Errorf("invalid config: %s: %w", c.configFile, err)
		}
		cmd.Printf("warning: %s: %v\n", c.configFile, err)
	}

	if c.Color == "" || strings.ToLower(c.Color) == "auto" {
		if _, ok := os.LookupEnv("NO_COLOR"); ok {
			c.color = false
		} else if stdout, ok := c.stdout.(*os.File); ok {
			c.color = term.IsTerminal(int(stdout.Fd()))
		} else {
			c.color = false
		}
	} else if color, err := parseBool(c.Color); err == nil {
		c.color = color
	} else if !boolAnnotation(cmd, doesNotRequireValidConfig) {
		return fmt.Errorf("%s: invalid color value", c.Color)
	}

	if c.color {
		if err := enableVirtualTerminalProcessing(c.stdout); err != nil {
			return err
		}
	}

	if c.sourceDirAbsPath, err = chezmoi.NewAbsPathFromExtPath(c.SourceDir, c.homeDirAbsPath); err != nil {
		return err
	}
	if c.destDirAbsPath, err = chezmoi.NewAbsPathFromExtPath(c.DestDir, c.homeDirAbsPath); err != nil {
		return err
	}

	log.Logger = log.Output(zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.Out = c.stderr
			w.NoColor = !c.color
			w.TimeFormat = time.RFC3339
		},
	))
	if c.debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	c.baseSystem = chezmoi.NewRealSystem(c.fs)
	if c.debug {
		c.baseSystem = chezmoi.NewDebugSystem(c.baseSystem)
	}

	switch {
	case cmd.Annotations[persistentStateMode] == persistentStateModeEmpty:
		c.persistentState = chezmoi.NewMockPersistentState()
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadOnly:
		persistentStateFile := c.persistentStateFile()
		c.persistentState, err = chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFile, chezmoi.BoltPersistentStateReadOnly)
		if err != nil {
			return err
		}
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadMockWrite:
		fallthrough
	case cmd.Annotations[persistentStateMode] == persistentStateModeReadWrite && c.dryRun:
		persistentStateFile := c.persistentStateFile()
		persistentState, err := chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFile, chezmoi.BoltPersistentStateReadOnly)
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
		persistentStateFile := c.persistentStateFile()
		c.persistentState, err = chezmoi.NewBoltPersistentState(c.baseSystem, persistentStateFile, chezmoi.BoltPersistentStateReadWrite)
		if err != nil {
			return err
		}
	default:
		c.persistentState = nil
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
		c.sourceSystem = chezmoi.NewGitDiffSystem(c.sourceSystem, c.stdout, c.sourceDirAbsPath, c.color)
		c.destSystem = chezmoi.NewGitDiffSystem(c.destSystem, c.stdout, c.destDirAbsPath, c.color)
	}

	switch c.Encryption {
	case "age":
		c.encryption = &c.AGE
	case "gpg":
		c.encryption = &c.GPG
	case "":
		c.encryption = chezmoi.NoEncryption{}
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
		if err := chezmoi.MkdirAll(c.baseSystem, c.sourceDirAbsPath, 0o777); err != nil {
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
			if err := chezmoi.MkdirAll(c.baseSystem, chezmoi.AbsPath(c.bds.RuntimeDir), 0o700); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Config) persistentStateFile() chezmoi.AbsPath {
	if c.configFile != "" {
		return chezmoi.AbsPath(c.configFile).Dir().Join(persistentStateFilename)
	}
	for _, configDir := range c.bds.ConfigDirs {
		configDirAbsPath := chezmoi.AbsPath(configDir)
		persistentStateFile := configDirAbsPath.Join(chezmoi.RelPath("chezmoi"), persistentStateFilename)
		if _, err := os.Stat(string(persistentStateFile)); err == nil {
			return persistentStateFile
		}
	}
	return defaultConfigFile(c.fs, c.bds).Dir().Join(persistentStateFilename)
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
	v := viper.New()
	v.SetConfigFile(string(c.configFileAbsPath))
	v.SetFs(vfsafero.NewAferoFS(c.fs))
	switch err := v.ReadInConfig(); {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	}
	if err := v.Unmarshal(c); err != nil {
		return err
	}
	return c.validateData()
}

func (c *Config) readLine(prompt string) (string, error) {
	var line string
	if err := c.withTerminal(prompt, func(t terminal) error {
		var err error
		line, err = t.ReadLine()
		return err
	}); err != nil {
		return "", err
	}
	return line, nil
}

func (c *Config) readPassword(prompt string) (string, error) {
	var password string
	if err := c.withTerminal("", func(t terminal) error {
		var err error
		password, err = t.ReadPassword(prompt)
		return err
	}); err != nil {
		return "", err
	}
	return password, nil
}

func (c *Config) run(dir chezmoi.AbsPath, name string, args []string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		dirRawAbsPath, err := c.baseSystem.RawPath(dir)
		if err != nil {
			return err
		}
		cmd.Dir = string(dirRawAbsPath)
	}
	cmd.Stdin = c.stdin
	cmd.Stdout = c.stdout
	cmd.Stderr = c.stderr
	return c.baseSystem.RunCmd(cmd)
}

func (c *Config) runEditor(args []string) error {
	editor, editorArgs := c.editor()
	return c.run("", editor, append(editorArgs, args...))
}

func (c *Config) sourceAbsPaths(sourceState *chezmoi.SourceState, args []string) (chezmoi.AbsPaths, error) {
	targetRelPaths, err := c.targetRelPaths(sourceState, args, targetRelPathsOptions{
		mustBeInSourceState: true,
	})
	if err != nil {
		return nil, err
	}
	sourceAbsPaths := make(chezmoi.AbsPaths, 0, len(targetRelPaths))
	for _, targetRelPath := range targetRelPaths {
		sourceAbsPath := c.sourceDirAbsPath.Join(sourceState.MustEntry(targetRelPath).SourceRelPath().RelPath())
		sourceAbsPaths = append(sourceAbsPaths, sourceAbsPath)
	}
	return sourceAbsPaths, nil
}

func (c *Config) sourceState() (*chezmoi.SourceState, error) {
	s := chezmoi.NewSourceState(
		chezmoi.WithDefaultTemplateDataFunc(c.defaultTemplateData),
		chezmoi.WithDestDir(c.destDirAbsPath),
		chezmoi.WithEncryption(c.encryption),
		chezmoi.WithPriorityTemplateData(c.Data),
		chezmoi.WithSourceDir(c.sourceDirAbsPath),
		chezmoi.WithSystem(c.sourceSystem),
		chezmoi.WithTemplateFuncs(c.templateFuncs),
		chezmoi.WithTemplateOptions(c.Template.Options),
	)

	if err := s.Read(); err != nil {
		return nil, err
	}

	if minVersion := s.MinVersion(); c.version != nil && !isDevVersion(c.version) && c.version.LessThan(minVersion) {
		return nil, fmt.Errorf("source state requires version %s or later, chezmoi is version %s", minVersion, c.version)
	}

	return s, nil
}

type targetRelPathsOptions struct {
	mustBeInSourceState bool
	recursive           bool
}

func (c *Config) targetRelPaths(sourceState *chezmoi.SourceState, args []string, options targetRelPathsOptions) (chezmoi.RelPaths, error) {
	targetRelPaths := make(chezmoi.RelPaths, 0, len(args))
	for _, arg := range args {
		argAbsPath, err := chezmoi.NewAbsPathFromExtPath(arg, c.homeDirAbsPath)
		if err != nil {
			return nil, err
		}
		targetRelPath, err := argAbsPath.TrimDirPrefix(c.destDirAbsPath)
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

func (c *Config) targetRelPathsBySourcePath(sourceState *chezmoi.SourceState, args []string) (chezmoi.RelPaths, error) {
	targetRelPaths := make(chezmoi.RelPaths, 0, len(args))
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
		sourceRelPath, err := argAbsPath.TrimDirPrefix(c.sourceDirAbsPath)
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

func (c *Config) useBuiltinGit() (bool, error) {
	if c.UseBuiltinGit == "" || strings.ToLower(c.UseBuiltinGit) == "auto" {
		if _, err := exec.LookPath(c.Git.Command); err == nil {
			return false, nil
		}
		return true, nil
	}
	return parseBool(c.UseBuiltinGit)
}

func (c *Config) validateData() error {
	return validateKeys(c.Data, identifierRx)
}

func (c *Config) withTerminal(prompt string, f func(terminal) error) error {
	if c.noTTY {
		return f(newNullTerminal(c.stdin))
	}

	if runtime.GOOS == "windows" {
		return f(newDumbTerminal(c.stdin, c.stdout, prompt))
	}

	if stdinFile, ok := c.stdin.(*os.File); ok && term.IsTerminal(int(stdinFile.Fd())) {
		fd := int(stdinFile.Fd())
		width, height, err := term.GetSize(fd)
		if err != nil {
			return err
		}
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}
		defer func() {
			_ = term.Restore(fd, oldState)
		}()
		t := term.NewTerminal(struct {
			io.Reader
			io.Writer
		}{
			Reader: c.stdin,
			Writer: c.stdout,
		}, prompt)
		if err := t.SetSize(width, height); err != nil {
			return err
		}
		return f(t)
	}

	devTTY, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer devTTY.Close()
	fd := int(devTTY.Fd())
	width, height, err := term.GetSize(fd)
	if err != nil {
		return err
	}
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return err
	}
	defer func() {
		_ = term.Restore(fd, oldState)
	}()
	t := term.NewTerminal(devTTY, prompt)
	if err := t.SetSize(width, height); err != nil {
		return err
	}
	return f(t)
}

func (c *Config) writeOutput(data []byte) error {
	if c.outputStr == "" || c.outputStr == "-" {
		_, err := c.stdout.Write(data)
		return err
	}
	return c.baseSystem.WriteFile(chezmoi.AbsPath(c.outputStr), data, 0o666)
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
