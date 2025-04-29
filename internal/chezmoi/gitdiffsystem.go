package chezmoi

import (
	"errors"
	"io"
	"io/fs"
	"os/exec"
	"runtime"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	vfs "github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiset"
)

// A GitDiffSystem wraps a System and logs all of the actions executed as a git
// diff.
type GitDiffSystem struct {
	system         System
	dirAbsPath     AbsPath
	filter         *EntryTypeFilter
	removedEntries chezmoiset.Set[AbsPath]
	reverse        bool
	scriptContents bool
	textConvFunc   TextConvFunc
	unifiedEncoder *diff.UnifiedEncoder
}

// GitDiffSystemOptions are options for NewGitDiffSystem.
type GitDiffSystemOptions struct {
	Color          bool
	Filter         *EntryTypeFilter
	Reverse        bool
	ScriptContents bool
	TextConvFunc   TextConvFunc
}

// NewGitDiffSystem returns a new GitDiffSystem. Output is written to w, the
// dirAbsPath is stripped from paths, and color controls whether the output
// contains ANSI color escape sequences.
func NewGitDiffSystem(system System, w io.Writer, dirAbsPath AbsPath, options *GitDiffSystemOptions) *GitDiffSystem {
	unifiedEncoder := diff.NewUnifiedEncoder(w, diff.DefaultContextLines)
	if options.Color {
		unifiedEncoder.SetColor(diff.NewColorConfig())
	}
	return &GitDiffSystem{
		system:         system,
		dirAbsPath:     dirAbsPath,
		filter:         options.Filter,
		removedEntries: chezmoiset.New[AbsPath](),
		reverse:        options.Reverse,
		scriptContents: options.ScriptContents,
		textConvFunc:   options.TextConvFunc,
		unifiedEncoder: unifiedEncoder,
	}
}

// Chmod implements System.Chmod.
func (s *GitDiffSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	fromInfo, err := s.system.Stat(name)
	if err != nil {
		return err
	}
	if s.filter.IncludeFileInfo(fromInfo) {
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
	}
	return s.system.Chmod(name, mode)
}

// Chtimes implements system.Chtimes.
func (s *GitDiffSystem) Chtimes(name AbsPath, atime, mtime time.Time) error {
	if s.isRemoved(name) {
		return fs.ErrNotExist
	}
	return s.system.Chtimes(name, atime, mtime)
}

// Glob implements System.Glob.
func (s *GitDiffSystem) Glob(pattern string) ([]string, error) {
	matches, err := s.system.Glob(pattern)
	if err != nil {
		return nil, err
	}
	n := 0
	for _, match := range matches {
		if s.isRemoved(NewAbsPath(match)) {
			continue
		}
		matches[n] = match
		n++
	}
	return matches[:n], nil
}

// Link implements System.Link.
func (s *GitDiffSystem) Link(oldName, newName AbsPath) error {
	// LATER generate a diff
	return s.system.Link(oldName, newName)
}

// Lstat implements System.Lstat.
func (s *GitDiffSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	if s.isRemoved(name) {
		return nil, fs.ErrNotExist
	}
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *GitDiffSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeDirs) {
		if err := s.encodeDiff(name, nil, fs.ModeDir|perm); err != nil {
			return err
		}
	}
	return s.system.Mkdir(name, perm)
}

// RawPath implements System.RawPath.
func (s *GitDiffSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *GitDiffSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	if s.isRemoved(name) {
		return nil, fs.ErrNotExist
	}
	dirEntries, err := s.system.ReadDir(name)
	if err != nil {
		return nil, err
	}
	n := 0
	for _, dirEntry := range dirEntries {
		if s.isRemoved(name.JoinString(dirEntry.Name())) {
			continue
		}
		dirEntries[n] = dirEntry
		n++
	}
	return dirEntries[:n], nil
}

// ReadFile implements System.ReadFile.
func (s *GitDiffSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.system.ReadFile(name)
}

// Readlink implements System.Readlink.
func (s *GitDiffSystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// Remove implements System.Remove.
func (s *GitDiffSystem) Remove(name AbsPath) error {
	// Only emit diffs for removing directories if the underlying directory is
	// empty. Keep track of removed entries to handled nested removes.
	switch fileInfo, err := s.system.Stat(name); {
	case err != nil:
		return err
	case fileInfo.IsDir():
		switch dirEntries, err := s.ReadDir(name); {
		case err != nil:
			return err
		case len(dirEntries) != 0:
			return fs.ErrExist
		}
	}
	if s.filter.IncludeEntryTypeBits(EntryTypeRemove) {
		if err := s.encodeDiff(name, nil, 0); err != nil {
			return err
		}
	}
	if err := s.system.Remove(name); err != nil {
		return err
	}
	s.removedEntries.Add(name)
	return nil
}

// RemoveAll implements System.RemoveAll.
func (s *GitDiffSystem) RemoveAll(name AbsPath) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeRemove) {
		if err := s.encodeDiff(name, nil, 0); err != nil {
			return err
		}
	}
	if err := s.system.RemoveAll(name); err != nil {
		return err
	}
	s.removedEntries.Add(name)
	return nil
}

