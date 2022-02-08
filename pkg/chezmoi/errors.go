package chezmoi

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// An ExitCodeError indicates the the main program should exit with the given
// code.
type ExitCodeError int

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit status %d", int(e))
}

// A TooOldErrror is returned when the source state requires a newer version of
// chezmoi.
type TooOldError struct {
	Have semver.Version
	Need semver.Version
}

func (e *TooOldError) Error() string {
	return fmt.Sprintf("source state requires chezmoi version %s or later, chezmoi is version %s", e.Need, e.Have)
}

type inconsistentStateError struct {
	targetRelPath RelPath
	origins       []string
}

func (e *inconsistentStateError) Error() string {
	return fmt.Sprintf("%s: inconsistent state (%s)", e.targetRelPath, strings.Join(e.origins, ", "))
}

type notInAbsDirError struct {
	pathAbsPath AbsPath
	dirAbsPath  AbsPath
}

func (e *notInAbsDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathAbsPath, e.dirAbsPath)
}

type notInRelDirError struct {
	pathRelPath RelPath
	dirRelPath  RelPath
}

func (e *notInRelDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathRelPath, e.dirRelPath)
}

type unsupportedFileTypeError struct {
	absPath AbsPath
	mode    fs.FileMode
}

func (e *unsupportedFileTypeError) Error() string {
	return fmt.Sprintf("%s: unsupported file type %s", e.absPath, modeTypeName(e.mode))
}
