package chezmoi

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/coreos/go-semver/semver"
)

// An ExitWithCodeError indicates the main program should exit with the given
// code.
type ExitWithCodeError struct {
	exitCode int
}

func NewExitWithCodeError(exitCode int) *ExitWithCodeError {
	return &ExitWithCodeError{
		exitCode: exitCode,
	}
}

func (e *ExitWithCodeError) Error() string {
	return fmt.Sprintf("exit status %d", e.exitCode)
}

func (e *ExitWithCodeError) ExitCode() int {
	return e.exitCode
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

type InconsistentStateError struct {
	targetRelPath RelPath
	origins       []string
}

func (e *InconsistentStateError) Error() string {
	format := "%s: inconsistent state (%s)"
	return fmt.Sprintf(format, e.targetRelPath, strings.Join(e.origins, ", "))
}

type InvalidPathError string

func (e InvalidPathError) Error() string {
	return string(e) + ": invalid path"
}

type NotInAbsDirError struct {
	pathAbsPath AbsPath
	dirAbsPath  AbsPath
}

func (e *NotInAbsDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathAbsPath, e.dirAbsPath)
}

type NotInRelDirError struct {
	pathRelPath RelPath
	dirRelPath  RelPath
}

func (e *NotInRelDirError) Error() string {
	return fmt.Sprintf("%s: not in %s", e.pathRelPath, e.dirRelPath)
}

type UnsupportedFileTypeError struct {
	absPath AbsPath
	mode    fs.FileMode
}

func (e *UnsupportedFileTypeError) Error() string {
	return fmt.Sprintf("%s: unsupported file type %s", e.absPath, modeTypeName(e.mode))
}
