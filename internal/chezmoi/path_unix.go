//go:build !windows
// +build !windows

package chezmoi

import (
	"path/filepath"
	"strings"
)

// NewAbsPathFromExtPath returns a new AbsPath by converting extPath to use
// slashes, performing tilde expansion, and making the path absolute.
func NewAbsPathFromExtPath(extPath string, homeDirAbsPath AbsPath) (AbsPath, error) {
	tildeSlashPath := expandTilde(filepath.ToSlash(extPath), homeDirAbsPath)
	if filepath.IsAbs(tildeSlashPath) {
		return NewAbsPath(tildeSlashPath), nil
	}
	slashPathAbsPath, err := filepath.Abs(tildeSlashPath)
	if err != nil {
		return EmptyAbsPath, err
	}
	return NewAbsPath(slashPathAbsPath), nil
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

// expandTilde expands a leading tilde in path.
func expandTilde(path string, homeDirAbsPath AbsPath) string {
	switch {
	case path == "~":
		return homeDirAbsPath.String()
	case strings.HasPrefix(path, "~/"):
		return homeDirAbsPath.JoinStr(path[2:]).String()
	default:
		return path
	}
}

// normalizeLinkname returns linkname normalized. On non-Windows systems, it
// returns linkname unchanged.
func normalizeLinkname(linkname string) string {
	return linkname
}
