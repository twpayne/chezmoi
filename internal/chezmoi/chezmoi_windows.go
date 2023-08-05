package chezmoi

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const nativeLineEnding = "\r\n"

var pathExt []string = nil

// getPathExt a singleton that obtains the PathExt environment variable and splits it up using the OS `ListSeparator`
func getPathExt() []string {
	if pathExt == nil {
		pathExt = strings.Split(os.Getenv("PathExt"), string(filepath.ListSeparator))
	}
	return pathExt
}

// isExecutable checks if the file has an extension listed in the `PathExt` variable as per:
// https://www.nextofwindows.com/what-is-pathext-environment-variable-in-windows then checks to see if it's regular file
func isExecutable(fileInfo fs.FileInfo) bool {
	foundPathExt := false
	cmdExt := filepath.Ext(fileInfo.Name())
	if cmdExt != "" {
		for _, ext := range getPathExt() {
			if strings.EqualFold(cmdExt, ext) {
				foundPathExt = true
				break
			}
		}
	}
	return foundPathExt && fileInfo.Mode().IsRegular()
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
