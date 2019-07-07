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

func IsPrivate(fs PrivacyStater, path string) (bool, error) {
	rawPath, err := fs.RawPath(path)
	if err != nil {
		return false, err
	}

	resolvedPath, err := resolveSymlink(rawPath)
	if err != nil {
		return false, err
	}

	mode, err := acl.GetEffectiveAccessMode(resolvedPath)
	if err != nil {
		return false, err
	}

	return mode&07 == 0, nil
}
