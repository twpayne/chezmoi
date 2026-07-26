package chezmoi

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"time"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zip"
	"github.com/klauspost/compress/zstd"
	"github.com/nwaples/rardecode/v2"
	"github.com/ulikunitz/xz"

	"chezmoi.io/chezmoi/v2/internal/chezmoiset"
)

// An ArchiveFormat is an archive format and implements the
// github.com/spf13/pflag.Value interface.
type ArchiveFormat string

// Archive formats.
const (
	ArchiveFormatUnknown ArchiveFormat = ""
	ArchiveFormatRar     ArchiveFormat = "rar"
	ArchiveFormatTar     ArchiveFormat = "tar"
	ArchiveFormatTarBz2  ArchiveFormat = "tar.bz2"
	ArchiveFormatTarGz   ArchiveFormat = "tar.gz"
	ArchiveFormatTarXz   ArchiveFormat = "tar.xz"
	ArchiveFormatTarZst  ArchiveFormat = "tar.zst"
	ArchiveFormatZip     ArchiveFormat = "zip"
)

// An WalkArchiveFunc is called once for each entry in an archive.
type WalkArchiveFunc func(name RelPath, info fs.FileInfo, r io.Reader, linkname string) error

// GuessArchiveFormat guesses the archive format from the name and data.
func GuessArchiveFormat(name string, data []byte) ArchiveFormat {
	switch nameLower := strings.ToLower(name); {
	case strings.HasSuffix(nameLower, ".rar"):
		return ArchiveFormatRar
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
	case len(data) >= 6 && bytes.Equal(data[:6], []byte{'R', 'a', 'r', '!', 0x1a, 0x07}):
		return ArchiveFormatRar
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
	switch format {
	case ArchiveFormatRar:
		return walkArchiveRar(bytes.NewReader(data), f)
	case ArchiveFormatZip:
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

func implicitTarDirHeader(dir RelPath, modTime time.Time) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     dir.String(),
		Mode:     0o777,
		Size:     0,
		ModTime:  modTime,
	}
}

// A RARFileInfo wraps a *rardecode.FileHeader so that it implements
// [fs.FileInfo].
type RARFileInfo struct {
	*rardecode.FileHeader
}

func (i RARFileInfo) IsDir() bool        { return i.FileHeader.IsDir }
func (i RARFileInfo) ModTime() time.Time { return i.ModificationTime }
func (i RARFileInfo) Mode() fs.FileMode  { return i.FileHeader.Mode() }
func (i RARFileInfo) Name() string       { return i.FileHeader.Name }
func (i RARFileInfo) Size() int64        { return i.UnPackedSize }
func (i RARFileInfo) Sys() any           { return nil }

// walkArchiveRar walks over all the entries in a rar archive.
func walkArchiveRar(r io.Reader, f WalkArchiveFunc) error {
	rarReader, err := rardecode.NewReader(r)
	if err != nil {
		return err
	}

	// Process a single header. Remember skipped directories.
	skippedDirs := chezmoiset.New[RelPath]()
	processHeader := func(header *rardecode.FileHeader) error {
		relPath, err := NewUntrustedRelPath(strings.TrimSuffix(header.Name, "/"))
		if err != nil {
			return err
		}
		if header.IsDir && skippedDirs.Contains(relPath) {
			return fs.SkipDir
		}
		switch err := f(relPath, RARFileInfo{FileHeader: header}, rarReader, ""); {
		case errors.Is(err, fs.SkipDir):
			skippedDirs.Add(relPath)
			return fs.SkipDir
		case err != nil:
			return err
		default:
			return nil
		}
	}

HEADER:
	for {
		fileHeader, err := rarReader.Next()
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		}
		switch err := processHeader(fileHeader); {
		case errors.Is(err, fs.SkipDir):
			continue HEADER
		case errors.Is(err, fs.SkipAll):
			return nil
		case err != nil:
			return err
		}
	}
}

