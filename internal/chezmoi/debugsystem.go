package chezmoi

import (
	"io/fs"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A DebugSystem logs all calls to a System.
type DebugSystem struct {
	logger *zerolog.Logger
	system System
}

// NewDebugSystem returns a new DebugSystem that logs methods on system to logger.
func NewDebugSystem(system System, logger *zerolog.Logger) *DebugSystem {
	return &DebugSystem{
		logger: logger,
		system: system,
	}
}

// Chtimes implements System.Chtimes.
func (s *DebugSystem) Chtimes(name AbsPath, atime, mtime time.Time) error {
	err := s.system.Chtimes(name, atime, mtime)
	s.logger.Err(err).
		Stringer("name", name).
		Time("atime", atime).
		Time("mtime", mtime).
		Msg("Chtimes")
	return err
}

// Chmod implements System.Chmod.
func (s *DebugSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	err := s.system.Chmod(name, mode)
	s.logger.Err(err).
		Stringer("name", name).
		Int("mode", int(mode)).
		Msg("Chmod")
	return err
}

// Glob implements System.Glob.
func (s *DebugSystem) Glob(name string) ([]string, error) {
	matches, err := s.system.Glob(name)
	s.logger.Err(err).
		Str("name", name).
		Strs("matches", matches).
		Msg("Glob")
	return matches, err
}

// Link implements System.Link.
func (s *DebugSystem) Link(oldpath, newpath AbsPath) error {
	err := s.system.Link(oldpath, newpath)
	s.logger.Err(err).
		Stringer("oldpath", oldpath).
		Stringer("newpath", newpath).
		Msg("Link")
	return err
}

// Lstat implements System.Lstat.
func (s *DebugSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	fileInfo, err := s.system.Lstat(name)
	s.logger.Err(err).
		Stringer("name", name).
		Msg("Lstat")
	return fileInfo, err
}

// Mkdir implements System.Mkdir.
func (s *DebugSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	err := s.system.Mkdir(name, perm)
	s.logger.Err(err).
		Stringer("name", name).
		Int("perm", int(perm)).
		Msg("Mkdir")
	return err
}

// RawPath implements System.RawPath.
func (s *DebugSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *DebugSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	dirEntries, err := s.system.ReadDir(name)
	s.logger.Err(err).
		Stringer("name", name).
		Msg("ReadDir")
	return dirEntries, err
}

// ReadFile implements System.ReadFile.
func (s *DebugSystem) ReadFile(name AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(name)
	s.logger.Err(err).
		Stringer("name", name).
		Bytes("data", chezmoilog.Output(data, err)).
		Int("size", len(data)).
		Msg("ReadFile")
	return data, err
}

// Readlink implements System.Readlink.
func (s *DebugSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.system.Readlink(name)
	s.logger.Err(err).
		Stringer("name", name).
		Str("linkname", linkname).
		Msg("Readlink")
	return linkname, err
}

// Remove implements System.Remove.
func (s *DebugSystem) Remove(name AbsPath) error {
	err := s.system.Remove(name)
	s.logger.Err(err).
		Stringer("name", name).
		Msg("Remove")
	return err
}

// RemoveAll implements System.RemoveAll.
func (s *DebugSystem) RemoveAll(name AbsPath) error {
	err := s.system.RemoveAll(name)
	s.logger.Err(err).
		Stringer("name", name).
		Msg("RemoveAll")
	return err
}

// Rename implements System.Rename.
func (s *DebugSystem) Rename(oldpath, newpath AbsPath) error {
	err := s.system.Rename(oldpath, newpath)
	s.logger.Err(err).
		Stringer("oldpath", oldpath).
		Stringer("newpath", newpath).
		Msg("Rename")
	return err
}

// RunCmd implements System.RunCmd.
func (s *DebugSystem) RunCmd(cmd *exec.Cmd) error {
	err := s.system.RunCmd(cmd)
	s.logger.Err(err).
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunCmd")
	return err
}

// RunScript implements System.RunScript.
func (s *DebugSystem) RunScript(
	scriptname RelPath,
	dir AbsPath,
	data []byte,
	options RunScriptOptions,
) error {
	err := s.system.RunScript(scriptname, dir, data, options)
	s.logger.Err(err).
		Stringer("scriptname", scriptname).
		Stringer("dir", dir).
		Bytes("data", chezmoilog.Output(data, err)).
		Object("interpreter", options.Interpreter).
		Str("condition", string(options.Condition)).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunScript")
	return err
}

// Stat implements System.Stat.
func (s *DebugSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	fileInfo, err := s.system.Stat(name)
	s.logger.Err(err).
		Stringer("name", name).
		Msg("Stat")
	return fileInfo, err
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DebugSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DebugSystem) WriteFile(name AbsPath, data []byte, perm fs.FileMode) error {
	err := s.system.WriteFile(name, data, perm)
	s.logger.Err(err).
		Stringer("name", name).
		Bytes("data", chezmoilog.Output(data, err)).
		Int("perm", int(perm)).
		Int("size", len(data)).
		Msg("WriteFile")
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *DebugSystem) WriteSymlink(oldname string, newname AbsPath) error {
	err := s.system.WriteSymlink(oldname, newname)
	s.logger.Err(err).
		Str("oldname", oldname).
		Stringer("newname", newname).
		Msg("WriteSymlink")
	return err
}
