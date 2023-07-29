//go:build plan9

package shell

// CurrentUserShell returns the current user's shell.
func CurrentUserShell() (string, bool) {
	return DefaultShell(), false
}
