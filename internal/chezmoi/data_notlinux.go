// +build !linux

package chezmoi

import (
	"github.com/twpayne/go-vfs/v3"
)

// KernelInfo returns nothing on non-Linux systems.
func KernelInfo(fileSystem vfs.FS) (map[string]string, error) {
	return nil, nil
}

// OSRelease returns nothing on non-Linux systems.
func OSRelease(fileSystem vfs.FS) (map[string]string, error) {
	return nil, nil
}
