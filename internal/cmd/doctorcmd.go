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
	"strings"
	"text/tabwriter"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/v61/github"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-xdg/v6"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoigit"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A checkResult is the result of a check.
type checkResult int

const (
	checkResultOmitted checkResult = -3 // The check was omitted.
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
	Name() string                             // Name returns the check's name.
	Run(config *Config) (checkResult, string) // Run runs the check.
}

var checkResultStr = map[checkResult]string{
	checkResultOmitted: "omitted",
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
	binaryName  string
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
	network       bool
	httpClient    *http.Client
	httpClientErr error
	version       semver.Version
}

// An osArchCheck checks that runtime.GOOS and runtime.GOARCH are supported.
type osArchCheck struct{}

// A omittedCheck is a check that is omitted.
type omittedCheck struct{}

// A suspiciousEntriesCheck checks that a source directory does not contain any
// suspicious files.
type suspiciousEntriesCheck struct {
	dirname           chezmoi.AbsPath
	encryptedSuffixes []string
}

// A upgradeMethodCheck checks the upgrade method.
type upgradeMethodCheck struct{}

// A versionCheck checks the version information.
type versionCheck struct {
	versionInfo VersionInfo
	versionStr  string
}

type doctorCmdConfig struct {
	noNetwork bool
}

func (c *Config) newDoctorCmd() *cobra.Command {
	doctorCmd := &cobra.Command{
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		Use:               "doctor",
		Short:             "Check your system for potential problems",
		Example:           example("doctor"),
		Long:              mustLongHelp("doctor"),
		RunE:              c.runDoctorCmd,
		Annotations: newAnnotations(
			doesNotRequireValidConfig,
			persistentStateModeNone,
			runsCommands,
		),
	}

	doctorCmd.Flags().BoolVar(&c.doctor.noNetwork, "no-network", c.doctor.noNetwork, "do not use network connection")

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
			network:       !c.doctor.noNetwork,
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
		},
		&dirCheck{
			name:    "source-dir",
			dirname: c.SourceDirAbsPath,
		},
		&suspiciousEntriesCheck{
			dirname: c.SourceDirAbsPath,
			encryptedSuffixes: []string{
				c.Age.Suffix,
				c.GPG.Suffix,
			},
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
			binaryName: cdCommand,
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
			binaryName: c.Diff.Command,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultWarning,
		},
		&binaryCheck{
			name:       "edit-command",
			binaryName: editCommand,
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
			binaryName:  c.Git.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^git\s+version\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "merge-command",
			binaryName: c.Merge.Command,
			ifNotSet:   checkResultWarning,
			ifNotExist: checkResultWarning,
		},
		&binaryCheck{
			name:       "shell-command",
			binaryName: shellCommand,
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
			binaryName:  c.Age.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(\d+\.\d+\.\d+\S*)`),
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
		},
		&binaryCheck{
			name:        "gpg-command",
			binaryName:  c.GPG.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(?m)^gpg\s+\(.*?\)\s+(\d+\.\d+\.\d+)`),
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
		},
		&binaryCheck{
			name:        "pinentry-command",
			binaryName:  c.PINEntry.Command,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^\S+\s+\(pinentry\)\s+(\d+\.\d+\.\d+)`),
			ifNotSet:    checkResultInfo,
			ifNotExist:  checkResultWarning,
		},
		&binaryCheck{
			name:        "1password-command",
			binaryName:  c.Onepassword.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   onepasswordVersionRx,
			minVersion:  &onepasswordMinVersion,
		},
		&binaryCheck{
			name:        "bitwarden-command",
			binaryName:  c.Bitwarden.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`(?m)^(\d+\.\d+\.\d+)$`),
		},
		&binaryCheck{
			name:        "bitwarden-secrets-command",
			binaryName:  c.BitwardenSecrets.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`Bitwarden\s+Secrets\s+CLI\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "dashlane-command",
			binaryName:  c.Dashlane.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "doppler-command",
			binaryName:  c.Doppler.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "gopass-command",
			binaryName:  c.Gopass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: gopassVersionArgs,
			versionRx:   gopassVersionRx,
			minVersion:  &gopassMinVersion,
		},
		&binaryCheck{
			name:        "keepassxc-command",
			binaryName:  c.Keepassxc.Command,
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
			binaryName:  c.Keeper.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Commander\s+Version:\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "lastpass-command",
			binaryName:  c.Lastpass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: lastpassVersionArgs,
			versionRx:   lastpassVersionRx,
			minVersion:  &lastpassMinVersion,
		},
		&binaryCheck{
			name:        "pass-command",
			binaryName:  c.Pass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "passhole-command",
			binaryName:  c.Passhole.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
			minVersion:  &passholeMinVersion,
		},
		&binaryCheck{
			name:        "rbw-command",
			binaryName:  c.RBW.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^rbw\s+(\d+\.\d+\.\d+)`),
			minVersion:  &rbwMinVersion,
		},
		&binaryCheck{
			name:        "vault-command",
			binaryName:  c.Vault.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "vlt-command",
			binaryName:  c.HCPVaultSecrets.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
			minVersion:  &vltMinVersion,
		},
		&binaryCheck{
			name:       "secret-command",
			binaryName: c.Secret.Command,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultInfo,
		},
	}

	worstResult := checkResultOK
	resultWriter := tabwriter.NewWriter(c.stdout, 3, 0, 3, ' ', 0)
	fmt.Fprint(resultWriter, "RESULT\tCHECK\tMESSAGE\n")
	for _, check := range checks {
		checkResult, message := check.Run(c)
		if checkResult == checkResultOmitted {
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

func (c *argsCheck) Run(config *Config) (checkResult, string) {
	return checkResultOK, shellQuoteCommand(c.command, c.args)
}

func (c *binaryCheck) Name() string {
	return c.name
}

func (c *binaryCheck) Run(config *Config) (checkResult, string) {
	if c.binaryName == "" {
		return c.ifNotSet, "not set"
	}

	var pathAbsPath chezmoi.AbsPath
	switch path, err := chezmoi.LookPath(c.binaryName); {
	case errors.Is(err, exec.ErrNotFound):
		return c.ifNotExist, c.binaryName + " not found in $PATH"
	case err != nil:
		return checkResultFailed, err.Error()
	default:
		pathAbsPath, err = chezmoi.NewAbsPathFromExtPath(path, config.homeDirAbsPath)
		if err != nil {
			return checkResultFailed, err.Error()
		}
	}

	if c.versionArgs == nil {
		return checkResultOK, fmt.Sprintf("found %s", pathAbsPath)
	}

	cmd := exec.Command(pathAbsPath.String(), c.versionArgs...)
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

func (c *configFileCheck) Run(config *Config) (checkResult, string) {
	configFileAbsPath, err := config.getConfigFileAbsPath()
	if err != nil {
		return checkResultError, err.Error()
	}

	fileInfo, err := config.baseSystem.Stat(configFileAbsPath)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return checkResultInfo, fmt.Sprintf("%s: not found", configFileAbsPath)
	case err != nil:
		return checkResultError, fmt.Sprintf("%s: %s", configFileAbsPath, err)
	case fileInfo.Mode().Type() != 0:
		return checkResultError, fmt.Sprintf(
			"found %s, which is a %s",
			configFileAbsPath,
			chezmoi.FileModeTypeNames[fileInfo.Mode().Type()],
		)
	}

	tmpConfig, err := newConfig()
	if err != nil {
		return checkResultError, err.Error()
	}
	if err := config.decodeConfigFile(configFileAbsPath, &tmpConfig.ConfigFile); err != nil {
		return checkResultError, fmt.Sprintf("%s: %v", configFileAbsPath, err)
	}

	message := fmt.Sprintf("found %s, last modified %s", configFileAbsPath, fileInfo.ModTime().Format(time.RFC3339))
	return checkResultOK, message
}

func (c *dirCheck) Name() string {
	return c.name
}

func (c *dirCheck) Run(config *Config) (checkResult, string) {
	dirEntries, err := config.baseSystem.ReadDir(c.dirname)
	if err != nil {
		return checkResultError, err.Error()
	}

	gitStatus := gitStatusNotAWorkingCopy
	for _, dirEntry := range dirEntries {
		if dirEntry.Name() != ".git" {
			continue
		}
		cmd := exec.Command("git", "-C", c.dirname.String(), "status", "--porcelain=v2")
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			gitStatus = gitStatusError
			break
		}
		switch status, err := chezmoigit.ParseStatusPorcelainV2(output); {
		case err != nil:
			gitStatus = gitStatusError
		case status.IsEmpty():
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

func (executableCheck) Run(config *Config) (checkResult, string) {
	executable, err := os.Executable()
	if err != nil {
		return checkResultError, err.Error()
	}
	executableAbsPath, err := chezmoi.NewAbsPathFromExtPath(executable, config.homeDirAbsPath)
	if err != nil {
		return checkResultError, err.Error()
	}
	return checkResultOK, executableAbsPath.String()
}

func (c *fileCheck) Name() string {
	return c.name
}

func (c *fileCheck) Run(config *Config) (checkResult, string) {
	if c.filename.IsEmpty() {
		return c.ifNotSet, "not set"
	}

	switch _, err := config.baseSystem.ReadFile(c.filename); {
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

func (goVersionCheck) Run(config *Config) (checkResult, string) {
	return checkResultOK, fmt.Sprintf("%s (%s)", runtime.Version(), runtime.Compiler)
}

func (c *latestVersionCheck) Name() string {
	return "latest-version"
}

func (c *latestVersionCheck) Run(config *Config) (checkResult, string) {
	switch {
	case !c.network:
		return checkResultSkipped, "no network"
	case c.httpClientErr != nil:
		return checkResultFailed, c.httpClientErr.Error()
	}

	ctx := context.Background()

	gitHubClient := chezmoi.NewGitHubClient(ctx, c.httpClient)
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

func (osArchCheck) Run(config *Config) (checkResult, string) {
	fields := []string{runtime.GOOS + "/" + runtime.GOARCH}
	if osRelease, err := chezmoi.OSRelease(config.baseSystem.UnderlyingFS()); err == nil {
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

func (omittedCheck) Name() string {
	return "omitted"
}

func (omittedCheck) Run(config *Config) (checkResult, string) {
	return checkResultOmitted, ""
}

func (c *suspiciousEntriesCheck) Name() string {
	return "suspicious-entries"
}

func (c *suspiciousEntriesCheck) Run(config *Config) (checkResult, string) {
	// FIXME check that config file templates are in root
	var suspiciousEntries []string
	walkFunc := func(absPath chezmoi.AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chezmoi.IsSuspiciousSourceDirEntry(absPath.Base(), fileInfo, c.encryptedSuffixes) {
			suspiciousEntries = append(suspiciousEntries, absPath.String())
		}
		return nil
	}
	switch err := chezmoi.WalkSourceDir(config.baseSystem, c.dirname, walkFunc); {
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

func (upgradeMethodCheck) Run(config *Config) (checkResult, string) {
	executable, err := os.Executable()
	if err != nil {
		return checkResultFailed, err.Error()
	}
	method, err := getUpgradeMethod(config.baseSystem.UnderlyingFS(), chezmoi.NewAbsPath(executable))
	if err != nil {
		return checkResultFailed, err.Error()
	}
	if method == "" {
		return checkResultOmitted, ""
	}
	return checkResultOK, method
}

func (c *versionCheck) Name() string {
	return "version"
}

func (c *versionCheck) Run(config *Config) (checkResult, string) {
	if c.versionInfo.Version == "" || c.versionInfo.Commit == "" {
		return checkResultWarning, c.versionStr
	}
	return checkResultOK, c.versionStr
}
