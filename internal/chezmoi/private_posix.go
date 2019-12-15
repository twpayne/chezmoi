// +build !windows

package chezmoi

import (
	vfs "github.com/twpayne/go-vfs"
)

// IsPrivate returns whether path should be considered private.
func IsPrivate(fs vfs.Stater, path string, want bool) (bool, error) {
	info, err := fs.Stat(path)
	if err != nil {
		return false, err
	}
	return info.Mode().Perm()&077 == 0, nil
}
