//go:build darwin
// +build darwin

package shell

import (
	"os"
	"os/exec"
	"os/user"
	"regexp"
)

var dsclUserShellRegexp = regexp.MustCompile(`\AUserShell:\s+(.*?)\s*\z`) //nolint:gochecknoglobals

// CurrentUserShell returns the current user's shell.
func CurrentUserShell() (string, bool) {
	// If the SHELL environment variable is set, use it.
	if shell, ok := os.LookupEnv("SHELL"); ok {
		return shell, true
	}

	// Try to get the current user. If we can't then fallback to the default
	// shell.
	u, err := user.Current()
	if err != nil {
		return DefaultShell(), false
	}

	// If getpwnam_r is available, use it.
	if shell, ok := cgoGetUserShell(u.Username); ok {
		return shell, true
	}

	// If dscl is available, use it.
	if output, err := exec.Command("dscl", ".", "-read", u.HomeDir, "UserShell").Output(); err == nil { //nolint:gosec
		if m := dsclUserShellRegexp.FindSubmatch(output); m != nil {
			return string(m[1]), true
		}
	}

	// Fallback to the default shell.
	return DefaultShell(), false
}
