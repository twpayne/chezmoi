package chezmoi

import (
	"archive/tar"
	"io"
	"io/fs"
)

// A TARWriterSystem is a System that writes to a TAR archive.
type TARWriterSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	w              *tar.Writer
	headerTemplate tar.Header
}

// NewTARWriterSystem returns a new TARWriterSystem that writes a TAR file to w.
func NewTARWriterSystem(w io.Writer, headerTemplate tar.Header) *TARWriterSystem {
	return &TARWriterSystem{
		w:              tar.NewWriter(w),
		headerTemplate: headerTemplate,
	}
}

// Close closes m.
func (s *TARWriterSystem) Close() error {
	return s.w.Close()
}

// Mkdir implements System.Mkdir.
func (s *TARWriterSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeDir
	header.Name = name.String() + "/"
	header.Mode = int64(perm)
	return s.w.WriteHeader(&header)
}

// RunScript implements System.RunScript.
func (s *TARWriterSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	return s.WriteFile(NewAbsPath(scriptname.String()), data, 0o700)
}

// WriteFile implements System.WriteFile.
func (s *TARWriterSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = filename.String()
	header.Size = int64(len(data))
	header.Mode = int64(perm)
	if err := s.w.WriteHeader(&header); err != nil {
		return err
	}
	_, err := s.w.Write(data)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *TARWriterSystem) WriteSymlink(oldname string, newname AbsPath) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeSymlink
	header.Name = newname.String()
	header.Linkname = oldname
	return s.w.WriteHeader(&header)
}
