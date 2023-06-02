package chezmoi

import (
	"archive/tar"
	"io"
	"io/fs"
	"os/exec"
)

// A TarWriterSystem is a System that writes to a tar archive.
type TarWriterSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	tarWriter      *tar.Writer
	headerTemplate tar.Header
}

// NewTarWriterSystem returns a new TarWriterSystem that writes a tar file to w.
func NewTarWriterSystem(w io.Writer, headerTemplate tar.Header) *TarWriterSystem {
	return &TarWriterSystem{
		tarWriter:      tar.NewWriter(w),
		headerTemplate: headerTemplate,
	}
}

// Close closes m.
func (s *TarWriterSystem) Close() error {
	return s.tarWriter.Close()
}

// Mkdir implements System.Mkdir.
func (s *TarWriterSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeDir
	header.Name = name.String() + "/"
	header.Mode = int64(perm)
	return s.tarWriter.WriteHeader(&header)
}

// RunCmd implements System.RunCmd.
func (s *TarWriterSystem) RunCmd(cmd *exec.Cmd) error {
	return nil
}

// RunScript implements System.RunScript.
func (s *TarWriterSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	return s.WriteFile(NewAbsPath(scriptname.String()), data, 0o700)
}

// UnderlyingSystem implements System.UnderlyingSystem.
func (s *TarWriterSystem) UnderlyingSystem() System {
	return s
}

// WriteFile implements System.WriteFile.
func (s *TarWriterSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = filename.String()
	header.Size = int64(len(data))
	header.Mode = int64(perm)
	if err := s.tarWriter.WriteHeader(&header); err != nil {
		return err
	}
	_, err := s.tarWriter.Write(data)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *TarWriterSystem) WriteSymlink(oldname string, newname AbsPath) error {
	header := s.headerTemplate
	header.Typeflag = tar.TypeSymlink
	header.Name = newname.String()
	header.Linkname = oldname
	return s.tarWriter.WriteHeader(&header)
}
