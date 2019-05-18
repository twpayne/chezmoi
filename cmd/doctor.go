package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

var doctorCmd = &cobra.Command{
	Args:    cobra.NoArgs,
	Use:     "doctor",
	Short:   "Check your system for potential problems",
	Example: getExample("doctor"),
	Long:    mustGetLongHelp("doctor"),
	RunE:    makeRunE(config.runDoctorCmd),
}

const (
	okPrefix      = "ok"
	warningPrefix = "warning"
	errorPrefix   = "ERROR"
)

type doctorCheck interface {
	Check() (bool, error)
	Enabled() bool
	MustSucceed() bool
	Result() string
	Skip() bool
}

type doctorCheckResult struct {
	ok     bool
	prefix string
	result string
}

type doctorBinaryCheck struct {
	name          string
	binaryName    string
	path          string
	minVersion    *semver.Version
	mustSucceed   bool
	versionArgs   []string
	versionRegexp *regexp.Regexp
	version       *semver.Version
}

type doctorDirectoryCheck struct {
	name         string
	path         string
	err          error
	dontWantPerm os.FileMode
	info         os.FileInfo
}

type doctorFileCheck struct {
	name        string
	path        string
	canSkip     bool
	mustSucceed bool
	info        os.FileInfo
}

type doctorSuspiciousFilesCheck struct {
	path      string
	filenames map[string]bool
	found     []string
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func (c *Config) runDoctorCmd(fs vfs.FS, args []string) error {
	var vcsCommandCheck doctorCheck
	if vcsInfo, err := c.getVCSInfo(); err == nil {
		vcsCommandCheck = &doctorBinaryCheck{
			name:          "source VCS command",
			binaryName:    c.SourceVCS.Command,
			versionArgs:   vcsInfo.versionArgs,
			versionRegexp: vcsInfo.versionRegexp,
		}
	} else {
		c.warn(fmt.Sprintf("%s: unsupported VCS command", c.SourceVCS.Command))
		vcsCommandCheck = &doctorBinaryCheck{
			name:       "source VCS command",
			binaryName: c.SourceVCS.Command,
		}
	}

	allOK := true
	for _, dc := range []doctorCheck{
		&doctorDirectoryCheck{
			name:         "source directory",
			path:         c.SourceDir,
			dontWantPerm: 077,
		},
		&doctorSuspiciousFilesCheck{
			path: c.SourceDir,
			filenames: map[string]bool{
				".chezmoignore": true,
			},
		},
		&doctorDirectoryCheck{
			name: "destination directory",
			path: c.DestDir,
		},
		&doctorFileCheck{
			name: "configuration file",
			path: c.configFile,
		},
		&doctorFileCheck{
			name:    "KeePassXC database",
			path:    c.KeePassXC.Database,
			canSkip: true,
		},
		&doctorBinaryCheck{
			name:        "editor",
			binaryName:  c.getEditor(),
			mustSucceed: true,
		},
		&doctorBinaryCheck{
			name:       "merge command",
			binaryName: c.Merge.Command,
		},
		vcsCommandCheck,
		&doctorBinaryCheck{
			name:          "GnuPG",
			binaryName:    "gpg",
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^gpg \(GnuPG\) (\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "1Password CLI",
			binaryName:    c.Onepassword.Command,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "Bitwarden CLI",
			binaryName:    c.Bitwarden.Command,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "KeePassXC CLI",
			binaryName:    c.KeePassXC.Command,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "LastPass CLI",
			binaryName:    c.Lastpass.Command,
			versionArgs:   lastpassVersionArgs,
			versionRegexp: lastpassVersionRegexp,
			minVersion:    &lastpassMinVersion,
		},
		&doctorBinaryCheck{
			name:          "pass CLI",
			binaryName:    c.Pass.Command,
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "Vault CLI",
			binaryName:    c.Vault.Command,
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:       "generic secret CLI",
			binaryName: c.GenericSecret.Command,
		},
	} {
		if dc.Skip() {
			continue
		}
		dcr := runDoctorCheck(dc)
		if !dcr.ok {
			allOK = false
		}
		if dcr.result != "" {
			fmt.Printf("%7s: %s\n", dcr.prefix, dcr.result)
		}
	}
	if !allOK {
		os.Exit(1)
	}
	return nil
}

func runDoctorCheck(dc doctorCheck) doctorCheckResult {
	if !dc.Enabled() {
		return doctorCheckResult{ok: true}
	}
	ok, err := dc.Check()
	if err != nil {
		return doctorCheckResult{result: err.Error()}
	}
	var prefix string
	switch {
	case ok:
		prefix = okPrefix
	case !ok && !dc.MustSucceed():
		prefix = warningPrefix
	default:
		prefix = errorPrefix
	}
	return doctorCheckResult{
		ok:     ok,
		prefix: prefix,
		result: dc.Result(),
	}
}

func (c *doctorBinaryCheck) Check() (bool, error) {
	var err error
	c.path, err = exec.LookPath(c.binaryName)
	if err != nil {
		return false, nil
	}

	if c.versionRegexp != nil {
		output, err := exec.Command(c.path, c.versionArgs...).CombinedOutput()
		if err != nil {
			return false, err
		}
		m := c.versionRegexp.FindSubmatch(output)
		if m == nil {
			return false, fmt.Errorf("%s: could not extract version from %q", c.path, output)
		}
		version, err := semver.NewVersion(string(m[1]))
		if err != nil {
			return false, err
		}
		c.version = version
		if c.minVersion != nil && c.version.LessThan(*c.minVersion) {
			return false, nil
		}
	}

	return true, nil
}

func (c *doctorBinaryCheck) Enabled() bool {
	return c.binaryName != ""
}

func (c *doctorBinaryCheck) MustSucceed() bool {
	return c.mustSucceed
}

func (c *doctorBinaryCheck) Result() string {
	if c.path == "" {
		return fmt.Sprintf("%s (%s, not found)", c.binaryName, c.name)
	}
	s := fmt.Sprintf("%s (%s", c.path, c.name)
	if c.version != nil {
		s += ", version " + c.version.String()
		if c.minVersion != nil && c.version.LessThan(*c.minVersion) {
			s += ", want version >=" + c.minVersion.String()
		}
	}
	s += ")"
	return s
}

func (c *doctorBinaryCheck) Skip() bool {
	return false
}

func (c *doctorDirectoryCheck) Check() (bool, error) {
	c.info, c.err = os.Stat(c.path)
	if c.err != nil && os.IsNotExist(c.err) {
		return false, nil
	} else if c.err != nil {
		return false, c.err
	}
	if c.info.Mode()&os.ModePerm&c.dontWantPerm != 0 {
		return false, nil
	}
	return true, nil
}

func (c *doctorDirectoryCheck) Enabled() bool {
	return true
}

func (c *doctorDirectoryCheck) MustSucceed() bool {
	return true
}

func (c *doctorDirectoryCheck) Result() string {
	switch {
	case os.IsNotExist(c.err):
		return fmt.Sprintf("%s: (%s, not found)", c.path, c.name)
	case c.err != nil:
		return fmt.Sprintf("%s: (%s, %v)", c.path, c.name, c.err)
	default:
		return fmt.Sprintf("%s (%s, perm %03o)", c.path, c.name, c.info.Mode()&os.ModePerm)
	}
}

func (c *doctorDirectoryCheck) Skip() bool {
	return false
}

func (c *doctorFileCheck) Check() (bool, error) {
	if c.path == "" {
		return false, nil
	}
	var err error
	c.info, err = os.Stat(c.path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (c *doctorFileCheck) Enabled() bool {
	return true
}

func (c *doctorFileCheck) MustSucceed() bool {
	return c.mustSucceed
}

func (c *doctorFileCheck) Result() string {
	if c.path == "" {
		return fmt.Sprintf("not set (%s)", c.name)
	}
	return fmt.Sprintf("%s (%s)", c.path, c.name)
}

func (c *doctorFileCheck) Skip() bool {
	return c.canSkip && c.path == ""
}

func (c *doctorSuspiciousFilesCheck) Check() (bool, error) {
	if err := filepath.Walk(c.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if c.filenames[info.Name()] {
			c.found = append(c.found, path)
		}
		return nil
	}); err != nil {
		return false, err
	}
	return len(c.found) == 0, nil
}

func (c *doctorSuspiciousFilesCheck) Enabled() bool {
	return len(c.filenames) > 0
}

func (c *doctorSuspiciousFilesCheck) MustSucceed() bool {
	return false
}

func (c *doctorSuspiciousFilesCheck) Result() string {
	if len(c.found) == 0 {
		return ""
	}
	return fmt.Sprintf("%s (suspicious filenames)", strings.Join(c.found, ", "))
}

func (c *doctorSuspiciousFilesCheck) Skip() bool {
	return false
}
