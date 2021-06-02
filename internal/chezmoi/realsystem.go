package chezmoi

import (
	"errors"
	"io/fs"
	"os/exec"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rs/zerolog/log"
	vfs "github.com/twpayne/go-vfs/v3"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// An Interpreter interprets scripts.
type Interpreter struct {
	Command string   `mapstructure:"command"`
	Args    []string `mapstructure:"args"`
}

// Glob implements System.Glob.
func (s *RealSystem) Glob(pattern string) ([]string, error) {
	return doublestar.Glob(s.UnderlyingFS(), pattern)
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *RealSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return chezmoilog.LogCmdCombinedOutput(log.Logger, cmd)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *RealSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// Lstat implements System.Lstat.
func (s *RealSystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Lstat(string(filename))
}

// Mkdir implements System.Mkdir.
func (s *RealSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	return s.fileSystem.Mkdir(string(name), perm)
}

// RawPath implements System.RawPath.
func (s *RealSystem) RawPath(absPath AbsPath) (AbsPath, error) {
	rawAbsPath, err := s.fileSystem.RawPath(string(absPath))
	if err != nil {
		return "", err
	}
	return AbsPath(rawAbsPath), nil
}

// ReadDir implements System.ReadDir.
func (s *RealSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.fileSystem.ReadDir(string(name))
}

// ReadFile implements System.ReadFile.
func (s *RealSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.fileSystem.ReadFile(string(name))
}

// RemoveAll implements System.RemoveAll.
func (s *RealSystem) RemoveAll(name AbsPath) error {
	return s.fileSystem.RemoveAll(string(name))
}

// Rename implements System.Rename.
func (s *RealSystem) Rename(oldpath, newpath AbsPath) error {
	return s.fileSystem.Rename(string(oldpath), string(newpath))
}

// RunCmd implements System.RunCmd.
func (s *RealSystem) RunCmd(cmd *exec.Cmd) error {
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// Stat implements System.Stat.
func (s *RealSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Stat(string(name))
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *RealSystem) UnderlyingFS() vfs.FS {
	return s.fileSystem
}

// getScriptWorkingDir returns the script's working directory.
//
// If this is a before_ script then the requested working directory may not
// actually exist yet, so search through the parent directory hierarchy till
// we find a suitable working directory.
func (s *RealSystem) getScriptWorkingDir(dir AbsPath) (string, error) {
	// This should always terminate because dir will eventually become ".", i.e.
	// the current directory.
	for {
		switch info, err := s.Stat(dir); {
		case err == nil && info.IsDir():
			// dir exists and is a directory. Use it.
			dirRawAbsPath, err := s.RawPath(dir)
			if err != nil {
				return "", err
			}
			return string(dirRawAbsPath), nil
		case err == nil || errors.Is(err, fs.ErrNotExist):
			// Either dir does not exist, or it exists and is not a directory.
			dir = dir.Dir()
		default:
			// Some other error occurred.
			return "", err
		}
	}
}
