package chezmoi

import (
	"io/fs"
	"os/exec"

	vfs "github.com/twpayne/go-vfs/v3"
)

// DryRunSystem is an System that reads from, but does not write to, to
// a wrapped System.
type DryRunSystem struct {
	system   System
	modified bool
}

// NewDryRunSystem returns a new DryRunSystem that wraps fs.
func NewDryRunSystem(system System) *DryRunSystem {
	return &DryRunSystem{
		system: system,
	}
}

// Chmod implements System.Chmod.
func (s *DryRunSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	s.setModified()
	return nil
}

// Glob implements System.Glob.
func (s *DryRunSystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *DryRunSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdCombinedOutput(cmd)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *DryRunSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdOutput(cmd)
}

// Lstat implements System.Lstat.
func (s *DryRunSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *DryRunSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	s.setModified()
	return nil
}

// Modified returns true if a method that would have modified the wrapped system
// has been called.
func (s *DryRunSystem) Modified() bool {
	return s.modified
}

// RawPath implements System.RawPath.
func (s *DryRunSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *DryRunSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.system.ReadDir(name)
}

// ReadFile implements System.ReadFile.
func (s *DryRunSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.system.ReadFile(name)
}

// Readlink implements System.Readlink.
func (s *DryRunSystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// RemoveAll implements System.RemoveAll.
func (s *DryRunSystem) RemoveAll(AbsPath) error {
	s.setModified()
	return nil
}

// Rename implements System.Rename.
func (s *DryRunSystem) Rename(oldpath, newpath AbsPath) error {
	s.setModified()
	return nil
}

// RunCmd implements System.RunCmd.
func (s *DryRunSystem) RunCmd(cmd *exec.Cmd) error {
	s.setModified()
	return nil
}

// RunIdempotentCmd implements System.RunIdempotentCmd.
func (s *DryRunSystem) RunIdempotentCmd(cmd *exec.Cmd) error {
	return s.system.RunIdempotentCmd(cmd)
}

// RunScript implements System.RunScript.
func (s *DryRunSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	s.setModified()
	return nil
}

// Stat implements System.Stat.
func (s *DryRunSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DryRunSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DryRunSystem) WriteFile(AbsPath, []byte, fs.FileMode) error {
	s.setModified()
	return nil
}

// WriteSymlink implements System.WriteSymlink.
func (s *DryRunSystem) WriteSymlink(string, AbsPath) error {
	s.setModified()
	return nil
}

// setModified sets the modified flag to true. It is a separate function so that
// it can act as a convenient breakpoint for detecting modifications to the
// underlying system.
func (s *DryRunSystem) setModified() {
	s.modified = true
}
