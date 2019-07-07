// +build !windows

package chezmoi

import "os"

// This implementation doesn't use the extra features of PrivacyStater, but the Windows implementation needs them.
// nolint:interfacer
func IsPrivate(fs PrivacyStater, file string, umask os.FileMode) bool {
	info, err := fs.Stat(file)
	if err != nil {
		return false
	}

	return info.Mode().Perm()&^umask == 0700&^umask
}
