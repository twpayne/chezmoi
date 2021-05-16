package chezmoi

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

// A TARReaderSystem a system constructed from reading a TAR archive.
type TARReaderSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	fileInfos map[AbsPath]fs.FileInfo
	contents  map[AbsPath][]byte
	linkname  map[AbsPath]string
}

// TARReaderSystemOptions are options to NewTARReaderSystem.
type TARReaderSystemOptions struct {
	RootAbsPath     AbsPath
	StripComponents int
}

// NewTARReaderSystem returns a new TARReaderSystem from tarReader.
func NewTARReaderSystem(tarReader *tar.Reader, options TARReaderSystemOptions) (*TARReaderSystem, error) {
	s := &TARReaderSystem{
		fileInfos: make(map[AbsPath]fs.FileInfo),
		contents:  make(map[AbsPath][]byte),
		linkname:  make(map[AbsPath]string),
	}
FOR:
	for {
		header, err := tarReader.Next()
		switch {
		case errors.Is(err, io.EOF):
			return s, nil
		case err != nil:
			return nil, err
		}

		name := strings.TrimSuffix(header.Name, "/")
		if options.StripComponents > 0 {
			components := strings.Split(name, "/")
			if len(components) <= options.StripComponents {
				continue FOR
			}
			name = strings.Join(components[options.StripComponents:], "/")
		}
		nameAbsPath := options.RootAbsPath.Join(RelPath(name))

		switch header.Typeflag {
		case tar.TypeDir:
			s.fileInfos[nameAbsPath] = header.FileInfo()
		case tar.TypeReg:
			s.fileInfos[nameAbsPath] = header.FileInfo()
			contents, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, err
			}
			s.contents[nameAbsPath] = contents
		case tar.TypeSymlink:
			s.fileInfos[nameAbsPath] = header.FileInfo()
			s.linkname[nameAbsPath] = header.Linkname
		case tar.TypeXGlobalHeader:
			continue FOR
		default:
			return nil, fmt.Errorf("unsupported typeflag '%c'", header.Typeflag)
		}
	}
}

// FileInfos returns s's fs.FileInfos.
func (s *TARReaderSystem) FileInfos() map[AbsPath]fs.FileInfo {
	return s.fileInfos
}

// Lstat implements System.Lstat.
func (s *TARReaderSystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	fileInfo, ok := s.fileInfos[filename]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return fileInfo, nil
}

// ReadFile implements System.ReadFile.
func (s *TARReaderSystem) ReadFile(name AbsPath) ([]byte, error) {
	if contents, ok := s.contents[name]; ok {
		return contents, nil
	}
	if _, ok := s.fileInfos[name]; ok {
		return nil, fs.ErrInvalid
	}
	return nil, fs.ErrNotExist
}

// Readlink implements System.Readlink.
func (s *TARReaderSystem) Readlink(name AbsPath) (string, error) {
	if linkname, ok := s.linkname[name]; ok {
		return linkname, nil
	}
	if _, ok := s.fileInfos[name]; ok {
		return "", fs.ErrInvalid
	}
	return "", fs.ErrNotExist
}
