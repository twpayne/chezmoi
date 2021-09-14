package chezmoi

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
)

// An ArchiveFormat is an archive format and implements the
// github.com/spf13/pflag.Value interface.
type ArchiveFormat string

// Archive formats.
const (
	ArchiveFormatUnknown ArchiveFormat = ""
	ArchiveFormatTar     ArchiveFormat = "tar"
	ArchiveFormatTarBz2  ArchiveFormat = "tar.bz2"
	ArchiveFormatTarGz   ArchiveFormat = "tar.gz"
	ArchiveFormatTbz2    ArchiveFormat = "tbz2"
	ArchiveFormatTgz     ArchiveFormat = "tgz"
	ArchiveFormatZip     ArchiveFormat = "zip"
)

type InvalidArchiveFormatError string

func (e InvalidArchiveFormatError) Error() string {
	if e == InvalidArchiveFormatError(ArchiveFormatUnknown) {
		return "invalid archive format"
	}
	return fmt.Sprintf("%s: invalid archive format", string(e))
}

// An walkArchiveFunc is called once for each entry in an archive.
type walkArchiveFunc func(name string, info fs.FileInfo, r io.Reader, linkname string) error

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
// and using path as a hint for the archive format.
func NewArchiveReaderSystem(path string, data []byte, format ArchiveFormat, options ArchiveReaderSystemOptions) (*ArchiveReaderSystem, error) {
	s := &ArchiveReaderSystem{
		fileInfos: make(map[AbsPath]fs.FileInfo),
		contents:  make(map[AbsPath][]byte),
		linkname:  make(map[AbsPath]string),
	}

	if format == ArchiveFormatUnknown {
		format = GuessArchiveFormat(path, data)
	}

	if err := walkArchive(data, format, func(name string, info fs.FileInfo, r io.Reader, linkname string) error {
		if options.StripComponents > 0 {
			components := strings.Split(name, "/")
			if len(components) <= options.StripComponents {
				return nil
			}
			name = strings.Join(components[options.StripComponents:], "/")
		}
		if name == "" {
			return nil
		}
		nameAbsPath := options.RootAbsPath.Join(RelPath(name))

		s.fileInfos[nameAbsPath] = info
		switch {
		case info.IsDir():
		case info.Mode()&fs.ModeType == 0:
			contents, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			s.contents[nameAbsPath] = contents
		case info.Mode()&fs.ModeType == fs.ModeSymlink:
			s.linkname[nameAbsPath] = linkname
		default:
			return fmt.Errorf("%s: unsupported mode %o", name, info.Mode()&fs.ModeType)
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

func (f *ArchiveFormat) Set(s string) error {
	*f = ArchiveFormat(s)
	return nil
}

func (f ArchiveFormat) String() string {
	return string(f)
}

func (f ArchiveFormat) Type() string {
	return "format"
}

// GuessArchiveFormat guesses the archive format from the path and data.
func GuessArchiveFormat(path string, data []byte) ArchiveFormat {
	switch pathLower := strings.ToLower(path); {
	case strings.HasSuffix(pathLower, ".tar"):
		return ArchiveFormatTar
	case strings.HasSuffix(pathLower, ".tar.bz2") || strings.HasSuffix(pathLower, ".tbz2"):
		return ArchiveFormatTarBz2
	case strings.HasSuffix(pathLower, ".tar.gz") || strings.HasSuffix(pathLower, ".tgz"):
		return ArchiveFormatTarGz
	case strings.HasSuffix(pathLower, ".zip"):
		return ArchiveFormatZip
	}

	switch {
	case len(data) >= 3 && bytes.Equal(data[:3], []byte{0x1f, 0x8b, 0x08}):
		return ArchiveFormatTarGz
	case len(data) >= 4 && bytes.Equal(data[:4], []byte{'P', 'K', 0x03, 0x04}):
		return ArchiveFormatZip
	case isTarArchive(bytes.NewReader(data)):
		return ArchiveFormatTar
	case isTarArchive(bzip2.NewReader(bytes.NewReader(data))):
		return ArchiveFormatTarBz2
	}

	return ArchiveFormatUnknown
}

// isTarArchive returns if r looks like a tar archive.
func isTarArchive(r io.Reader) bool {
	tarReader := tar.NewReader(r)
	_, err := tarReader.Next()
	return err == nil
}

// walkArchive walks over all the entries in an archive.
func walkArchive(data []byte, format ArchiveFormat, f walkArchiveFunc) error {
	if format == ArchiveFormatZip {
		return walkArchiveZip(bytes.NewReader(data), int64(len(data)), f)
	}
	var r io.Reader = bytes.NewReader(data)
	switch format {
	case ArchiveFormatTar:
	case ArchiveFormatTarBz2, ArchiveFormatTbz2:
		r = bzip2.NewReader(r)
	case ArchiveFormatTarGz, ArchiveFormatTgz:
		var err error
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	default:
		return InvalidArchiveFormatError(format)
	}
	return walkArchiveTar(r, f)
}

// walkArchiveTar walks over all the entries in a tar archive.
func walkArchiveTar(r io.Reader, f walkArchiveFunc) error {
	tarReader := tar.NewReader(r)
	for {
		header, err := tarReader.Next()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}
		name := strings.TrimSuffix(header.Name, "/")
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg:
			if err := f(name, header.FileInfo(), tarReader, ""); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := f(name, header.FileInfo(), nil, header.Linkname); err != nil {
				return err
			}
		case tar.TypeXGlobalHeader:
		default:
			return fmt.Errorf("%s: unsupported typeflag '%c'", header.Name, header.Typeflag)
		}
	}
}

// walkArchiveZip walks over all the entries in a zip archive.
func walkArchiveZip(r io.ReaderAt, size int64, f walkArchiveFunc) error {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}
	for _, zipFile := range zipReader.File {
		zipFileReader, err := zipFile.Open()
		if err != nil {
			return err
		}
		name := path.Clean(zipFile.Name)
		if strings.HasPrefix(name, "../") {
			return fmt.Errorf("%s: invalid filename", zipFile.Name)
		}
		err = f(name, zipFile.FileInfo(), zipFileReader, "")
		zipFileReader.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
