//go:build !windows

package chezmoi

import (
	"io/fs"

	"golang.org/x/sys/unix"
)

const nativeLineEnding = "\n"

func init() {
	Umask = fs.FileMode(unix.Umask(0))
	unix.Umask(int(Umask))
}

// isExecutable returns if fileInfo is executable.
func isExecutable(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o111 != 0
}

// isPrivate returns if fileInfo is private.
func isPrivate(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o77 == 0
}

// isReadOnly returns if fileInfo is read-only.
func isReadOnly(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o222 == 0
}
