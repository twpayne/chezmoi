// +build windows

package chezmoi

import (
	"os"
	"path/filepath"

	"github.com/hectane/go-acl"
)

// maxSymLinks is the maximum number of symlinks to follow. Use the same default
// value as Linux (as of kernel 4.19).
const maxSymLinks = 40

func resolveSymlink(file string) (string, error) {
	// If file is a symlink, get the path it links to. This emulates unix-style
	// behavior, where symlinks can't have their own independent permissions.

	resolved := file
	for i := 0; i < maxSymLinks; i++ {
		fi, err := os.Lstat(resolved)
		if err != nil {
			return "", err
		}

		if fi.Mode()&os.ModeSymlink == 0 {
			// Not a link, all done.
			break
		}

		next, err := os.Readlink(resolved)
		if err != nil {
			return "", err
		}

		if next != "" && !filepath.IsAbs(next) {
			resolved = filepath.Join(filepath.Dir(resolved), next)
		}
	}

	return resolved, nil
}

func IsPrivate(fs PrivacyStater, file string, umask os.FileMode) bool {
	file, err := fs.RawPath(file)
	if err != nil {
		return false
	}

	file, err = resolveSymlink(file)
	if err != nil {
		return false
	}

	mode, err := acl.GetEffectiveAccessMode(file)
	if err != nil {
		return false
	}

	return (uint32(mode) & 0007) == 0
}
