// +build !linux

package chezmoi

import (
	"github.com/twpayne/go-vfs/v2"
)

// KernelInfo returns nothing on non-Linux systems.
func KernelInfo(fs vfs.FS) (map[string]string, error) {
	return nil, nil
}

// OSRelease returns nothing on non-Linux systems.
func OSRelease(fs vfs.FS) (map[string]string, error) {
	return nil, nil
}
