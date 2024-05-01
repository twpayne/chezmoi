package cmd

// FIXME add check for $TMPDIR mount options (specifically noexec)

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v62/github"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-xdg/v6"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoigit"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

// A checkResult is the result of a check.
type checkResult int

const (
	checkResultFailed  checkResult = -2 // The check could not be completed.
	checkResultSkipped checkResult = -1 // The check was skipped.
	checkResultOK      checkResult = 0  // The check completed and did not find any problems.
	checkResultInfo    checkResult = 1  // The check completed and found something interesting, but not a problem.
	checkResultWarning checkResult = 2  // The check completed and found something that might indicate a problem.
	checkResultError   checkResult = 3  // The check completed and found a definite problem.
)

// A gitStatus is the status of a git working copy.
type gitStatus string

const (
	gitStatusNotAWorkingCopy gitStatus = ""
	gitStatusClean           gitStatus = "clean"
	gitStatusDirty           gitStatus = "dirty"
	gitStatusError           gitStatus = "error"
)

// A check is an individual check.
type check interface {
	Name() string                                                                    // Name returns the check's name.
	Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) // Run runs the check.
}

var checkResultStr = map[checkResult]string{
	checkResultFailed:  "failed",
	checkResultSkipped: "skipped",
	checkResultOK:      "ok",
	checkResultInfo:    "info",
	checkResultWarning: "warning",
	checkResultError:   "error",
}

// An argsCheck checks that arguments for a binary.
type argsCheck struct {
	name    string
	command string
	args    []string
}

// A binaryCheck checks that a binary called name is installed and optionally at
// least version minVersion.
type binaryCheck struct {
	name        string
	binaryname  string
	ifNotSet    checkResult
	ifNotExist  checkResult
	versionArgs []string
	versionRx   *regexp.Regexp
	minVersion  *semver.Version
}

// A configFileCheck checks that only one config file exists and that is
// readable.
type configFileCheck struct {
	basename chezmoi.RelPath
	bds      *xdg.BaseDirectorySpecification
	expected chezmoi.AbsPath
}

// A dirCheck checks that a directory exists.
type dirCheck struct {
	name    string
	dirname chezmoi.AbsPath
}

// An executableCheck checks the executable.
type executableCheck struct{}

// A fileCheck checks that a file exists.
type fileCheck struct {
	name       string
	filename   chezmoi.AbsPath
	ifNotSet   checkResult
	ifNotExist checkResult
}

// A goVersionCheck checks the Go version.
type goVersionCheck struct{}

// A latestVersionCheck checks the latest version.
type latestVersionCheck struct {
	httpClient    *http.Client
	httpClientErr error
	version       semver.Version
}

// An osArchCheck checks that runtime.GOOS and runtime.GOARCH are supported.
type osArchCheck struct{}

// A skippedCheck is a check that is skipped.
type skippedCheck struct{}

// A suspiciousEntriesCheck checks that a source directory does not contain any
// suspicious files.
type suspiciousEntriesCheck struct {
	dirname chezmoi.AbsPath
}

// A upgradeMethodCheck checks the upgrade method.
type upgradeMethodCheck struct{}

// A versionCheck checks the version information.
type versionCheck struct {
	versionInfo VersionInfo
	versionStr  string
}

func (c *Config) newDoctorCmd() *cobra.Command {
	doctorCmd := &cobra.Command{
		Args:    cobra.NoArgs,
		Use:     "doctor",
		Short:   "Check your system for potential problems",
		Example: example("doctor"),
		Long:    mustLongHelp("doctor"),
		RunE:    c.runDoctorCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			runsCommands,
		),
	}

	return doctorCmd
}

