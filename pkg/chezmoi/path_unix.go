//go:build !windows

package chezmoi

import (
	"path/filepath"
	"strings"
)

var devNullAbsPath = NewAbsPath("/dev/null")

// NewAbsPathFromExtPath returns a new AbsPath by converting extPath to use
// slashes, performing tilde expansion, and making the path absolute.
func NewAbsPathFromExtPath(extPath string, homeDirAbsPath AbsPath) (AbsPath, error) {
	switch {
	case extPath == "~":
		return homeDirAbsPath, nil
	case strings.HasPrefix(extPath, "~/"):
		return homeDirAbsPath.JoinString(extPath[2:]), nil
	case filepath.IsAbs(extPath):
		return NewAbsPath(extPath), nil
	default:
		absPath, err := filepath.Abs(extPath)
		if err != nil {
			return EmptyAbsPath, err
		}
		return NewAbsPath(absPath), nil
	}
}

// NormalizePath returns path normalized. On non-Windows systems, normalized
// paths are absolute paths.
func NormalizePath(path string) (AbsPath, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return EmptyAbsPath, err
	}
	return NewAbsPath(absPath), nil
}

// normalizeLinkname returns linkname normalized. On non-Windows systems, it
// returns linkname unchanged.
func normalizeLinkname(linkname string) string {
	return linkname
}
