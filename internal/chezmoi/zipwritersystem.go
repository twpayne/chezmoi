package chezmoi

import (
	"archive/zip"
	"io"
	"io/fs"
	"time"
)

// A ZIPWriterSystem is a System that writes to a ZIP archive.
type ZIPWriterSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	w        *zip.Writer
	modified time.Time
}

// NewZIPWriterSystem returns a new ZIPWriterSystem that writes a ZIP archive to
// w.
func NewZIPWriterSystem(w io.Writer, modified time.Time) *ZIPWriterSystem {
	return &ZIPWriterSystem{
		w:        zip.NewWriter(w),
		modified: modified,
	}
}

// Close closes m.
func (s *ZIPWriterSystem) Close() error {
	return s.w.Close()
}

// Mkdir implements System.Mkdir.
func (s *ZIPWriterSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	fh := zip.FileHeader{
		Name:     name.String(),
		Modified: s.modified,
	}
	fh.SetMode(fs.ModeDir | perm)
	_, err := s.w.CreateHeader(&fh)
	return err
}

// RunScript implements System.RunScript.
func (s *ZIPWriterSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	return s.WriteFile(NewAbsPath(scriptname.String()), data, 0o700)
}

// WriteFile implements System.WriteFile.
func (s *ZIPWriterSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	fh := zip.FileHeader{
		Name:               filename.String(),
		Method:             zip.Deflate,
		Modified:           s.modified,
		UncompressedSize64: uint64(len(data)),
	}
	fh.SetMode(perm)
	fw, err := s.w.CreateHeader(&fh)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *ZIPWriterSystem) WriteSymlink(oldname string, newname AbsPath) error {
	data := []byte(oldname)
	fh := zip.FileHeader{
		Name:               newname.String(),
		Modified:           s.modified,
		UncompressedSize64: uint64(len(data)),
	}
	fh.SetMode(fs.ModeSymlink)
	fw, err := s.w.CreateHeader(&fh)
	if err != nil {
		return err
	}
	_, err = fw.Write(data)
	return err
}
