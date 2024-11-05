package chezmoi

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
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
	ArchiveFormatTarXz   ArchiveFormat = "tar.xz"
	ArchiveFormatTarZst  ArchiveFormat = "tar.zst"
	ArchiveFormatZip     ArchiveFormat = "zip"
)

// An WalkArchiveFunc is called once for each entry in an archive.
type WalkArchiveFunc func(name string, info fs.FileInfo, r io.Reader, linkname string) error

// GuessArchiveFormat guesses the archive format from the name and data.
func GuessArchiveFormat(name string, data []byte) ArchiveFormat {
	switch nameLower := strings.ToLower(name); {
	case strings.HasSuffix(nameLower, ".tar"):
		return ArchiveFormatTar
	case strings.HasSuffix(nameLower, ".tar.bz2") || strings.HasSuffix(nameLower, ".tbz2"):
		return ArchiveFormatTarBz2
	case strings.HasSuffix(nameLower, ".tar.gz") || strings.HasSuffix(nameLower, ".tgz"):
		return ArchiveFormatTarGz
	case strings.HasSuffix(nameLower, ".tar.xz") || strings.HasSuffix(nameLower, ".txz"):
		return ArchiveFormatTarXz
	case strings.HasSuffix(nameLower, ".tar.zst"):
		return ArchiveFormatTarZst
	case strings.HasSuffix(nameLower, ".zip"):
		return ArchiveFormatZip
	}

	switch {
	case len(data) >= 3 && bytes.Equal(data[:3], []byte{0x1f, 0x8b, 0x08}):
		return ArchiveFormatTarGz
	case len(data) >= 4 && bytes.Equal(data[:4], []byte{'P', 'K', 0x03, 0x04}):
		return ArchiveFormatZip
	case len(data) >= xz.HeaderLen && xz.ValidHeader(data):
		return ArchiveFormatTarXz
	case (&zstd.Header{}).Decode(data) == nil:
		return ArchiveFormatTarZst
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
	// r will read bytes in tar format.
	var r io.Reader = bytes.NewReader(data)
	switch format {
	case ArchiveFormatTar:
		// Already in tar format, do nothing.
	case ArchiveFormatTarBz2:
		// Decompress with bzip2.
		r = bzip2.NewReader(r)
	case ArchiveFormatTarGz:
		// Decompress with gzip.
		var err error
		r, err = gzip.NewReader(r)
		if err != nil {
			return err
		}
	case ArchiveFormatTarXz:
		// Decompress with xz.
		var err error
		r, err = xz.NewReader(r)
		if err != nil {
			return err
		}
	case ArchiveFormatTarZst:
		// Decompress with zstd.
		var err error
		r, err = zstd.NewReader(r)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unknown archive format", format)
	}
	return walkArchiveTar(r, f)
}

// isTarArchive returns if r looks like a tar archive.
func isTarArchive(r io.Reader) bool {
	tarReader := tar.NewReader(r)
	_, err := tarReader.Next()
	return err == nil
}

func implicitDirHeader(dir string, modTime time.Time) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     dir,
		Mode:     0o777,
		Size:     0,
		ModTime:  modTime,
	}
}

