package chezmoi

import (
	"path/filepath"
	"strings"
)

var devNullAbsPath = NewAbsPath("NUL:")

// NewAbsPathFromExtPath returns a new AbsPath by converting extPath to use
// slashes, performing tilde expansion, making the path absolute, and converting
// the volume name to uppercase.
func NewAbsPathFromExtPath(extPath string, homeDirAbsPath AbsPath) (AbsPath, error) {
	extPath = filepath.Clean(extPath)
	switch {
	case extPath == "~":
		return homeDirAbsPath, nil
	case len(extPath) >= 2 && extPath[0] == '~' && isSlash(extPath[1]):
		return homeDirAbsPath.JoinString(filepath.ToSlash(extPath[2:])), nil
	case filepath.IsAbs(extPath):
		return NewAbsPath(volumeNameToUpper(extPath)).ToSlash(), nil
	default:
		extPath, err := filepath.Abs(extPath)
		if err != nil {
			return EmptyAbsPath, err
		}
		return NewAbsPath(volumeNameToUpper(extPath)).ToSlash(), nil
	}
}

// NormalizePath returns path normalized. On Windows, normalized paths are
// absolute paths with a uppercase volume name and forward slashes.
func NormalizePath(path string) (AbsPath, error) {
	var err error
	path, err = filepath.Abs(path)
	if err != nil {
		return EmptyAbsPath, err
	}
	if n := volumeNameLen(path); n > 0 {
		path = strings.ToUpper(path[:n]) + path[n:]
	}
	return NewAbsPath(path).ToSlash(), nil
}

// normalizeLinkname returns linkname normalized. On Windows, backslashes are
// converted to forward slashes and if linkname is an absolute path then the
// volume name is converted to uppercase.
func normalizeLinkname(linkname string) string {
	if filepath.IsAbs(linkname) {
		return filepath.ToSlash(volumeNameToUpper(linkname))
	}
	return filepath.ToSlash(linkname)
}

// volumeNameLen returns length of the leading volume name on Windows. It
// returns 0 elsewhere.
func volumeNameLen(path string) int {
	if len(path) < 2 {
		return 0
	}
	// with drive letter
	c := path[0]
	if path[1] == ':' && ('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z') {
		return 2
	}
	// is it UNC? https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx
	if l := len(path); l >= 5 && isSlash(path[0]) && isSlash(path[1]) && !isSlash(path[2]) && path[2] != '.' {
		// first, leading `\\` and next shouldn't be `\`. its server name.
		for n := 3; n < l-1; n++ {
			// second, next '\' shouldn't be repeated.
			if isSlash(path[n]) {
				n++
				// third, following something characters. its share name.
				if !isSlash(path[n]) {
					if path[n] == '.' {
						break
					}
					for ; n < l; n++ {
						if isSlash(path[n]) {
							break
						}
					}
					return n
				}
				break
			}
		}
	}
	return 0
}

// volumeNameToUpper returns path with the volume name converted to uppercase.
func volumeNameToUpper(path string) string {
	if n := volumeNameLen(path); n > 0 {
		return strings.ToUpper(path[:n]) + path[n:]
	}
	return path
}
