package chezmoi

import (
	"io/fs"
)

const nativeLineEnding = "\r\n"

// isExecutable returns false on Windows.
func isExecutable(fileInfo fs.FileInfo) bool {
	return false
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
