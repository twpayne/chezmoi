package chezmoi

import (
	"io"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	vfs "github.com/twpayne/go-vfs"
)

// A GitDiffSystem wraps a System and logs all of the actions executed as a git
// diff.
type GitDiffSystem struct {
	system         System
	dirAbsPath     AbsPath
	unifiedEncoder *diff.UnifiedEncoder
}

// NewGitDiffSystem returns a new GitDiffSystem.
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
func (s *GitDiffSystem) Chmod(name AbsPath, mode os.FileMode) error {
	var fromData, toData []byte
	fromInfo, err := s.system.Stat(name)
	switch {
	case err != nil:
		return err
	case fromInfo.IsDir():
		mode |= os.ModeDir
	default:
		data, err := s.ReadFile(name)
		if err != nil {
			return err
		}
		fromData = data
		toData = data
	}
	diffPatch, err := DiffPatch(s.trimPrefix(name), fromData, fromInfo.Mode(), toData, mode)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.Chmod(name, mode)
}

// Glob implements System.Glob.
func (s *GitDiffSystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *GitDiffSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdOutput(cmd)
}

// Lstat implements System.Lstat.
func (s *GitDiffSystem) Lstat(name AbsPath) (os.FileInfo, error) {
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *GitDiffSystem) Mkdir(name AbsPath, perm os.FileMode) error {
	diffPatch, err := DiffPatch(s.trimPrefix(name), nil, 0, nil, os.ModeDir|perm)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.Mkdir(name, perm)
}

// RawPath implements System.RawPath.
func (s *GitDiffSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *GitDiffSystem) ReadDir(dirname AbsPath) ([]os.FileInfo, error) {
	return s.system.ReadDir(dirname)
}

// ReadFile implements System.ReadFile.
func (s *GitDiffSystem) ReadFile(filename AbsPath) ([]byte, error) {
	return s.system.ReadFile(filename)
}

// Readlink implements System.Readlink.
func (s *GitDiffSystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// RemoveAll implements System.RemoveAll.
func (s *GitDiffSystem) RemoveAll(name AbsPath) error {
	var fromData []byte
	var fromMode os.FileMode
	switch fromInfo, err := s.system.Stat(name); {
	case err == nil:
		fromMode = fromInfo.Mode()
		if !fromInfo.IsDir() {
			fromData, err = s.system.ReadFile(name)
			if err != nil {
				return err
			}
		}
	case os.IsNotExist(err):
	default:
		return err
	}
	diffPatch, err := DiffPatch(s.trimPrefix(name), fromData, fromMode, nil, 0)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
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

// RunScript implements System.RunScript.
func (s *GitDiffSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte) error {
	mode := os.FileMode(filemode.Executable)
	diffPatch, err := DiffPatch(s.trimPrefix(AbsPath(scriptname)), nil, mode, data, mode)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.RunScript(scriptname, dir, data)
}

// Stat implements System.Stat.
func (s *GitDiffSystem) Stat(name AbsPath) (os.FileInfo, error) {
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *GitDiffSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *GitDiffSystem) WriteFile(filename AbsPath, data []byte, perm os.FileMode) error {
	var fromData []byte
	var fromMode os.FileMode
	switch fromInfo, err := s.system.Stat(filename); {
	case err == nil:
		fromData, err = s.system.ReadFile(filename)
		if err != nil {
			return err
		}
		fromMode = fromInfo.Mode()
	case os.IsNotExist(err):
	default:
		return err
	}
	diffPatch, err := DiffPatch(s.trimPrefix(filename), fromData, fromMode, data, perm)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.WriteFile(filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *GitDiffSystem) WriteSymlink(oldname string, newname AbsPath) error {
	var fromMode os.FileMode
	var fromData []byte
	switch fromInfo, err := s.system.Stat(newname); {
	case os.IsNotExist(err):
	case err != nil:
		return err
	default:
		fromLinkname, err := s.system.Readlink(newname)
		if err != nil {
			return err
		}
		fromMode = fromInfo.Mode()
		fromData = []byte(fromLinkname)
	}

	diffPatch, err := DiffPatch(s.trimPrefix(newname), fromData, fromMode, []byte(oldname), os.ModeSymlink)
	if err != nil {
		return err
	}
	if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
		return err
	}
	return s.system.WriteSymlink(oldname, newname)
}

func (s *GitDiffSystem) trimPrefix(absPath AbsPath) RelPath {
	return absPath.MustTrimDirPrefix(s.dirAbsPath)
}
