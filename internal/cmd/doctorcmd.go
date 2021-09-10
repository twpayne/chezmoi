package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/coreos/go-semver/semver"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-shell"
	"github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoi"
	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A checkResult is the result of a check.
type checkResult int

const (
	checkResultSkipped checkResult = -1 // The check was skipped.
	checkResultOK      checkResult = 0  // The check completed and did not find any problems.
	checkResultInfo    checkResult = 1  // The check completed and found something interesting, but not a problem.
	checkResultWarning checkResult = 2  // The check completed and found something that might indicate a problem.
	checkResultError   checkResult = 3  // The check completed and found a definite problem.
	checkResultFailed  checkResult = 4  // The check could not be completed.
)

// A check is an individual check.
type check interface {
	Name() string               // Name returns the check's name.
	Run() (checkResult, string) // Run runs the check.
}

var checkResultStr = map[checkResult]string{
	checkResultSkipped: "skipped",
	checkResultOK:      "ok",
	checkResultInfo:    "info",
	checkResultWarning: "warning",
	checkResultError:   "error",
	checkResultFailed:  "failed",
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

// A dirCheck checks that a directory exists.
type dirCheck struct {
	name    string
	dirname string
}

// An executableCheck checks the executable.
type executableCheck struct{}

// A fileCheck checks that a file exists.
type fileCheck struct {
	name       string
	filename   string
	ifNotSet   checkResult
	ifNotExist checkResult
}

// An osArchCheck checks that runtime.GOOS and runtime.GOARCH are supported.
type osArchCheck struct{}

// A suspiciousEntriesCheck checks that a source directory does not contain any
// suspicious files.
type suspiciousEntriesCheck struct {
	dirname string
}

// A umaskCheck checks the umask.
type umaskCheck struct{}

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
		Annotations: map[string]string{
			doesNotRequireValidConfig: "true",
			runsCommands:              "true",
		},
	}

	return doctorCmd
}