// walkArchiveTar walks over all the entries in a tar archive.
func walkArchiveTar(r io.Reader, f WalkArchiveFunc) error {
	tarReader := tar.NewReader(r)
	var skippedDirPrefixes []string
	seenDirs := chezmoiset.New[string]()
	processHeader := func(header *tar.Header, dir string) error {
		for _, skippedDirPrefix := range skippedDirPrefixes {
			if strings.HasPrefix(header.Name, skippedDirPrefix) {
				return fs.SkipDir
			}
		}
		if seenDirs.Contains(dir) {
			return nil
		}
		seenDirs.Add(dir)
		name := strings.TrimSuffix(header.Name, "/")
		switch err := f(name, header.FileInfo(), tarReader, header.Linkname); {
		case errors.Is(err, fs.SkipDir):
			skippedDirPrefixes = append(skippedDirPrefixes, header.Name)
		case err != nil:
			return err
		}
		return nil
	}
HEADER:
	for {
		header, err := tarReader.Next()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}
		switch header.Typeflag {
		case tar.TypeReg, tar.TypeDir, tar.TypeSymlink:
			if header.Typeflag == tar.TypeReg {
				dirs, _ := path.Split(header.Name)
				dirComponents := strings.Split(strings.TrimSuffix(dirs, "/"), "/")
				for i := range dirComponents {
					if dir := strings.Join(dirComponents[0:i+1], "/"); dir != "" {
						switch err := processHeader(implicitDirHeader(dir+"/", header.ModTime), dir+"/"); {
						case errors.Is(err, fs.SkipDir):
							continue HEADER
						case errors.Is(err, fs.SkipAll):
							return nil
						case err != nil:
							return err
						}
					}
				}
			}
			switch err := processHeader(header, header.Name); {
			case errors.Is(err, fs.SkipDir):
				continue HEADER
			case errors.Is(err, fs.SkipAll):
				return nil
			case err != nil:
				return err
			}
		case tar.TypeXGlobalHeader:
			// Do nothing.
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
	var skippedDirPrefixes []string
	seenDirs := chezmoiset.New[string]()
	processHeader := func(fileInfo fs.FileInfo, dir string) error {
		for _, skippedDirPrefix := range skippedDirPrefixes {
			if strings.HasPrefix(dir, skippedDirPrefix) {
				return fs.SkipDir
			}
		}
		if seenDirs.Contains(dir) {
			return nil
		}
		seenDirs.Add(dir)
		name := strings.TrimSuffix(dir, "/")
		dirFileInfo := implicitDirHeader(dir, fileInfo.ModTime()).FileInfo()
		switch err := f(name, dirFileInfo, nil, ""); {
		case errors.Is(err, fs.SkipDir):
			skippedDirPrefixes = append(skippedDirPrefixes, dir)
			return err
		case err != nil:
			return err
		}
		return nil
	}
FILE:
	for _, zipFile := range zipReader.File {
		zipFileReader, err := zipFile.Open()
		if err != nil {
			return err
		}

		name := path.Clean(zipFile.Name)
		if strings.HasPrefix(name, "../") || strings.Contains(name, "/../") {
			return fmt.Errorf("%s: invalid filename", zipFile.Name)
		}

		for _, skippedDirPrefix := range skippedDirPrefixes {
			if strings.HasPrefix(zipFile.Name, skippedDirPrefix) {
				continue FILE
			}
		}

		switch fileInfo := zipFile.FileInfo(); fileInfo.Mode() & fs.ModeType {
		case 0:
			dirs, _ := path.Split(name)
			dirComponents := strings.Split(strings.TrimSuffix(dirs, "/"), "/")
			for i := range dirComponents {
				if dir := strings.Join(dirComponents[0:i+1], "/"); dir != "" {
					switch err := processHeader(fileInfo, dir+"/"); {
					case errors.Is(err, fs.SkipDir):
						continue FILE
					case errors.Is(err, fs.SkipAll):
						return nil
					case err != nil:
						return err
					}
				}
			}

			err = f(name, fileInfo, zipFileReader, "")
		case fs.ModeDir:
			err = processHeader(fileInfo, name+"/")
		case fs.ModeSymlink:
			var linknameBytes []byte
			linknameBytes, err = io.ReadAll(zipFileReader)
			if err != nil {
				return err
			}
			err = f(name, fileInfo, nil, string(linknameBytes))
		}

		err2 := zipFileReader.Close()

		switch {
		case errors.Is(err, fs.SkipDir):
			skippedDirPrefixes = append(skippedDirPrefixes, zipFile.Name+"/")
		case errors.Is(err, fs.SkipAll):
			return nil
		case err != nil:
			return err
		}

		if err2 != nil {
			return err2
		}
	}
	return nil
}
