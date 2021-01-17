package chezmoi

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// A TARReaderSystem a system constructed from reading a TAR archive.
type TARReaderSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	fileInfos map[AbsPath]os.FileInfo
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
		fileInfos: make(map[AbsPath]os.FileInfo),
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
			contents, err := ioutil.ReadAll(tarReader)
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

// FileInfos retunrs s's os.FileInfos.
func (s *TARReaderSystem) FileInfos() map[AbsPath]os.FileInfo {
	return s.fileInfos
}

// Lstat implements System.Lstat.
func (s *TARReaderSystem) Lstat(filename AbsPath) (os.FileInfo, error) {
	fileInfo, ok := s.fileInfos[filename]
	if !ok {
		return nil, os.ErrNotExist
	}
	return fileInfo, nil
}

// ReadFile implements System.ReadFile.
func (s *TARReaderSystem) ReadFile(filename AbsPath) ([]byte, error) {
	if contents, ok := s.contents[filename]; ok {
		return contents, nil
	}
	if _, ok := s.fileInfos[filename]; ok {
		return nil, os.ErrInvalid
	}
	return nil, os.ErrNotExist
}

// Readlink implements System.Readlink.
func (s *TARReaderSystem) Readlink(name AbsPath) (string, error) {
	if linkname, ok := s.linkname[name]; ok {
		return linkname, nil
	}
	if _, ok := s.fileInfos[name]; ok {
		return "", os.ErrInvalid
	}
	return "", os.ErrNotExist
}
