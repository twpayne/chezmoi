package chezmoi

import (
	"io/fs"
	"os/exec"

	"github.com/rs/zerolog/log"
	vfs "github.com/twpayne/go-vfs/v4"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A DebugSystem logs all calls to a System.
type DebugSystem struct {
	system System
}

// NewDebugSystem returns a new DebugSystem.
func NewDebugSystem(system System) *DebugSystem {
	return &DebugSystem{
		system: system,
	}
}

// Chmod implements System.Chmod.
func (s *DebugSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	err := s.system.Chmod(name, mode)
	log.Err(err).
		Str("name", string(name)).
		Int("mode", int(mode)).
		Msg("Chmod")
	return err
}

// Glob implements System.Glob.
func (s *DebugSystem) Glob(name string) ([]string, error) {
	matches, err := s.system.Glob(name)
	log.Err(err).
		Str("name", name).
		Strs("matches", matches).
		Msg("Glob")
	return matches, err
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *DebugSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	output, err := s.system.IdempotentCmdCombinedOutput(cmd)
	log.Err(err).
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		Bytes("output", chezmoilog.Output(output, err)).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("IdempotentCmdCombinedOutput")
	return output, err
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *DebugSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	output, err := s.system.IdempotentCmdOutput(cmd)
	log.Err(err).
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		Bytes("output", chezmoilog.Output(output, err)).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("IdempotentCmdOutput")
	return output, err
}

// Lstat implements System.Lstat.
func (s *DebugSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	info, err := s.system.Lstat(name)
	log.Err(err).
		Str("name", string(name)).
		Msg("Lstat")
	return info, err
}

// Mkdir implements System.Mkdir.
func (s *DebugSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	err := s.system.Mkdir(name, perm)
	log.Err(err).
		Str("name", string(name)).
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
	log.Err(err).
		Str("name", string(name)).
		Msg("ReadDir")
	return dirEntries, err
}

// ReadFile implements System.ReadFile.
func (s *DebugSystem) ReadFile(name AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(name)
	log.Err(err).
		Str("filename", string(name)).
		Bytes("data", chezmoilog.Output(data, err)).
		Msg("ReadFile")
	return data, err
}

// Readlink implements System.Readlink.
func (s *DebugSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.system.Readlink(name)
	log.Err(err).
		Str("name", string(name)).
		Str("linkname", linkname).
		Msg("Readlink")
	return linkname, err
}

// RemoveAll implements System.RemoveAll.
func (s *DebugSystem) RemoveAll(name AbsPath) error {
	err := s.system.RemoveAll(name)
	log.Err(err).
		Str("name", string(name)).
		Msg("RemoveAll")
	return err
}

// Rename implements System.Rename.
func (s *DebugSystem) Rename(oldpath, newpath AbsPath) error {
	err := s.system.Rename(oldpath, newpath)
	log.Err(err).
		Str("oldpath", string(oldpath)).
		Str("newpath", string(newpath)).
		Msg("Rename")
	return err
}

// RunCmd implements System.RunCmd.
func (s *DebugSystem) RunCmd(cmd *exec.Cmd) error {
	err := s.system.RunCmd(cmd)
	log.Err(err).
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunCmd")
	return err
}

// RunIdempotentCmd implements System.RunIdempotentCmd.
func (s *DebugSystem) RunIdempotentCmd(cmd *exec.Cmd) error {
	err := s.system.RunIdempotentCmd(cmd)
	log.Err(err).
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunIdempotentCmd")
	return err
}

// RunScript implements System.RunScript.
func (s *DebugSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	err := s.system.RunScript(scriptname, dir, data, interpreter)
	log.Err(err).
		Str("scriptname", string(scriptname)).
		Str("dir", string(dir)).
		Bytes("data", chezmoilog.Output(data, err)).
		Object("interpreter", interpreter).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunScript")
	return err
}

// Stat implements System.Stat.
func (s *DebugSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	info, err := s.system.Stat(name)
	log.Err(err).
		Str("name", string(name)).
		Msg("Stat")
	return info, err
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DebugSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DebugSystem) WriteFile(name AbsPath, data []byte, perm fs.FileMode) error {
	err := s.system.WriteFile(name, data, perm)
	log.Err(err).
		Str("name", string(name)).
		Bytes("data", chezmoilog.Output(data, err)).
		Int("perm", int(perm)).
		Msg("WriteFile")
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *DebugSystem) WriteSymlink(oldname string, newname AbsPath) error {
	err := s.system.WriteSymlink(oldname, newname)
	log.Err(err).
		Str("oldname", oldname).
		Str("newname", string(newname)).
		Msg("WriteSymlink")
	return err
}
