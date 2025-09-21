package chezmoi

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const nativeLineEnding = "\r\n"

var pathExts = strings.Split(os.Getenv("PATHEXT"), string(filepath.ListSeparator))

// findExecutableExtensions returns valid OS executable extensions for the
// provided file if it does not already have an extension. The executable
// extensions are derived from %PathExt%.
func findExecutableExtensions(path string) []string {
	cmdExt := filepath.Ext(path)
	if cmdExt != "" {
		return []string{path}
	}
	result := make([]string, len(pathExts))
	withoutSuffix := strings.TrimSuffix(path, cmdExt)
	for i, ext := range pathExts {
		result[i] = withoutSuffix + ext
	}
	return result
}

// IsExecutable checks if the file is a regular file and has an
// extension listed in the PATHEXT environment variable as per
// https://www.nextofwindows.com/what-is-pathext-environment-variable-in-windows.
func IsExecutable(fileInfo fs.FileInfo) bool {
	if fileInfo.Mode().Perm()&0o111 != 0 {
		return true
	}
	if !fileInfo.Mode().IsRegular() {
		return false
	}
	ext := filepath.Ext(fileInfo.Name())
	if ext == "" {
		return false
	}
	return slices.ContainsFunc(pathExts, func(pathExt string) bool {
		return strings.EqualFold(pathExt, ext)
	})
}

// UserHomeDir on Windows returns the value of $HOME if it is set and either
// Cygwin or msys2 is detected, otherwise it falls back to os.UserHomeDir.
func UserHomeDir() (string, error) {
	if os.Getenv("CYGWIN") != "" || os.Getenv("MSYSTEM") != "" {
		if userHomeDir := os.Getenv("HOME"); userHomeDir != "" {
			return userHomeDir, nil
		}
	}
	return os.UserHomeDir()
}

// isSlash returns if c is a slash character.
func isSlash(c byte) bool {
	return c == '\\' || c == '/'
}
