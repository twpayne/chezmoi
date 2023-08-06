//go:build windows
// +build windows

package vfs

import (
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

var ignoreErrnoInContains = map[syscall.Errno]struct{}{
	syscall.ELOOP:                       {},
	syscall.EMLINK:                      {},
	syscall.ENAMETOOLONG:                {},
	syscall.ENOENT:                      {},
	syscall.EOVERFLOW:                   {},
	windows.ERROR_CANT_RESOLVE_FILENAME: {},
}

// relativizePath, on Windows, strips any leading volume name from path. Since
// this is used to prepare paths to have the prefix prepended, returned values
// use slashes instead of backslashes.
func relativizePath(path string) string {
	if volumeName := filepath.VolumeName(path); volumeName != "" {
		path = path[len(volumeName):]
	}
	return filepath.ToSlash(path)
}

// trimPrefix, on Windows, trims prefix from path and returns an absolute path.
// prefix must be a /-separated path. Since this is used to prepare results to
// be returned to the calling client, returned values use backslashes instead of
// slashes
func trimPrefix(path, prefix string) (string, error) {
	trimmedPath, err := filepath.Abs(strings.TrimPrefix(filepath.ToSlash(path), prefix))
	if err != nil {
		return "", err
	}
	return filepath.FromSlash(trimmedPath), nil
}
