package chezmoi

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

// A ArchiveReaderSystem a system constructed from reading an archive.
type ArchiveReaderSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	fileInfos map[AbsPath]fs.FileInfo
	contents  map[AbsPath][]byte
	linkname  map[AbsPath]string
}

// ArchiveReaderSystemOptions are options to NewArchiveReaderSystem.
type ArchiveReaderSystemOptions struct {
	RootAbsPath     AbsPath
	StripComponents int
}

// NewArchiveReaderSystem returns a new ArchiveReaderSystem reading from data
// and using archivePath as a hint for the archive format.
func NewArchiveReaderSystem(
	archivePath string, data []byte, format ArchiveFormat, options ArchiveReaderSystemOptions,
) (*ArchiveReaderSystem, error) {
	s := &ArchiveReaderSystem{
		fileInfos: make(map[AbsPath]fs.FileInfo),
		contents:  make(map[AbsPath][]byte),
		linkname:  make(map[AbsPath]string),
	}

	if format == ArchiveFormatUnknown {
		format = GuessArchiveFormat(archivePath, data)
	}

	if err := WalkArchive(data, format, func(name string, fileInfo fs.FileInfo, r io.Reader, linkname string) error {
		if options.StripComponents > 0 {
			components := strings.Split(name, "/")
			if len(components) <= options.StripComponents {
				return nil
			}
			name = path.Join(components[options.StripComponents:]...)
		}
		if name == "" {
			return nil
		}
		nameAbsPath := options.RootAbsPath.JoinString(name)

		s.fileInfos[nameAbsPath] = fileInfo
		switch {
		case fileInfo.IsDir():
			// Do nothing.
		case fileInfo.Mode()&fs.ModeType == 0:
			contents, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			s.contents[nameAbsPath] = contents
		case fileInfo.Mode()&fs.ModeType == fs.ModeSymlink:
			s.linkname[nameAbsPath] = linkname
		default:
			return fmt.Errorf("%s: unsupported mode %o", name, fileInfo.Mode()&fs.ModeType)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s, nil
}

// FileInfos returns s's fs.FileInfos.
func (s *ArchiveReaderSystem) FileInfos() map[AbsPath]fs.FileInfo {
	return s.fileInfos
}

// Lstat implements System.Lstat.
func (s *ArchiveReaderSystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	fileInfo, ok := s.fileInfos[filename]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return fileInfo, nil
}

// ReadFile implements System.ReadFile.
func (s *ArchiveReaderSystem) ReadFile(name AbsPath) ([]byte, error) {
	if contents, ok := s.contents[name]; ok {
		return contents, nil
	}
	if _, ok := s.fileInfos[name]; ok {
		return nil, fs.ErrInvalid
	}
	return nil, fs.ErrNotExist
}

// Readlink implements System.Readlink.
func (s *ArchiveReaderSystem) Readlink(name AbsPath) (string, error) {
	if linkname, ok := s.linkname[name]; ok {
		return linkname, nil
	}
	if _, ok := s.fileInfos[name]; ok {
		return "", fs.ErrInvalid
	}
	return "", fs.ErrNotExist
}
