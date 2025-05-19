package chezmoi

import (
	"io/fs"
	"log/slog"
	"os/exec"
	"time"

	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

// A DebugSystem logs all calls to a System.
type DebugSystem struct {
	logger *slog.Logger
	system System
}

// NewDebugSystem returns a new DebugSystem that logs methods on system to logger.
func NewDebugSystem(system System, logger *slog.Logger) *DebugSystem {
	return &DebugSystem{
		logger: logger,
		system: system,
	}
}

// Chtimes implements System.Chtimes.
func (s *DebugSystem) Chtimes(name AbsPath, atime, mtime time.Time) error {
	err := s.system.Chtimes(name, atime, mtime)
	chezmoilog.InfoOrError(s.logger, "Chtimes", err,
		chezmoilog.Stringer("name", name),
		slog.Time("atime", atime),
		slog.Time("mtime", mtime),
	)
	return err
}

// Chmod implements System.Chmod.
func (s *DebugSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	err := s.system.Chmod(name, mode)
	chezmoilog.InfoOrError(s.logger, "Chmod", err,
		chezmoilog.Stringer("name", name),
		slog.Int("mode", int(mode)),
	)
	return err
}

// Glob implements System.Glob.
func (s *DebugSystem) Glob(name string) ([]string, error) {
	matches, err := s.system.Glob(name)
	chezmoilog.InfoOrError(s.logger, "Glob", err,
		slog.String("name", name),
		slog.Any("matches", matches),
	)
	return matches, err
}

// Link implements System.Link.
func (s *DebugSystem) Link(oldPath, newPath AbsPath) error {
	err := s.system.Link(oldPath, newPath)
	chezmoilog.InfoOrError(s.logger, "Link", err,
		chezmoilog.Stringer("oldPath", oldPath),
		chezmoilog.Stringer("newPath", newPath),
	)
	return err
}

// Lstat implements System.Lstat.
func (s *DebugSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	fileInfo, err := s.system.Lstat(name)
	chezmoilog.InfoOrError(s.logger, "Lstat", err,
		chezmoilog.Stringer("name", name),
	)
	return fileInfo, err
}

// Mkdir implements System.Mkdir.
func (s *DebugSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	err := s.system.Mkdir(name, perm)
	chezmoilog.InfoOrError(s.logger, "Mkdir", err,
		chezmoilog.Stringer("name", name),
		slog.Int("perm", int(perm)),
	)
	return err
}

// RawPath implements System.RawPath.
func (s *DebugSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *DebugSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	dirEntries, err := s.system.ReadDir(name)
	chezmoilog.InfoOrError(s.logger, "ReadDir", err,
		chezmoilog.Stringer("name", name),
	)
	return dirEntries, err
}

// ReadFile implements System.ReadFile.
func (s *DebugSystem) ReadFile(name AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(name)
	chezmoilog.InfoOrError(s.logger, "ReadFile", err,
		chezmoilog.Stringer("name", name),
		slog.Int("size", len(data)),
		chezmoilog.FirstFewBytes("data", data),
	)
	return data, err
}

// Readlink implements System.Readlink.
func (s *DebugSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.system.Readlink(name)
	chezmoilog.InfoOrError(s.logger, "ReadLink", err,
		slog.String("linkname", linkname),
	)
	return linkname, err
}

// Remove implements System.Remove.
func (s *DebugSystem) Remove(name AbsPath) error {
	err := s.system.Remove(name)
	chezmoilog.InfoOrError(s.logger, "Remove", err,
		chezmoilog.Stringer("name", name),
	)
	return err
}

// RemoveAll implements System.RemoveAll.
func (s *DebugSystem) RemoveAll(name AbsPath) error {
	err := s.system.RemoveAll(name)
	chezmoilog.InfoOrError(s.logger, "RemoveAll", err,
		chezmoilog.Stringer("name", name),
	)
	return err
}

// Rename implements System.Rename.
func (s *DebugSystem) Rename(oldPath, newPath AbsPath) error {
	err := s.system.Rename(oldPath, newPath)
	chezmoilog.InfoOrError(s.logger, "RemoveAll", err,
		chezmoilog.Stringer("oldPath", oldPath),
		chezmoilog.Stringer("newPath", newPath),
	)
	return err
}

// RunCmd implements System.RunCmd.
func (s *DebugSystem) RunCmd(cmd *exec.Cmd) error {
	start := time.Now()
	err := s.system.RunCmd(cmd)
	attrs := []slog.Attr{
		slog.Any("cmd", chezmoilog.OSExecCmdLogValuer{Cmd: cmd}),
		slog.Duration("duration", time.Since(start)),
	}
	attrs = chezmoilog.AppendExitErrorAttrs(attrs, err)
	chezmoilog.InfoOrError(s.logger, "RunCmd", err, attrs...)
	return err
}

// RunScript implements System.RunScript.
func (s *DebugSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	err := s.system.RunScript(scriptName, dir, data, options)
	attrs := []slog.Attr{
		chezmoilog.Stringer("scriptName", scriptName),
		chezmoilog.Stringer("dir", dir),
		chezmoilog.FirstFewBytes("data", data),
		slog.Any("interpreter", options.Interpreter),
		slog.String("condition", string(options.Condition)),
	}
	attrs = chezmoilog.AppendExitErrorAttrs(attrs, err)
	chezmoilog.InfoOrError(s.logger, "RunScript", err, attrs...)
	return err
}

// Stat implements System.Stat.
func (s *DebugSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	fileInfo, err := s.system.Stat(name)
	chezmoilog.InfoOrError(s.logger, "Stat", err,
		chezmoilog.Stringer("name", name),
	)
	return fileInfo, err
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DebugSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *DebugSystem) WriteFile(name AbsPath, data []byte, perm fs.FileMode) error {
	err := s.system.WriteFile(name, data, perm)
	chezmoilog.InfoOrError(s.logger, "WriteFile", err,
		chezmoilog.Stringer("name", name),
		slog.Int("size", len(data)),
		chezmoilog.FirstFewBytes("data", data),
		slog.Int("perm", int(perm)),
	)
	return err
}

// WriteSymlink implements System.WriteSymlink.
func (s *DebugSystem) WriteSymlink(oldName string, newName AbsPath) error {
	err := s.system.WriteSymlink(oldName, newName)
	chezmoilog.InfoOrError(s.logger, "WriteSymlink", err,
		slog.String("oldName", oldName),
		chezmoilog.Stringer("newName", newName),
	)
	return err
}
