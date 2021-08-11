package chezmoi

import (
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"runtime"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	vfs "github.com/twpayne/go-vfs/v3"
)

// A GitDiffSystem wraps a System and logs all of the actions executed as a git
// diff.
type GitDiffSystem struct {
	system         System
	dirAbsPath     AbsPath
	unifiedEncoder *diff.UnifiedEncoder
}

// NewGitDiffSystem returns a new GitDiffSystem. Output is written to w, the
// dirAbsPath is stripped from paths, and color controls whether the output
// contains ANSI color escape sequences.
func NewGitDiffSystem(system System, w io.Writer, dirAbsPath AbsPath, color bool) *GitDiffSystem {
	unifiedEncoder := diff.NewUnifiedEncoder(w, diff.DefaultContextLines)
	if color {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	return &GitDiffSystem{
		system:         system,
		dirAbsPath:     dirAbsPath,
		unifiedEncoder: unifiedEncoder,
	}
}

// Chmod implements System.Chmod.
func (s *GitDiffSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	fromInfo, err := s.system.Stat(name)
	if err != nil {
		return err
	}
	toMode := fromInfo.Mode().Type() | mode
	var toData []byte
	if fromInfo.Mode().IsRegular() {
		toData, err = s.ReadFile(name)
		if err != nil {
			return err
		}
	}
	if err := s.encodeDiff(name, toData, toMode); err != nil {
		return err
	}
	return s.system.Chmod(name, mode)
}

// Glob implements System.Glob.
func (s *GitDiffSystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *GitDiffSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdCombinedOutput(cmd)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *GitDiffSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdOutput(cmd)
}

// Lstat implements System.Lstat.
func (s *GitDiffSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *GitDiffSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	if err := s.encodeDiff(name, nil, fs.ModeDir|perm); err != nil {
		return err
	}
	return s.system.Mkdir(name, perm)
}

// RawPath implements System.RawPath.
func (s *GitDiffSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *GitDiffSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.system.ReadDir(name)
}

// ReadFile implements System.ReadFile.
func (s *GitDiffSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.system.ReadFile(name)
}

// Readlink implements System.Readlink.
func (s *GitDiffSystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// RemoveAll implements System.RemoveAll.
func (s *GitDiffSystem) RemoveAll(name AbsPath) error {
	if err := s.encodeDiff(name, nil, 0); err != nil {
		return err
	}
	return s.system.RemoveAll(name)
}

// Rename implements System.Rename.
func (s *GitDiffSystem) Rename(oldpath, newpath AbsPath) error {
	var fileMode filemode.FileMode
	var hash plumbing.Hash
	switch fromFileInfo, err := s.Stat(oldpath); {
	case err != nil:
		return err
	case fromFileInfo.Mode().IsDir():
		hash = plumbing.ZeroHash // LATER be more intelligent here
	case fromFileInfo.Mode().IsRegular():
		data, err := s.system.ReadFile(oldpath)
		if err != nil {
			return err
		}
		hash = plumbing.ComputeHash(plumbing.BlobObject, data)
	default:
		fileMode = filemode.FileMode(fromFileInfo.Mode())
	}
	if err := s.unifiedEncoder.Encode(&gitDiffPatch{
		filePatches: []diff.FilePatch{
			&gitDiffFilePatch{
				from: &gitDiffFile{
					fileMode: fileMode,
					relPath:  s.trimPrefix(oldpath),
					hash:     hash,
				},
				to: &gitDiffFile{
					fileMode: fileMode,
					relPath:  s.trimPrefix(newpath),
					hash:     hash,
				},
			},
		},
	}); err != nil {
		return err
	}
	return s.system.Rename(oldpath, newpath)
}

// RunCmd implements System.RunCmd.
func (s *GitDiffSystem) RunCmd(cmd *exec.Cmd) error {
	return s.system.RunCmd(cmd)
}

// RunIdempotentCmd implements System.RunIdempotentCmd.
func (s *GitDiffSystem) RunIdempotentCmd(cmd *exec.Cmd) error {
	return s.system.RunIdempotentCmd(cmd)
}

// RunScript implements System.RunScript.
func (s *GitDiffSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	mode := fs.FileMode(filemode.Executable)
	diffPatch, err := DiffPatch(scriptname, nil, mode, data, mode)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.RunScript(scriptname, dir, data, interpreter)
}

// Stat implements System.Stat.
func (s *GitDiffSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *GitDiffSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *GitDiffSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	if err := s.encodeDiff(filename, data, perm); err != nil {
		return err
	}
	return s.system.WriteFile(filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *GitDiffSystem) WriteSymlink(oldname string, newname AbsPath) error {
	toData := append([]byte(normalizeLinkname(oldname)), '\n')
	toMode := fs.ModeSymlink
	if runtime.GOOS == "windows" {
		toMode |= 0o666
	}
	if err := s.encodeDiff(newname, toData, toMode); err != nil {
		return err
	}
	return s.system.WriteSymlink(oldname, newname)
}

// encodeDiff encodes the diff between the actual state of absPath and the
// target state of toData and toMode.
func (s *GitDiffSystem) encodeDiff(absPath AbsPath, toData []byte, toMode fs.FileMode) error {
	var fromData []byte
	var fromMode fs.FileMode
	switch fromInfo, err := s.system.Lstat(absPath); {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return err
	case fromInfo.Mode().IsRegular():
		fromData, err = s.system.ReadFile(absPath)
		if err != nil {
			return err
		}
		fromMode = fromInfo.Mode()
	case fromInfo.Mode().Type() == fs.ModeSymlink:
		fromDataStr, err := s.system.Readlink(absPath)
		if err != nil {
			return err
		}
		fromData = append([]byte(fromDataStr), '\n')
		fromMode = fromInfo.Mode()
	default:
		fromMode = fromInfo.Mode()
	}

	diffPatch, err := DiffPatch(s.trimPrefix(absPath), fromData, fromMode, toData, toMode)
	if err != nil {
		return err
	}
	return s.unifiedEncoder.Encode(diffPatch)
}

// trimPrefix removes s's directory prefix from absPath.
func (s *GitDiffSystem) trimPrefix(absPath AbsPath) RelPath {
	return absPath.MustTrimDirPrefix(s.dirAbsPath)
}