func (c *Config) runDoctorCmd(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	shell, _ := shell.CurrentUserShell()
	editor, _ := c.editor()
	checks := []check{
		&versionCheck{
			versionInfo: c.versionInfo,
			versionStr:  c.versionStr,
		},
		&osArchCheck{},
		&executableCheck{},
		&fileCheck{
			name:       "config-file",
			filename:   string(c.configFileAbsPath),
			ifNotExist: checkResultInfo,
		},
		&dirCheck{
			name:    "source-dir",
			dirname: string(c.SourceDirAbsPath),
		},
		&suspiciousEntriesCheck{
			dirname: string(c.SourceDirAbsPath),
		},
		&dirCheck{
			name:    "dest-dir",
			dirname: string(c.DestDirAbsPath),
		},
		&binaryCheck{
			name:       "shell",
			binaryname: shell,
			ifNotSet:   checkResultError,
		},
		&binaryCheck{
			name:       "editor",
			binaryname: editor,
			ifNotSet:   checkResultWarning,
		},
		&umaskCheck{},
		&binaryCheck{
			name:        "git-cli",
			binaryname:  c.Git.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^git\s+version\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "merge-cli",
			binaryname: c.Merge.Command,
			ifNotSet:   checkResultWarning,
			ifNotExist: checkResultWarning,
		},
		&binaryCheck{
			name:        "age-cli",
			binaryname:  "age",
			versionArgs: []string{"-version"},
			versionRx:   regexp.MustCompile(`v(\d+\.\d+\.\d+\S*)`),
			ifNotSet:    checkResultWarning,
		},
		&binaryCheck{
			name:        "gnupg-cli",
			binaryname:  "gpg",
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^gpg\s+\(.*?\)\s+(\d+\.\d+\.\d+)`),
			ifNotSet:    checkResultWarning,
		},
		&binaryCheck{
			name:        "1password-cli",
			binaryname:  c.Onepassword.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "bitwarden-cli",
			binaryname:  c.Bitwarden.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "gopass-cli",
			binaryname:  c.Gopass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: gopassVersionArgs,
			versionRx:   gopassVersionRx,
			minVersion:  &gopassMinVersion,
		},
		&binaryCheck{
			name:        "keepassxc-cli",
			binaryname:  c.Keepassxc.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&fileCheck{
			name:       "keepassxc-db",
			filename:   c.Keepassxc.Database,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultInfo,
		},
		&binaryCheck{
			name:        "lastpass-cli",
			binaryname:  c.Lastpass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: lastpassVersionArgs,
			versionRx:   lastpassVersionRx,
			minVersion:  &lastpassMinVersion,
		},
		&binaryCheck{
			name:        "pass-cli",
			binaryname:  c.Pass.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "vault-cli",
			binaryname:  c.Vault.Command,
			ifNotSet:    checkResultWarning,
			ifNotExist:  checkResultInfo,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "secret-cli",
			binaryname: c.Secret.Command,
			ifNotSet:   checkResultInfo,
			ifNotExist: checkResultInfo,
		},
	}

	worstResult := checkResultOK
	resultWriter := tabwriter.NewWriter(c.stdout, 3, 0, 3, ' ', 0)
	fmt.Fprint(resultWriter, "RESULT\tCHECK\tMESSAGE\n")
	for _, check := range checks {
		checkResult, message := check.Run()
		// Conceal the user's actual home directory in the message as the
		// output of chezmoi doctor is often posted publicly and would otherwise
		// reveal the user's username.
		message = strings.ReplaceAll(message, home, "~")
		fmt.Fprintf(resultWriter, "%s\t%s\t%s\n", checkResultStr[checkResult], check.Name(), message)
		if checkResult > worstResult {
			worstResult = checkResult
		}
	}
	resultWriter.Flush()

	if worstResult > checkResultWarning {
		return ExitCodeError(1)
	}

	return nil
}

func (c *binaryCheck) Name() string {
	return c.name
}

func (c *binaryCheck) Run() (checkResult, string) {
	if c.binaryname == "" {
		return c.ifNotSet, "not set"
	}

	path, err := exec.LookPath(c.binaryname)
	switch {
	case errors.Is(err, exec.ErrNotFound):
		return c.ifNotExist, fmt.Sprintf("%s not found in $PATH", c.binaryname)
	case err != nil:
		return checkResultFailed, err.Error()
	}

	if c.versionArgs == nil {
		return checkResultOK, fmt.Sprintf("found %s", path)
	}

	cmd := exec.Command(path, c.versionArgs...)
	output, err := chezmoilog.LogCmdCombinedOutput(log.Logger, cmd)
	if err != nil {
		return checkResultFailed, err.Error()
	}

	versionBytes := output
	if c.versionRx != nil {
		match := c.versionRx.FindSubmatch(versionBytes)
		if len(match) != 2 {
			return checkResultFailed, fmt.Sprintf("found %s, could not parse version from %s", path, versionBytes)
		}
		versionBytes = match[1]
	}
	version, err := semver.NewVersion(string(versionBytes))
	if err != nil {
		return checkResultFailed, err.Error()
	}

	if c.minVersion != nil && version.LessThan(*c.minVersion) {
		return checkResultError, fmt.Sprintf("found %s, version %s, need %s", path, version, c.minVersion)
	}

	return checkResultOK, fmt.Sprintf("found %s, version %s", path, version)
}

func (c *dirCheck) Name() string {
	return c.name
}

func (c *dirCheck) Run() (checkResult, string) {
	if _, err := os.ReadDir(c.dirname); err != nil {
		return checkResultError, fmt.Sprintf("%s: %v", c.dirname, err)
	}
	return checkResultOK, fmt.Sprintf("%s is a directory", c.dirname)
}

func (c *executableCheck) Name() string {
	return "executable"
}

func (c *executableCheck) Run() (checkResult, string) {
	executable, err := os.Executable()
	if err != nil {
		return checkResultError, err.Error()
	}
	return checkResultOK, executable
}

func (c *fileCheck) Name() string {
	return c.name
}

func (c *fileCheck) Run() (checkResult, string) {
	if c.filename == "" {
		return c.ifNotSet, "not set"
	}

	_, err := os.ReadFile(c.filename)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return c.ifNotExist, fmt.Sprintf("%s does not exist", c.filename)
	case err != nil:
		return checkResultError, fmt.Sprintf("%s: %v", c.filename, err)
	default:
		return checkResultOK, fmt.Sprintf("%s is a file", c.filename)
	}
}

func (osArchCheck) Name() string {
	return "os-arch"
}

func (osArchCheck) Run() (checkResult, string) {
	fields := []string{runtime.GOOS + "/" + runtime.GOARCH}
	if osRelease, err := chezmoi.OSRelease(vfs.OSFS); err == nil {
		if name, ok := osRelease["NAME"].(string); ok {
			if version, ok := osRelease["VERSION"].(string); ok {
				fields = append(fields, "("+name+"/"+version+")")
			}
		}
	}
	return checkResultOK, strings.Join(fields, " ")
}

func (c *suspiciousEntriesCheck) Name() string {
	return "suspicious-entries"
}

func (c *suspiciousEntriesCheck) Run() (checkResult, string) {
	// FIXME check that config file templates are in root
	var suspiciousEntries []string
	if err := filepath.Walk(c.dirname, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chezmoi.SuspiciousSourceDirEntry(filepath.Base(path), info) {
			relPath, err := filepath.Rel(c.dirname, path)
			if err != nil {
				return err
			}
			suspiciousEntries = append(suspiciousEntries, relPath)
		}
		return nil
	}); err != nil {
		return checkResultError, err.Error()
	}
	if len(suspiciousEntries) > 0 {
		return checkResultWarning, fmt.Sprintf("%s: %s", c.dirname, englishList(suspiciousEntries))
	}
	return checkResultOK, fmt.Sprintf("%s: no suspicious entries", c.dirname)
}

func (c *umaskCheck) Name() string {
	return "umask"
}

func (c *versionCheck) Name() string {
	return "version"
}

func (c *versionCheck) Run() (checkResult, string) {
	if c.versionInfo.Version == "" ||
		c.versionInfo.Commit == "" ||
		c.versionInfo.Date == "" ||
		c.versionInfo.BuiltBy == "" {
		return checkResultWarning, c.versionStr
	}
	return checkResultOK, c.versionStr
}
