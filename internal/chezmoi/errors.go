package chezmoi

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// An ExitCodeError indicates the main program should exit with the given
// code.
type ExitCodeError int

func (e ExitCodeError) Error() string {
	return fmt.Sprintf("exit status %d", int(e))
}

// A TooOldError is returned when the source state requires a newer version of
// chezmoi.
type TooOldError struct {
	Have semver.Version
	Need semver.Version
}

func (e *TooOldError) Error() string {
	format := "source state requires chezmoi version %s or later, chezmoi is version %s"
	return fmt.Sprintf(format, e.Need, e.Have)
}

type inconsistentStateError struct {
	targetRelPath RelPath
	origins       []string
}

func (e *inconsistentStateError) Error() string {
	format := "%s: inconsistent state (%s)"
	return fmt.Sprintf(format, e.targetRelPath, strings.Join(e.origins, ", "))
}

type NotInAbsDirError struct {
	pathAbsPath AbsPath
	dirAbsPath  AbsPath
}

func (e *NotInAbsDirError) Error() string {
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
