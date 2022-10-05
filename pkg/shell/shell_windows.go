package shell

import (
	"os"
)

// CurrentUserShell returns the current user's shell.
func CurrentUserShell() (string, bool) {
	// If the SHELL environment variable is set, use it.
	if shell, ok := os.LookupEnv("SHELL"); ok {
		return shell, true
	}

	// If the ComSpec environment variable is set, use it.
	if comSpec, ok := os.LookupEnv("ComSpec"); ok {
		return comSpec, true
	}

	// Fallback to the default shell.
	return DefaultShell(), false
}
