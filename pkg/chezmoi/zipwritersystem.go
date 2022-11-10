package chezmoi

import (
	"archive/zip"
	"io"
	"io/fs"
	"os/exec"
	"time"
)

// A ZIPWriterSystem is a System that writes to a ZIP archive.
type ZIPWriterSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	zipWriter *zip.Writer
	modified  time.Time
}

// NewZIPWriterSystem returns a new ZIPWriterSystem that writes a ZIP archive to
// w.
func NewZIPWriterSystem(w io.Writer, modified time.Time) *ZIPWriterSystem {
	return &ZIPWriterSystem{
		zipWriter: zip.NewWriter(w),
		modified:  modified,
	}
}

// Close closes m.
func (s *ZIPWriterSystem) Close() error {
	return s.zipWriter.Close()
}

// Mkdir implements System.Mkdir.
func (s *ZIPWriterSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	fileHeader := zip.FileHeader{
		Name:     name.String(),
		Modified: s.modified,
	}
	fileHeader.SetMode(fs.ModeDir | perm)
	_, err := s.zipWriter.CreateHeader(&fileHeader)
	return err
}

// RunCmd implements System.RunCmd.
func (s *ZIPWriterSystem) RunCmd(cmd *exec.Cmd) error {
	return nil
}

// RunScript implements System.RunScript.
func (s *ZIPWriterSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	return s.WriteFile(NewAbsPath(scriptname.String()), data, 0o700)
}

// WriteFile implements System.WriteFile.
func (s *ZIPWriterSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	fileHeader := zip.FileHeader{
		Name:               filename.String(),
		Method:             zip.Deflate,
		Modified:           s.modified,
		UncompressedSize64: uint64(len(data)),
	}
	fileHeader.SetMode(perm)
	fileWriter, err := s.zipWriter.CreateHeader(&fileHeader)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(data)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *ZIPWriterSystem) WriteSymlink(oldname string, newname AbsPath) error {
	data := []byte(oldname)
	fileHeader := zip.FileHeader{
		Name:               newname.String(),
		Modified:           s.modified,
		UncompressedSize64: uint64(len(data)),
	}
	fileHeader.SetMode(fs.ModeSymlink)
	fileWriter, err := s.zipWriter.CreateHeader(&fileHeader)
	if err != nil {
		return err
	}
	_, err = fileWriter.Write(data)
	return err
}
