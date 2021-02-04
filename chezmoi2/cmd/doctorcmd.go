package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
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

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoi"
	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoilog"
)

// A checkResult is the result of a check.
type checkResult int

const (
	checkSkipped checkResult = -1 // The check was skipped.
	checkOK      checkResult = 0  // The check completed and did not find any problems.
	checkWarning checkResult = 1  // The check completed and found something that might indicate a problem.
	checkError   checkResult = 2  // The check completed and found a definite problem.
	checkFailed  checkResult = 3  // The check could not be completed.
)

// A check is an individual check.
type check interface {
	Name() string               // Name returns the check's name.
	Run() (checkResult, string) // Run runs the check.
}

var checkResultStr = map[checkResult]string{
	checkSkipped: "skipped",
	checkOK:      "ok",
	checkWarning: "warning",
	checkError:   "error",
	checkFailed:  "failed",
}

// A binaryCheck checks that a binary called name is installed and optionally at
// least version minVersion.
type binaryCheck struct {
	name        string
	binaryname  string
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

// A fileCheck checks that a file exists.
type fileCheck struct {
	name       string
	filename   string
	ifNotExist checkResult
}

// An osArchCheck checks that runtime.GOOS and runtime.GOARCH are supported.
type osArchCheck struct{}

// A suspiciousEntriesCheck checks that a source directory does not contain any
// suspicious files.
type suspiciousEntriesCheck struct {
	dirname string
}

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
	shell, _ := shell.CurrentUserShell()
	editor, _ := c.editor()
	checks := []check{
		&versionCheck{
			versionInfo: c.versionInfo,
			versionStr:  c.versionStr,
		},
		&osArchCheck{},
		&fileCheck{
			name:       "config-file",
			filename:   string(c.configFileAbsPath),
			ifNotExist: checkWarning,
		},
		&dirCheck{
			name:    "source-dir",
			dirname: string(c.sourceDirAbsPath),
		},
		&suspiciousEntriesCheck{
			dirname: string(c.sourceDirAbsPath),
		},
		&dirCheck{
			name:    "dest-dir",
			dirname: string(c.destDirAbsPath),
		},
		&binaryCheck{
			name:       "shell",
			binaryname: shell,
		},
		&binaryCheck{
			name:       "editor",
			binaryname: editor,
		},
		&binaryCheck{
			name:        "git-cli",
			binaryname:  c.Git.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^git\s+version\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "merge-cli",
			binaryname: c.Merge.Command,
			ifNotExist: checkWarning,
		},
		&binaryCheck{
			name:        "gnupg-cli",
			binaryname:  "gpg",
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^gpg\s+\(.*?\)\s+(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "1password-cli",
			binaryname:  c.Onepassword.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "bitwarden-cli",
			binaryname:  c.Bitwarden.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "gopass-cli",
			binaryname:  c.Gopass.Command,
			ifNotExist:  checkWarning,
			versionArgs: gopassVersionArgs,
			versionRx:   gopassVersionRx,
			minVersion:  &gopassMinVersion,
		},
		&binaryCheck{
			name:        "keepassxc-cli",
			binaryname:  c.Keepassxc.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"--version"},
			versionRx:   regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&fileCheck{
			name:       "keepassxc-db",
			filename:   c.Keepassxc.Database,
			ifNotExist: checkWarning,
		},
		&binaryCheck{
			name:        "lastpass-cli",
			binaryname:  c.Lastpass.Command,
			ifNotExist:  checkWarning,
			versionArgs: lastpassVersionArgs,
			versionRx:   lastpassVersionRx,
			minVersion:  &lastpassMinVersion,
		},
		&binaryCheck{
			name:        "pass-cli",
			binaryname:  c.Pass.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:        "vault-cli",
			binaryname:  c.Vault.Command,
			ifNotExist:  checkWarning,
			versionArgs: []string{"version"},
			versionRx:   regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&binaryCheck{
			name:       "secret-cli",
			binaryname: c.Secret.Command,
			ifNotExist: checkWarning,
		},
	}

	//nolint:ifshort
	worstResult := checkOK
	resultWriter := tabwriter.NewWriter(c.stdout, 3, 0, 3, ' ', 0)
	fmt.Fprint(resultWriter, "RESULT\tCHECK\tMESSAGE\n")
	for _, check := range checks {
		checkResult, message := check.Run()
		fmt.Fprintf(resultWriter, "%s\t%s\t%s\n", checkResultStr[checkResult], check.Name(), message)
		if checkResult > worstResult {
			worstResult = checkResult
		}
	}
	resultWriter.Flush()

	if worstResult > checkWarning {
		return ErrExitCode(1)
	}

	return nil
}

func (c *binaryCheck) Name() string {
	return c.name
}

func (c *binaryCheck) Run() (checkResult, string) {
	if c.binaryname == "" {
		return checkWarning, "not set"
	}

	path, err := exec.LookPath(c.binaryname)
	switch {
	case errors.Is(err, exec.ErrNotFound):
		return c.ifNotExist, fmt.Sprintf("%s not found in $PATH", c.binaryname)
	case err != nil:
		return checkFailed, err.Error()
	}

	if c.versionArgs == nil {
		return checkOK, fmt.Sprintf("found %s", path)
	}

	cmd := exec.Command(path, c.versionArgs...)
	output, err := chezmoilog.LogCmdCombinedOutput(log.Logger, cmd)
	if err != nil {
		return checkFailed, err.Error()
	}

	versionBytes := output
	if c.versionRx != nil {
		match := c.versionRx.FindSubmatch(versionBytes)
		if len(match) != 2 {
			return checkFailed, fmt.Sprintf("found %s, could not parse version from %s", path, versionBytes)
		}
		versionBytes = match[1]
	}
	version, err := semver.NewVersion(string(versionBytes))
	if err != nil {
		return checkFailed, err.Error()
	}

	if c.minVersion != nil && version.LessThan(*c.minVersion) {
		return checkError, fmt.Sprintf("found %s, version %s, need %s", path, version, c.minVersion)
	}

	return checkOK, fmt.Sprintf("found %s, version %s", path, version)
}

func (c *dirCheck) Name() string {
	return c.name
}

func (c *dirCheck) Run() (checkResult, string) {
	if _, err := ioutil.ReadDir(c.dirname); err != nil {
		return checkError, fmt.Sprintf("%s: %v", c.dirname, err)
	}
	return checkOK, fmt.Sprintf("%s is a directory", c.dirname)
}

func (c *fileCheck) Name() string {
	return c.name
}

func (c *fileCheck) Run() (checkResult, string) {
	if c.filename == "" {
		return checkWarning, "not set"
	}

	_, err := ioutil.ReadFile(c.filename)
	switch {
	case os.IsNotExist(err):
		return c.ifNotExist, fmt.Sprintf("%s does not exist", c.filename)
	case err != nil:
		return checkError, fmt.Sprintf("%s: %v", c.filename, err)
	default:
		return checkOK, fmt.Sprintf("%s is a file", c.filename)
	}
}

func (osArchCheck) Name() string {
	return "os-arch"
}

func (osArchCheck) Run() (checkResult, string) {
	return checkOK, fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

func (c *suspiciousEntriesCheck) Name() string {
	return "suspicious-entries"
}

func (c *suspiciousEntriesCheck) Run() (checkResult, string) {
	// FIXME check that config file templates are in root
	var suspiciousEntries []string
	if err := filepath.Walk(c.dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chezmoi.SuspiciousSourceDirEntry(filepath.Base(path), info) {
			suspiciousEntries = append(suspiciousEntries, path)
		}
		return nil
	}); err != nil {
		return checkError, err.Error()
	}
	if len(suspiciousEntries) > 0 {
		return checkWarning, fmt.Sprintf("found suspicious entries in %s: %s", c.dirname, strings.Join(suspiciousEntries, ", "))
	}

	return checkOK, fmt.Sprintf("no suspicious entries found in %s", c.dirname)
}

func (c *versionCheck) Name() string {
	return "version"
}

func (c *versionCheck) Run() (checkResult, string) {
	if c.versionInfo.Version == "" ||
		c.versionInfo.Commit == "" ||
		c.versionInfo.Date == "" ||
		c.versionInfo.BuiltBy == "" {
		return checkWarning, c.versionStr
	}
	return checkOK, c.versionStr
}
