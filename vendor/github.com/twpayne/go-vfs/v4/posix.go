//go:build !windows
// +build !windows

package vfs

import (
	"strings"
	"syscall"
)

//nolint:gochecknoglobals
var ignoreErrnoInContains = map[syscall.Errno]struct{}{
	syscall.ELOOP:        {},
	syscall.EMLINK:       {},
	syscall.ENAMETOOLONG: {},
	syscall.ENOENT:       {},
	syscall.EOVERFLOW:    {},
}

// relativizePath, on POSIX systems, just returns path.
func relativizePath(path string) string {
	return path
}

// trimPrefix, on POSIX systems, trims prefix from path.
func trimPrefix(path, prefix string) (string, error) {
	return strings.TrimPrefix(path, prefix), nil
}
