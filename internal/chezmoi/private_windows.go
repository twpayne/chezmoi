// +build windows

package chezmoi

import (
	vfs "github.com/twpayne/go-vfs"
)

// IsPrivate always returns want on Windows.
func IsPrivate(fs vfs.Stater, path string, want bool) (bool, error) {
	return want, nil
}
