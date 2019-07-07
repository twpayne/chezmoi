// +build !windows

package chezmoi

import "os"

func IsPrivate(fs PrivacyStater, file string, umask os.FileMode) bool {
	info, err := fs.Stat(file)
	if err != nil {
		return false
	}

	return info.Mode().Perm()&^umask == 0700&^umask
}
