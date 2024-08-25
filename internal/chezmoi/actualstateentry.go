package chezmoi

import (
	"crypto/sha256"
	"errors"
	"io/fs"
	"sync"
)

// An ActualStateEntry represents the actual state of an entry in the
// filesystem.
type ActualStateEntry interface {
	EntryState() (*EntryState, error)
	Path() AbsPath
	Remove(system System) error
	OriginString() string
}

// A ActualStateAbsent represents the absence of an entry in the filesystem.
type ActualStateAbsent struct {
	absPath AbsPath
}

// A ActualStateDir represents the state of a directory in the filesystem.
type ActualStateDir struct {
	absPath AbsPath
	perm    fs.FileMode
}

// A ActualStateFile represents the state of a file in the filesystem.
type ActualStateFile struct {
	absPath      AbsPath
	perm         fs.FileMode
	contentsFunc func() ([]byte, error)
}

// A ActualStateSymlink represents the state of a symlink in the filesystem.
type ActualStateSymlink struct {
	absPath      AbsPath
	linknameFunc func() (string, error)
}

// NewActualStateEntry returns a new ActualStateEntry populated with absPath
// from system.
func NewActualStateEntry(system System, absPath AbsPath, fileInfo fs.FileInfo, err error) (ActualStateEntry, error) {
	if fileInfo == nil {
		fileInfo, err = system.Lstat(absPath)
	}
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return &ActualStateAbsent{
			absPath: absPath,
		}, nil
	case err != nil:
		return nil, err
	}
	switch fileInfo.Mode().Type() {
	case 0:
		return &ActualStateFile{
			absPath: absPath,
			perm:    fileInfo.Mode().Perm(),
			contentsFunc: sync.OnceValues(func() ([]byte, error) {
				return system.ReadFile(absPath)
			}),
		}, nil
	case fs.ModeDir:
		return &ActualStateDir{
			absPath: absPath,
			perm:    fileInfo.Mode().Perm(),
		}, nil
	case fs.ModeSymlink:
		return &ActualStateSymlink{
			absPath: absPath,
			linknameFunc: sync.OnceValues(func() (string, error) {
				linkname, err := system.Readlink(absPath)
				if err != nil {
					return "", err
				}
				return normalizeLinkname(linkname), nil
			}),
		}, nil
	default:
		return nil, &unsupportedFileTypeError{
			absPath: absPath,
			mode:    fileInfo.Mode(),
		}
	}
}

// EntryState returns s's entry state.
func (s *ActualStateAbsent) EntryState() (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeRemove,
	}, nil
}

// Path returns s's path.
func (s *ActualStateAbsent) Path() AbsPath {
	return s.absPath
}

// Remove removes s.
func (s *ActualStateAbsent) Remove(system System) error {
	return nil
}

// OriginString returns s's origin.
func (s *ActualStateAbsent) OriginString() string {
	return s.absPath.String()
}

// EntryState returns s's entry state.
func (s *ActualStateDir) EntryState() (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeDir,
		Mode: fs.ModeDir | s.perm,
	}, nil
}

// Path returns s's path.
func (s *ActualStateDir) Path() AbsPath {
	return s.absPath
}

// Remove removes s.
func (s *ActualStateDir) Remove(system System) error {
	return system.RemoveAll(s.absPath)
}

// OriginString returns s's origin.
func (s *ActualStateDir) OriginString() string {
	return s.absPath.String()
}

// Contents returns s's contents.
func (s *ActualStateFile) Contents() ([]byte, error) {
	return s.contentsFunc()
}

// EntryState returns s's entry state.
func (s *ActualStateFile) EntryState() (*EntryState, error) {
	contents, err := s.Contents()
	if err != nil {
		return nil, err
	}
	contentsSHA256 := sha256.Sum256(contents)
	return &EntryState{
		Type:           EntryStateTypeFile,
		Mode:           s.perm,
		ContentsSHA256: HexBytes(contentsSHA256[:]),
		contents:       contents,
	}, nil
}

// Path returns s's path.
func (s *ActualStateFile) Path() AbsPath {
	return s.absPath
}

// Perm returns s's perm.
func (s *ActualStateFile) Perm() fs.FileMode {
	return s.perm
}

// Remove removes s.
func (s *ActualStateFile) Remove(system System) error {
	return system.RemoveAll(s.absPath)
}

// OriginString returns s's origin.
func (s *ActualStateFile) OriginString() string {
	return s.absPath.String()
}

// EntryState returns s's entry state.
func (s *ActualStateSymlink) EntryState() (*EntryState, error) {
	linkname, err := s.Linkname()
	if err != nil {
		return nil, err
	}
	linknameSHA256 := sha256.Sum256([]byte(linkname))
	return &EntryState{
		Type:           EntryStateTypeSymlink,
		ContentsSHA256: HexBytes(linknameSHA256[:]),
		contents:       []byte(linkname),
	}, nil
}

// Linkname returns s's linkname.
func (s *ActualStateSymlink) Linkname() (string, error) {
	return s.linknameFunc()
}

// Path returns s's path.
func (s *ActualStateSymlink) Path() AbsPath {
	return s.absPath
}

// Remove removes s.
func (s *ActualStateSymlink) Remove(system System) error {
	return system.RemoveAll(s.absPath)
}

// OriginString returns s's origin.
func (s *ActualStateSymlink) OriginString() string {
	return s.absPath.String()
}