func (c *Config) runDoctorCmd(cmd *cobra.Command, args []string) error {
	homeDirAbsPath, err := chezmoi.HomeDirAbsPath()
	if err != nil {
		return err
	}
	httpClient, httpClientErr := c.getHTTPClient()
	shellCommand, _ := shell.CurrentUserShell()
	shellCommand, shellArgs, _ := parseCommand(shellCommand, nil)
	cdCommand, cdArgs, _ := c.cdCommand()
	editCommand, editArgs, _ := c.editor(nil)
	checks := []check{
		&versionCheck{
			versionInfo: c.versionInfo,
			versionStr:  c.versionStr,
		},
		&latestVersionCheck{
			httpClient:    httpClient,
			httpClientErr: httpClientErr,
			version:       c.version,
		},
		osArchCheck{},
		unameCheck{},
		systeminfoCheck{},
		goVersionCheck{},
		executableCheck{},
		upgradeMethodCheck{},
		&configFileCheck{
			basename: chezmoiRelPath,
			bds:      c.bds,
			expected: c.getConfigFileAbsPath(),
		},
		&dirCheck{
			name:    "source-dir",
			dirname: c.SourceDirAbsPath,
		},
		&suspiciousEntriesCheck{
			dirname: c.SourceDirAbsPath,
		},
		&dirCheck{
			name:    "working-tree",
			dirname: c.WorkingTreeAbsPath,
		},
		&dirCheck{
			name:    "dest-dir",
			dirname: c.DestDirAbsPath,
		},
		umaskCheck{},
		&binaryCheck{
			name:       "cd-command",
			binaryname: cdCommand,
			ifNotSet:   checkResultError,
			ifNotExist: checkResultError,
		},
		&argsCheck{
			name:    "cd-args",
			command: cdCommand,
			args:    cdArgs,
		},
		&binaryCheck{
			name:       "diff-command",
			binaryname: c.Diff.Command,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultWarning,
		},
		&binaryCheck{
			name:       "edit-command",
			binaryname: editCommand,
			ifNotSet:   checkResultWarning,
			ifNotExist: checkResultWarning,
		},
		&argsCheck{
			name:    "edit-args",
			command: editCommand,
			args:    editArgs,
		},
		&binaryCheck{
			name:        "git-command",
			binaryname:  c.Git.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^git\s+version\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "merge-command",
			binaryname: c.Merge.Command,
			ifNotSet:   checkResultWarning,
			ifNotExist: checkResultWarning,
		},
		&binaryCheck{
			name:       "shell-command",
			binaryname: shellCommand,
			ifNotSet:   checkResultError,
			ifNotExist: checkResultError,
		},
		&argsCheck{
			name:    "shell-args",
			command: shellCommand,
			args:    shellArgs,
		},
		&binaryCheck{
			name:        "age-command",
			binaryname:  c.Age.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(\d+\.\d+\.\d+\S*)`),
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
		},
		&binaryCheck{
			name:        "gpg-command",
			binaryname:  c.GPG.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(?m)^gpg\s+\(.*?\)\s+(\d+\.\d+\.\d+)`),
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
		},
		&binaryCheck{
			name:        "pinentry-command",
			binaryname:  c.PINEntry.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^\S+\s+\(pinentry\)\s+(\d+\.\d+\.\d+)`),
			ifNotSet:    checkResultInfo,
			ifNotExist:  checkResultWarning,
		},
		&binaryCheck{
			name:        "1password-command",
			binaryname:  c.Onepassword.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   onepasswordVersionRx,
			minVersion:  &onepasswordMinVersion,
		},
		&binaryCheck{
			name:        "bitwarden-command",
			binaryname:  c.Bitwarden.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(?m)^(\d+\.\d+\.\d+)$`),
		},
		&binaryCheck{
			name:        "bitwarden-secrets-command",
			binaryname:  c.BitwardenSecrets.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`Bitwarden\s+Secrets\s+CLI\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "dashlane-command",
			binaryname:  c.Dashlane.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "doppler-command",
			binaryname:  c.Doppler.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "gopass-command",
			binaryname:  c.Gopass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: gopassVersionArgs,
			versionRx:   gopassVersionRx,
			minVersion:  &gopassMinVersion,
		},
		&binaryCheck{
			name:        "keepassxc-command",
			binaryname:  c.Keepassxc.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
			minVersion:  &keepassxcMinVersion,
		},
		&fileCheck{
			name:       "keepassxc-db",
			filename:   c.Keepassxc.Database,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultInfo,
		},
		&binaryCheck{
			name:        "keeper-command",
			binaryname:  c.Keeper.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Commander\s+Version:\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "lastpass-command",
			binaryname:  c.Lastpass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: lastpassVersionArgs,
			versionRx:   lastpassVersionRx,
			minVersion:  &lastpassMinVersion,
		},
		&binaryCheck{
			name:        "pass-command",
			binaryname:  c.Pass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "passhole-command",
			binaryname:  c.Passhole.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
			minVersion:  &passholeMinVersion,
		},
		&binaryCheck{
			name:        "rbw-command",
			binaryname:  c.RBW.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^rbw\s+(\d+\.\d+\.\d+)`),
			minVersion:  &rbwMinVersion,
		},
		&binaryCheck{
			name:        "vault-command",
			binaryname:  c.Vault.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "vlt-command",
			binaryname:  c.HCPVaultSecrets.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
			minVersion:  &vltMinVersion,
		},
		&binaryCheck{
			name:       "secret-command",
			binaryname: c.Secret.Command,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultInfo,
		},
	}

	worstResult := checkResultOK
	resultWriter := tabwriter.NewWriter(c.stdout, 3, 0, 3, ' ', 0)
	fmt.Fprint(resultWriter, "RESULT\tCHECK\tMESSAGE\n")
	for _, check := range checks {
		checkResult, message := check.Run(c.baseSystem, homeDirAbsPath)
		if checkResult == checkResultSkipped {
			continue
		}
		// Conceal the user's actual home directory in the message as the
		// output of chezmoi doctor is often posted publicly and would otherwise
		// reveal the user's username.
		message = strings.ReplaceAll(message, homeDirAbsPath.String(), "~")
		fmt.Fprintf(resultWriter, "%s\t%s\t%s\n", checkResultStr[checkResult], check.Name(), message)
		if checkResult > worstResult {
			worstResult = checkResult
		}
	}
	resultWriter.Flush()

	if worstResult > checkResultWarning {
		return chezmoi.ExitCodeError(1)
	}

	return nil
}

