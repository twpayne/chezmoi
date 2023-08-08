package chezmoi

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
)

const nativeLineEnding = "\r\n"

var pathExts = strings.Split(os.Getenv("PATHEXT"), string(filepath.ListSeparator))

// findExecutableExtensions returns valid OS executable extensions for a given executable basename - does not check the
// existence.
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

// IsExecutable checks if the file is a regular file and has an extension listed
// in the PATHEXT environment variable as per
// https://www.nextofwindows.com/what-is-pathext-environment-variable-in-windows.
func IsExecutable(fileInfo fs.FileInfo) bool {
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

// isPrivate returns false on Windows.
func isPrivate(fileInfo fs.FileInfo) bool {
	return false
}

// isReadOnly returns false on Windows.
func isReadOnly(fileInfo fs.FileInfo) bool {
	return false
}

// isSlash returns if c is a slash character.
func isSlash(c byte) bool {
	return c == '\\' || c == '/'
}