// Rename implements System.Rename.
func (s *GitDiffSystem) Rename(oldPath, newPath AbsPath) error {
	fromFileInfo, err := s.Stat(oldPath)
	if err != nil {
		return err
	}
	if s.filter.IncludeFileInfo(fromFileInfo) {
		var fileMode filemode.FileMode
		var hash plumbing.Hash
		switch {
		case fromFileInfo.Mode().IsDir():
			hash = plumbing.ZeroHash // LATER be more intelligent here
		case fromFileInfo.Mode().IsRegular():
			data, err := s.system.ReadFile(oldPath)
			if err != nil {
				return err
			}
			hash = plumbing.ComputeHash(plumbing.BlobObject, data)
		default:
			fileMode = filemode.FileMode(fromFileInfo.Mode())
		}
		fromPath, toPath := s.trimPrefix(oldPath), s.trimPrefix(newPath)
		if s.reverse {
			fromPath, toPath = toPath, fromPath
		}
		if err := s.unifiedEncoder.Encode(&gitDiffPatch{
			filePatches: []diff.FilePatch{
				&gitDiffFilePatch{
					from: &gitDiffFile{
						fileMode: fileMode,
						relPath:  fromPath,
						hash:     hash,
					},
					to: &gitDiffFile{
						fileMode: fileMode,
						relPath:  toPath,
						hash:     hash,
					},
				},
			},
		}); err != nil {
			return err
		}
	}
	if err := s.system.Rename(oldPath, newPath); err != nil {
		return err
	}
	s.removedEntries.Add(oldPath)
	return nil
}

// RunCmd implements System.RunCmd.
func (s *GitDiffSystem) RunCmd(cmd *exec.Cmd) error {
	return s.system.RunCmd(cmd)
}

// RunScript implements System.RunScript.
func (s *GitDiffSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	bits := EntryTypeScripts
	if options.Condition == ScriptConditionAlways {
		bits |= EntryTypeAlways
	}
	if s.filter.IncludeEntryTypeBits(bits) {
		fromData, toData := []byte(nil), data
		fromMode, toMode := fs.FileMode(0), fs.FileMode(filemode.Executable)
		if !s.scriptContents {
			toData = nil
		}
		if s.reverse {
			fromData, toData = toData, fromData
			fromMode, toMode = toMode, fromMode
		}
		diffPatch, err := DiffPatch(scriptName, fromData, fromMode, toData, toMode)
		if err != nil {
			return err
		}
		if err := s.unifiedEncoder.Encode(diffPatch); err != nil {
			return err
		}
	}
	return s.system.RunScript(scriptName, dir, data, options)
}

// Stat implements System.Stat.
func (s *GitDiffSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	if s.isRemoved(name) {
		return nil, fs.ErrNotExist
	}
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *GitDiffSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *GitDiffSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeFiles) {
		if err := s.encodeDiff(filename, data, perm); err != nil {
			return err
		}
	}
	return s.system.WriteFile(filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *GitDiffSystem) WriteSymlink(oldName string, newName AbsPath) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeSymlinks) {
		toData := append([]byte(normalizeLinkname(oldName)), '\n')
		toMode := fs.ModeSymlink
		if runtime.GOOS == "windows" {
			toMode |= 0o666
		}
		if err := s.encodeDiff(newName, toData, toMode); err != nil {
			return err
		}
	}
	return s.system.WriteSymlink(oldName, newName)
}

// encodeDiff encodes the diff between the actual state of absPath and the
// target state of toData and toMode.
func (s *GitDiffSystem) encodeDiff(absPath AbsPath, toData []byte, toMode fs.FileMode) error {
	var fromData []byte
	var fromMode fs.FileMode
	switch fromInfo, err := s.system.Lstat(absPath); {
	case errors.Is(err, fs.ErrNotExist):
		// Leave fromData and fromMode at their zero values.
	case err != nil:
		return err
	case fromInfo.Mode().IsRegular():
		fromData, err = s.system.ReadFile(absPath)
		if err != nil {
			return err
		}
		if s.textConvFunc != nil {
			fromData, _, err = s.textConvFunc(absPath.String(), fromData)
			if err != nil {
				return err
			}
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

	if s.textConvFunc != nil {
		var err error
		toData, _, err = s.textConvFunc(absPath.String(), toData)
		if err != nil {
			return err
		}
	}

	if s.reverse {
		fromData, toData = toData, fromData
		fromMode, toMode = toMode, fromMode
	}

	diffPatch, err := DiffPatch(s.trimPrefix(absPath), fromData, fromMode, toData, toMode)
	if err != nil {
		return err
	}

	return s.unifiedEncoder.Encode(diffPatch)
}

func (s *GitDiffSystem) isRemoved(absPath AbsPath) bool {
	if s.removedEntries.IsEmpty() {
		return false
	}
	if s.removedEntries.Contains(absPath) {
		return true
	}
	var lastPrefixAbsPath AbsPath
	prefixAbsPath, _ := absPath.Split()
	for prefixAbsPath != lastPrefixAbsPath {
		if s.removedEntries.Contains(prefixAbsPath) {
			return true
		}
		lastPrefixAbsPath = prefixAbsPath
		prefixAbsPath, _ = prefixAbsPath.Split()
	}
	return false
}

// trimPrefix removes s's directory prefix from absPath.
func (s *GitDiffSystem) trimPrefix(absPath AbsPath) RelPath {
	return absPath.MustTrimDirPrefix(s.dirAbsPath)
}