func (c *argsCheck) Name() string {
	return c.name
}

func (c *argsCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultOK, shellQuoteCommand(c.command, c.args)
}

func (c *binaryCheck) Name() string {
	return c.name
}

func (c *binaryCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	if c.binaryname == "" {
		return c.ifNotSet, "not set"
	}

	var pathAbsPath chezmoi.AbsPath
	switch path, err := chezmoi.LookPath(c.binaryname); {
	case errors.Is(err, exec.ErrNotFound):
		return c.ifNotExist, c.binaryname + " not found in $PATH"
	case err != nil:
		return checkResultFailed, err.Error()
	default:
		pathAbsPath, err = chezmoi.NewAbsPathFromExtPath(path, homeDirAbsPath)
		if err != nil {
			return checkResultFailed, err.Error()
		}
	}

	if c.versionArgs == nil {
		return checkResultOK, fmt.Sprintf("found %s", pathAbsPath)
	}

	cmd := exec.Command(pathAbsPath.String(), c.versionArgs...) //nolint:gosec
	output, err := chezmoilog.LogCmdCombinedOutput(slog.Default(), cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}

	versionBytes := output
	if c.versionRx != nil {
		match := c.versionRx.FindSubmatch(versionBytes)
		if len(match) != 2 {
			s := fmt.Sprintf("found %s, cannot parse version from %s", pathAbsPath, bytes.TrimSpace(versionBytes))
			return checkResultWarning, s
		}
		versionBytes = match[1]
	}
	version, err := semver.NewVersion(string(versionBytes))
	if err != nil {
		return checkResultFailed, err.Error()
	}

	if c.minVersion != nil && version.LessThan(*c.minVersion) {
		return checkResultError, fmt.Sprintf("found %s, version %s, need %s", pathAbsPath, version, c.minVersion)
	}

	return checkResultOK, fmt.Sprintf("found %s, version %s", pathAbsPath, version)
}

func (c *configFileCheck) Name() string {
	return "config-file"
}

