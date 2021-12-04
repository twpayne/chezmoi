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

// An WalkArchiveFunc is called once for each entry in an archive.
type WalkArchiveFunc func(name string, info fs.FileInfo, r io.Reader, linkname string) error

// Set implements github.com/spf13/pflag.Value.Set.
func (f *ArchiveFormat) Set(s string) error {
	*f = ArchiveFormat(s)
	return nil
}

// String implements github.com/spf13/pflag.Value.String.
func (f ArchiveFormat) String() string {
	return string(f)
}

// Type implements github.com/spf13/pflag.Value.Type.
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

// WalkArchive walks over all the entries in an archive.
func WalkArchive(data []byte, format ArchiveFormat, f WalkArchiveFunc) error {
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

// isTarArchive returns if r looks like a tar archive.
func isTarArchive(r io.Reader) bool {
	tarReader := tar.NewReader(r)
	_, err := tarReader.Next()
	return err == nil
}

// walkArchiveTar walks over all the entries in a tar archive.
func walkArchiveTar(r io.Reader, f WalkArchiveFunc) error {
	tarReader := tar.NewReader(r)
	var skipPrefixes []string
FOR:
	for {
		header, err := tarReader.Next()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}
		for _, skipPrefix := range skipPrefixes {
			if strings.HasPrefix(header.Name, skipPrefix) {
				continue FOR
			}
		}
		name := strings.TrimSuffix(header.Name, "/")
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg:
			switch err := f(name, header.FileInfo(), tarReader, ""); {
			case errors.Is(err, Break):
				return nil
			case errors.Is(err, Skip):
				skipPrefixes = append(skipPrefixes, header.Name)
			case err != nil:
				return err
			}
		case tar.TypeSymlink:
			switch err := f(name, header.FileInfo(), nil, header.Linkname); {
			case errors.Is(err, Break):
				return nil
			case err != nil:
				return err
			}
		case tar.TypeXGlobalHeader:
		default:
			return fmt.Errorf("%s: unsupported typeflag '%c'", header.Name, header.Typeflag)
		}
	}
}

// walkArchiveZip walks over all the entries in a zip archive.
func walkArchiveZip(r io.ReaderAt, size int64, f WalkArchiveFunc) error {
	zipReader, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}
	var skipPrefixes []string
FOR:
	for _, zipFile := range zipReader.File {
		zipFileReader, err := zipFile.Open()
		if err != nil {
			return err
		}
		name := path.Clean(zipFile.Name)
		if strings.HasPrefix(name, "../") || strings.Contains(name, "/../") {
			return fmt.Errorf("%s: invalid filename", zipFile.Name)
		}
		for _, skipPrefix := range skipPrefixes {
			if strings.HasPrefix(name, skipPrefix) {
				continue FOR
			}
		}
		switch fileInfo := zipFile.FileInfo(); fileInfo.Mode() & fs.ModeType {
		case 0:
			err = f(name, fileInfo, zipFileReader, "")
		case fs.ModeDir:
			err = f(name, fileInfo, nil, "")
		case fs.ModeSymlink:
			var linknameBytes []byte
			linknameBytes, err = io.ReadAll(zipFileReader)
			if err != nil {
				return err
			}
			err = f(name, fileInfo, nil, string(linknameBytes))
		}
		zipFileReader.Close()
		switch {
		case errors.Is(err, Break):
			return nil
		case errors.Is(err, Skip):
			skipPrefixes = append(skipPrefixes, name)
		case err != nil:
			return err
		}
	}
	return nil
}
