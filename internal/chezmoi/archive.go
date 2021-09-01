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

var errUnknownFormat = errors.New("unknown format")

// An walkArchiveFunc is called once for each entry in an archive.
type walkArchiveFunc func(name string, info fs.FileInfo, r io.Reader, linkname string) error

// walkArchive walks over all the entries in an archive. path is used as a hint
// for the archive format.
func walkArchive(path string, data []byte, f walkArchiveFunc) error {
	pathLower := strings.ToLower(path)
	if strings.HasSuffix(pathLower, ".zip") {
		return walkArchiveZip(bytes.NewReader(data), int64(len(data)), f)
	}
	var r io.Reader = bytes.NewReader(data)
	switch {
	case strings.HasSuffix(pathLower, ".tar"):
	case strings.HasSuffix(pathLower, ".tar.bz2") || strings.HasSuffix(pathLower, ".tbz2"):
		r = bzip2.NewReader(r)
	case strings.HasSuffix(pathLower, ".tar.gz") || strings.HasSuffix(pathLower, ".tgz"):
		var err error
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	default:
		return errUnknownFormat
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