func (c *configFileCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	filenameAbsPaths := chezmoiset.New[chezmoi.AbsPath]()
	for _, dir := range append([]string{c.bds.ConfigHome}, c.bds.ConfigDirs...) {
		configDirAbsPath, err := chezmoi.NewAbsPathFromExtPath(dir, homeDirAbsPath)
		if err != nil {
			return checkResultFailed, err.Error()
		}
		for _, extension := range chezmoi.FormatExtensions {
			filenameAbsPath := configDirAbsPath.Join(c.basename, chezmoi.NewRelPath(c.basename.String()+"."+extension))
			if _, err := system.Stat(filenameAbsPath); err == nil {
				filenameAbsPaths.Add(filenameAbsPath)
			}
		}
	}
	switch len(filenameAbsPaths) {
	case 0:
		return checkResultOK, "no config file found"
	case 1:
		filenameAbsPath := filenameAbsPaths.AnyElement()
		if filenameAbsPath != c.expected {
			return checkResultFailed, fmt.Sprintf("found %s, expected %s", filenameAbsPath, c.expected)
		}
		config, err := newConfig()
		if err != nil {
			return checkResultError, err.Error()
		}
		if err := config.decodeConfigFile(filenameAbsPath, &config.ConfigFile); err != nil {
			return checkResultError, fmt.Sprintf("%s: %v", filenameAbsPath, err)
		}
		fileInfo, err := system.Stat(filenameAbsPath)
		if err != nil {
			return checkResultError, fmt.Sprintf("%s: %v", filenameAbsPath, err)
		}
		message := fmt.Sprintf("%s, last modified %s", filenameAbsPath.String(), fileInfo.ModTime().Format(time.RFC3339))
		return checkResultOK, message
	default:
		filenameStrs := make([]string, 0, len(filenameAbsPaths))
		for filenameAbsPath := range filenameAbsPaths {
			filenameStrs = append(filenameStrs, filenameAbsPath.String())
		}
		sort.Strings(filenameStrs)
		return checkResultWarning, englishList(filenameStrs) + ": multiple config files"
	}
}

func (c *dirCheck) Name() string {
	return c.name
}

func (c *dirCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	dirEntries, err := system.ReadDir(c.dirname)
	if err != nil {
		return checkResultError, err.Error()
	}

	gitStatus := gitStatusNotAWorkingCopy
	for _, dirEntry := range dirEntries {
		if dirEntry.Name() != ".git" {
			continue
		}
		cmd := exec.Command( //nolint:gosec
			"git",
			"-C",
			c.dirname.String(),
			"status",
			"--porcelain=v2",
		)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			gitStatus = gitStatusError
			break
		}
		switch status, err := chezmoigit.ParseStatusPorcelainV2(output); {
		case err != nil:
			gitStatus = gitStatusError
		case status.Empty():
			gitStatus = gitStatusClean
		default:
			gitStatus = gitStatusDirty
		}
		break
	}
	switch gitStatus {
	case gitStatusNotAWorkingCopy:
		return checkResultOK, fmt.Sprintf("%s is a directory", c.dirname)
	case gitStatusClean:
		return checkResultOK, fmt.Sprintf("%s is a git working tree (clean)", c.dirname)
	case gitStatusDirty:
		return checkResultWarning, fmt.Sprintf("%s is a git working tree (dirty)", c.dirname)
	case gitStatusError:
		return checkResultError, fmt.Sprintf("%s is a git working tree (error)", c.dirname)
	default:
		panic(fmt.Sprintf("%s: unknown git status", gitStatus))
	}
}

func (executableCheck) Name() string {
	return "executable"
}

func (executableCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	executable, err := os.Executable()
	if err != nil {
		return checkResultError, err.Error()
	}
	executableAbsPath, err := chezmoi.NewAbsPathFromExtPath(executable, homeDirAbsPath)
	if err != nil {
		return checkResultError, err.Error()
	}
	return checkResultOK, executableAbsPath.String()
}

func (c *fileCheck) Name() string {
	return c.name
}

func (c *fileCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	if c.filename.Empty() {
		return c.ifNotSet, "not set"
	}

	switch _, err := system.ReadFile(c.filename); {
	case errors.Is(err, fs.ErrNotExist):
		return c.ifNotExist, fmt.Sprintf("%s does not exist", c.filename)
	case err != nil:
		return checkResultError, fmt.Sprintf("%s: %v", c.filename, err)
	default:
		return checkResultOK, fmt.Sprintf("%s is a file", c.filename)
	}
}