// walkArchiveTar walks over all the entries in a tar archive.
func walkArchiveTar(r io.Reader, f WalkArchiveFunc) error {
	tarReader := tar.NewReader(r)

	// Process a single header, which might be an implicit parent directory.
	// Remember already-seen directories so we do not visit them twice.
	seenDirErrors := make(map[RelPath]error)
	processHeader := func(relPath RelPath, header *tar.Header) error {
		if header.Typeflag == tar.TypeDir {
			if seenDirError, ok := seenDirErrors[relPath]; ok {
				return seenDirError
			}
		}
		switch err := f(relPath, header.FileInfo(), tarReader, header.Linkname); {
		case errors.Is(err, fs.SkipDir):
			seenDirErrors[relPath] = fs.SkipDir
			return fs.SkipDir
		case err != nil:
			return err
		default:
			if header.Typeflag == tar.TypeDir {
				seenDirErrors[relPath] = nil
			}
			return nil
		}
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

		relPath, err := NewUntrustedRelPath(strings.TrimSuffix(header.Name, "/"))
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeReg, tar.TypeSymlink, tar.TypeDir:
			// Create implicit parent directories.
			if dir, _ := relPath.Split(); !dir.IsEmpty() {
				dirComponents := dir.TrimTrailingSlash().SplitAll()
				for i := range dirComponents {
					implicitParentDirRelPath := NewRelPathFromComponents(dirComponents[:i+1]...)
					implicitParentDirHeader := implicitTarDirHeader(implicitParentDirRelPath, header.ModTime)
					switch err := processHeader(implicitParentDirRelPath, implicitParentDirHeader); {
					case errors.Is(err, fs.SkipDir):
						continue HEADER
					case errors.Is(err, fs.SkipAll):
						return nil
					case err != nil:
						return err
					}
				}
			}
			switch err := processHeader(relPath, header); {
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

	// Process a single header, which might be an implicit parent directory.
	// Remember already-seen directories so we do not visit them twice.
	seenDirErrors := make(map[RelPath]error)
	processHeader := func(relPath RelPath, fileInfo fs.FileInfo) error {
		if fileInfo.IsDir() {
			if seenDirError, ok := seenDirErrors[relPath]; ok {
				return seenDirError
			}
		}
		switch err := f(relPath, fileInfo, nil, ""); {
		case errors.Is(err, fs.SkipDir):
			seenDirErrors[relPath] = fs.SkipDir
			return fs.SkipDir
		case err != nil:
			return err
		default:
			if fileInfo.IsDir() {
				seenDirErrors[relPath] = nil
			}
			return nil
		}
	}

FILE:
	for _, zipFile := range zipReader.File {
		relPath, err := NewUntrustedRelPath(zipFile.Name)
		if err != nil {
			return err
		}

		fileInfo := zipFile.FileInfo()
		switch fileInfo.Mode() & fs.ModeType {
		case 0, fs.ModeDir, fs.ModeSymlink:
			// Create implicit parent directories.
			if dir, _ := relPath.Split(); !dir.IsEmpty() {
				dirComponents := dir.TrimTrailingSlash().SplitAll()
				for i := range dirComponents {
					implicitParentDirRelPath := NewRelPathFromComponents(dirComponents[:i+1]...)
					implicitParentDirFileInfo := implicitTarDirHeader(implicitParentDirRelPath, fileInfo.ModTime()).FileInfo()
					switch err := processHeader(implicitParentDirRelPath, implicitParentDirFileInfo); {
					case errors.Is(err, fs.SkipDir):
						continue FILE
					case errors.Is(err, fs.SkipAll):
						return nil
					case err != nil:
						return err
					}
				}
			}
		default:
			return fmt.Errorf("%s: unsupported file mode %o", zipFile.Name, fileInfo.Mode())
		}

		switch fileInfo.Mode() & fs.ModeType {
		case 0:
			zipFileReader, err := zipFile.Open()
			if err != nil {
				return err
			}
			err = f(relPath, fileInfo, zipFileReader, "")
			err2 := zipFileReader.Close()
			switch {
			case errors.Is(err, fs.SkipAll):
				return err2
			case err != nil:
				return err
			case err2 != nil:
				return err2
			}
		case fs.ModeDir:
			switch err := processHeader(relPath, fileInfo); {
			case errors.Is(err, fs.SkipDir):
				continue FILE
			case errors.Is(err, fs.SkipAll):
				return nil
			case err != nil:
				return err
			}
		case fs.ModeSymlink:
			zipFileReader, err := zipFile.Open()
			if err != nil {
				return err
			}
			linknameBytes, err := io.ReadAll(zipFileReader)
			if err != nil {
				return err
			}
			if err := zipFileReader.Close(); err != nil {
				return err
			}
			switch err := f(relPath, fileInfo, nil, string(linknameBytes)); {
			case errors.Is(err, fs.SkipAll):
				return nil
			case err != nil:
				return err
			}
		}
	}

	return nil
}
