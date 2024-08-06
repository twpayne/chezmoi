//go:build unix

package chezmoi

import (
	"io/fs"
	"os"

	"golang.org/x/sys/unix"
)

const nativeLineEnding = "\n"

func init() {
	Umask = fs.FileMode(unix.Umask(0))
	unix.Umask(int(Umask))
}

// findExecutableExtensions returns valid OS executable extensions, on unix it
// can be anything.
func findExecutableExtensions(path string) []string {
	return []string{path}
}

// IsExecutable returns if fileInfo is executable.
func IsExecutable(fileInfo fs.FileInfo) bool {
	return fileInfo.Mode().Perm()&0o111 != 0
}

// UserHomeDir on UNIX returns the value of os.UserHomeDir.
func UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
