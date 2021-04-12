package chezmoi

import (
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	vfs "github.com/twpayne/go-vfs/v2"

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
func (s *DebugSystem) Chmod(name AbsPath, mode os.FileMode) error {
	err := s.system.Chmod(name, mode)
	log.Logger.Debug().
		Str("name", string(name)).
		Int("mode", int(mode)).
		Err(err).
		Msg("Chmod")
	return err
}

// Glob implements System.Glob.
func (s *DebugSystem) Glob(name string) ([]string, error) {
	matches, err := s.system.Glob(name)
	log.Logger.Debug().
		Str("name", name).
		Strs("matches", matches).
		Err(err).
		Msg("Glob")
	return matches, err
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *DebugSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	output, err := s.system.IdempotentCmdCombinedOutput(cmd)
	log.Logger.Debug().
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		Bytes("output", chezmoilog.FirstFewBytes(output)).
		Err(err).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("IdempotentCmdCombinedOutput")
	return output, err
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *DebugSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	output, err := s.system.IdempotentCmdOutput(cmd)
	log.Logger.Debug().
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		Bytes("output", chezmoilog.FirstFewBytes(output)).
		Err(err).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("IdempotentCmdOutput")
	return output, err
}

// Lstat implements System.Lstat.
func (s *DebugSystem) Lstat(name AbsPath) (os.FileInfo, error) {
	info, err := s.system.Lstat(name)
	log.Logger.Debug().
		Str("name", string(name)).
		Err(err).
		Msg("Lstat")
	return info, err
}

// Mkdir implements System.Mkdir.
func (s *DebugSystem) Mkdir(name AbsPath, perm os.FileMode) error {
	err := s.system.Mkdir(name, perm)
	log.Logger.Debug().
		Str("name", string(name)).
		Int("perm", int(perm)).
		Err(err).
		Msg("Mkdir")
	return err
}

// RawPath implements System.RawPath.
func (s *DebugSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *DebugSystem) ReadDir(name AbsPath) ([]os.DirEntry, error) {
	dirEntries, err := s.system.ReadDir(name)
	log.Logger.Debug().
		Str("name", string(name)).
		Err(err).
		Msg("ReadDir")
	return dirEntries, err
}

// ReadFile implements System.ReadFile.
func (s *DebugSystem) ReadFile(name AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(name)
	log.Logger.Debug().
		Str("filename", string(name)).
		Bytes("data", chezmoilog.FirstFewBytes(data)).
		Err(err).
		Msg("ReadFile")
	return data, err
}

// Readlink implements System.Readlink.
func (s *DebugSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.system.Readlink(name)
	log.Logger.Debug().
		Str("name", string(name)).
		Str("linkname", linkname).
		Err(err).
		Msg("Readlink")
	return linkname, err
}

// RemoveAll implements System.RemoveAll.
func (s *DebugSystem) RemoveAll(name AbsPath) error {
	err := s.system.RemoveAll(name)
	log.Logger.Debug().
		Str("name", string(name)).
		Err(err).
		Msg("RemoveAll")
	return err
}

// Rename implements System.Rename.
func (s *DebugSystem) Rename(oldpath, newpath AbsPath) error {
	err := s.system.Rename(oldpath, newpath)
	log.Logger.Debug().
		Str("oldpath", string(oldpath)).
		Str("newpath", string(newpath)).
		Err(err).
		Msg("Rename")
	return err
}

// RunCmd implements System.RunCmd.
func (s *DebugSystem) RunCmd(cmd *exec.Cmd) error {
	err := s.system.RunCmd(cmd)
	log.Logger.Debug().
		EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: cmd}).
		Err(err).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunCmd")
	return err
}

// RunScript implements System.RunScript.
func (s *DebugSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte) error {
	err := s.system.RunScript(scriptname, dir, data)
	log.Logger.Debug().
		Str("scriptname", string(scriptname)).
		Str("dir", string(dir)).
		Bytes("data", chezmoilog.FirstFewBytes(data)).
		Err(err).
		EmbedObject(chezmoilog.OSExecExitErrorLogObject{Err: err}).
		Msg("RunScript")
	return err
}

// Stat implements System.Stat.
func (s *DebugSystem) Stat(name AbsPath) (os.FileInfo, error) {
	info, err := s.system.Stat(name)
	log.Logger.Debug().
		Str("name", string(name)).
		Err(err).
		Msg("Stat")
	return info, err
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DebugSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DebugSystem) WriteFile(name AbsPath, data []byte, perm os.FileMode) error {
	err := s.system.WriteFile(name, data, perm)
	log.Logger.Debug().
		Str("name", string(name)).
		Bytes("data", chezmoilog.FirstFewBytes(data)).
		Int("perm", int(perm)).
		Err(err).
		Msg("WriteFile")
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *DebugSystem) WriteSymlink(oldname string, newname AbsPath) error {
	err := s.system.WriteSymlink(oldname, newname)
	log.Logger.Debug().
		Str("oldname", oldname).
		Str("newname", string(newname)).
		Err(err).
		Msg("WriteSymlink")
	return err
}
