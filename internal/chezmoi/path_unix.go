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
		return AbsPath(tildeSlashPath), nil
	}
	slashPathAbsPath, err := filepath.Abs(tildeSlashPath)
	if err != nil {
		return "", err
	}
	return AbsPath(slashPathAbsPath), nil
}

// NormalizePath returns path normalized. On non-Windows systems, normalized
// paths are absolute paths.
func NormalizePath(path string) (AbsPath, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return AbsPath(absPath), nil
}

// expandTilde expands a leading tilde in path.
func expandTilde(path string, homeDirAbsPath AbsPath) string {
	switch {
	case path == "~":
		return string(homeDirAbsPath)
	case strings.HasPrefix(path, "~/"):
		return string(homeDirAbsPath.Join(RelPath(path[2:])))
	default:
		return path
	}
}

// normalizeLinkname returns linkname normalized. On non-Windows systems, it
// returns linkname unchanged.
func normalizeLinkname(linkname string) string {
	return linkname
}
