package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	vfs "github.com/twpayne/go-vfs"
)

// doctorCmd represents the doctor command
var doctorCommand = &cobra.Command{
	Args:  cobra.NoArgs,
	Use:   "doctor",
	Short: "Check your system for potential problems",
	RunE:  makeRunE(config.runDoctorCommandE),
}

const (
	okPrefix      = "     ok: "
	warningPrefix = "warning: "
	errorPrefix   = "  ERROR: "
)

type vcsInfo struct {
	versionArgs   []string
	versionRegexp *regexp.Regexp
}

var (
	vcsInfos = map[string]vcsInfo{
		"bzr": vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^Bazaar (bzr) (\d+\.\d+\.\d+)`),
		},
		"cvs": vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^Concurrent Versions System \(CVS\) (\d+\.\d+\.\d+)`),
		},
		"git": vcsInfo{
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^git version (\d+\.\d+\.\d+)`),
		},
		"hg": vcsInfo{
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^Mercurial Distributed SCM \(version (\d+\.\d+\.\d+\))`),
		},
		"svn": vcsInfo{
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^svn, version (\d+\.\d+\.\d+)`),
		},
	}
)

type doctorCheck interface {
	Check() (bool, error)
	Enabled() bool
	MustSucceed() bool
	Result() string
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
	dontWantPerm os.FileMode
	info         os.FileInfo
}

type doctorFileCheck struct {
	name        string
	path        string
	mustSucceed bool
	info        os.FileInfo
}

func init() {
	rootCommand.AddCommand(doctorCommand)
}

func (c *Config) runDoctorCommandE(fs vfs.FS, args []string) error {
	allOK := true
	for _, dc := range []doctorCheck{
		&doctorDirectoryCheck{
			name:         "source directory",
			path:         c.SourceDir,
			dontWantPerm: 077,
		},
		&doctorDirectoryCheck{
			name: "destination directory",
			path: c.DestDir,
		},
		&doctorFileCheck{
			name: "configuration file",
			path: c.configFile,
		},
		&doctorBinaryCheck{
			name:        "editor",
			binaryName:  c.getEditor(),
			mustSucceed: true,
		},
		&doctorBinaryCheck{
			name:          "source VCS command",
			binaryName:    c.SourceVCSCommand,
			versionArgs:   vcsInfos[c.SourceVCSCommand].versionArgs,
			versionRegexp: vcsInfos[c.SourceVCSCommand].versionRegexp,
		},
		&doctorBinaryCheck{
			name:          "1Password CLI",
			binaryName:    c.OnePassword.Op,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "Bitwarden CLI",
			binaryName:    c.Bitwarden.Bw,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "LastPass CLI",
			binaryName:    c.LastPass.Lpass,
			versionArgs:   []string{"--version"},
			versionRegexp: regexp.MustCompile(`^LastPass CLI v(\d+\.\d+\.\d+)`),
			// chezmoi uses lpass show --json which was added in
			// https://github.com/lastpass/lastpass-cli/commit/e5a22e2eeef31ab6c54595616e0f57ca0a1c162d
			// and the first tag containing that commit is v1.3.0~6.
			minVersion: &semver.Version{Major: 1, Minor: 3, Patch: 0},
		},
		&doctorBinaryCheck{
			name:          "pass CLI",
			binaryName:    c.Pass.Pass,
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`(?m)=\s*v(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:          "Vault CLI",
			binaryName:    c.Vault.Vault,
			versionArgs:   []string{"version"},
			versionRegexp: regexp.MustCompile(`^Vault\s+v(\d+\.\d+\.\d+)`),
		},
		&doctorBinaryCheck{
			name:       "generic secret CLI",
			binaryName: c.GenericSecret.Command,
		},
	} {
		if !dc.Enabled() {
			continue
		}
		ok, err := dc.Check()
		var prefix string
		switch {
		case ok:
			prefix = okPrefix
		case !ok && !dc.MustSucceed():
			prefix = warningPrefix
		default:
			prefix = errorPrefix
		}
		if _, err := fmt.Printf("%s%s\n", prefix, dc.Result()); err != nil {
			return err
		}
		if err != nil {
			if _, err := fmt.Println(err); err != nil {
				return err
			}
		}
	}
	if !allOK {
		os.Exit(1)
	}
	return nil
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
		version, err := semver.Parse(string(m[1]))
		if err != nil {
			return false, err
		}
		c.version = &version
		if c.minVersion != nil && c.version.LT(*c.minVersion) {
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
		if c.minVersion != nil && c.version.LT(*c.minVersion) {
			s += ", want version >=" + c.minVersion.String()
		}
	}
	s += ")"
	return s
}

func (c *doctorDirectoryCheck) Check() (bool, error) {
	var err error
	c.info, err = os.Stat(c.path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
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
	return fmt.Sprintf("%s (%s, perm %03o)", c.path, c.name, c.info.Mode()&os.ModePerm)
}

func (c *doctorFileCheck) Check() (bool, error) {
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
	return fmt.Sprintf("%s (%s)", c.path, c.name)
}