func (goVersionCheck) Name() string {
	return "go-version"
}

func (goVersionCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultOK, fmt.Sprintf("%s (%s)", runtime.Version(), runtime.Compiler)
}

func (c *latestVersionCheck) Name() string {
	return "latest-version"
}

func (c *latestVersionCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	if c.httpClientErr != nil {
		return checkResultFailed, c.httpClientErr.Error()
	}

	ctx := context.Background()

	gitHubClient := chezmoi.NewGitHubClient(ctx, c.httpClient, "github.com")
	rr, _, err := gitHubClient.Repositories.GetLatestRelease(ctx, "twpayne", "chezmoi")
	var rateLimitErr *github.RateLimitError
	var abuseRateLimitErr *github.AbuseRateLimitError
	switch {
	case err == nil:
		// Do nothing.
	case errors.As(err, &rateLimitErr):
		return checkResultFailed, "GitHub rate limit exceeded"
	case errors.As(err, &abuseRateLimitErr):
		return checkResultFailed, "GitHub abuse rate limit exceeded"
	default:
		return checkResultFailed, err.Error()
	}

	version, err := semver.NewVersion(strings.TrimPrefix(rr.GetName(), "v"))
	if err != nil {
		return checkResultError, err.Error()
	}

	checkResult := checkResultOK
	if c.version.LessThan(*version) {
		checkResult = checkResultWarning
	}
	return checkResult, "v" + version.String()
}

func (osArchCheck) Name() string {
	return "os-arch"
}

func (osArchCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	fields := []string{runtime.GOOS + "/" + runtime.GOARCH}
	if osRelease, err := chezmoi.OSRelease(system.UnderlyingFS()); err == nil {
		if name, ok := osRelease["NAME"].(string); ok {
			if version, ok := osRelease["VERSION"].(string); ok {
				fields = append(fields, "("+name+" "+version+")")
			} else {
				fields = append(fields, "("+name+")")
			}
		}
	}
	return checkResultOK, strings.Join(fields, " ")
}

func (skippedCheck) Name() string {
	return "skipped"
}

func (skippedCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	return checkResultSkipped, ""
}

func (c *suspiciousEntriesCheck) Name() string {
	return "suspicious-entries"
}

func (c *suspiciousEntriesCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	// FIXME include user-defined suffixes from age and gpg configs
	encryptedSuffixes := []string{
		defaultAgeEncryptionConfig.Suffix,
		defaultGPGEncryptionConfig.Suffix,
	}
	// FIXME check that config file templates are in root
	var suspiciousEntries []string
	walkFunc := func(absPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chezmoi.SuspiciousSourceDirEntry(absPath.Base(), fileInfo, encryptedSuffixes) {
			suspiciousEntries = append(suspiciousEntries, absPath.String())
		}
		return nil
	}
	switch err := chezmoi.WalkSourceDir(system, c.dirname, walkFunc); {
	case errors.Is(err, fs.ErrNotExist):
		return checkResultOK, fmt.Sprintf("%s: no such file or directory", c.dirname)
	case err != nil:
		return checkResultError, err.Error()
	}
	if len(suspiciousEntries) > 0 {
		return checkResultWarning, englishList(suspiciousEntries)
	}
	return checkResultOK, "no suspicious entries"
}

func (upgradeMethodCheck) Name() string {
	return "upgrade-method"
}

func (upgradeMethodCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	executable, err := os.Executable()
	if err != nil {
		return checkResultFailed, err.Error()
	}
	method, err := getUpgradeMethod(system.UnderlyingFS(), chezmoi.NewAbsPath(executable))
	if err != nil {
		return checkResultFailed, err.Error()
	}
	if method == "" {
		return checkResultSkipped, ""
	}
	return checkResultOK, method
}

func (c *versionCheck) Name() string {
	return "version"
}

func (c *versionCheck) Run(system chezmoi.System, homeDirAbsPath chezmoi.AbsPath) (checkResult, string) {
	if c.versionInfo.Version == "" || c.versionInfo.Commit == "" {
		return checkResultWarning, c.versionStr
	}
	return checkResultOK, c.versionStr
}
