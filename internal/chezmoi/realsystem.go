package chezmoi

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	vfs "github.com/twpayne/go-vfs/v5"

	"chezmoi.io/chezmoi/v2/internal/chezmoierrors"
	"chezmoi.io/chezmoi/v2/internal/chezmoilog"
)

// A RealSystemOption sets an option on a RealSystem.
type RealSystemOption func(*RealSystem)

// Chtimes implements System.Chtimes.
func (s *RealSystem) Chtimes(name AbsPath, atime, mtime time.Time) error {
	return s.fileSystem.Chtimes(name.String(), atime, mtime)
}

// Glob implements System.Glob.
func (s *RealSystem) Glob(pattern string) ([]string, error) {
	return Glob(s.UnderlyingFS(), filepath.ToSlash(pattern))
}

// Link implements System.Link.
func (s *RealSystem) Link(oldName, newName AbsPath) error {
	return s.fileSystem.Link(oldName.String(), newName.String())
}

// Lstat implements System.Lstat.
func (s *RealSystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Lstat(filename.String())
}

// Mkdir implements System.Mkdir.
func (s *RealSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	return s.fileSystem.Mkdir(name.String(), perm)
}

// RawPath implements System.RawPath.
func (s *RealSystem) RawPath(absPath AbsPath) (AbsPath, error) {
	rawAbsPath, err := s.fileSystem.RawPath(absPath.String())
	if err != nil {
		return EmptyAbsPath, err
	}
	return NewAbsPath(rawAbsPath), nil
}

// ReadDir implements System.ReadDir.
func (s *RealSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.fileSystem.ReadDir(name.String())
}

// ReadFile implements System.ReadFile.
func (s *RealSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.fileSystem.ReadFile(name.String())
}

// Remove implements System.Remove.
func (s *RealSystem) Remove(name AbsPath) error {
	return s.fileSystem.Remove(name.String())
}

// RemoveAll implements System.RemoveAll.
func (s *RealSystem) RemoveAll(name AbsPath) error {
	return s.fileSystem.RemoveAll(name.String())
}

// Rename implements System.Rename.
func (s *RealSystem) Rename(oldPath, newPath AbsPath) error {
	return s.fileSystem.Rename(oldPath.String(), newPath.String())
}

// RunCmd implements System.RunCmd.
func (s *RealSystem) RunCmd(cmd *exec.Cmd) error {
	return chezmoilog.LogCmdRun(slog.Default(), cmd)
}

type runScriptArgs struct {
	scriptName    string
	workingDir    AbsPath
	setWorkingDir bool
	data          []byte
	interpreter   *Interpreter
	sourceRelPath SourceRelPath
}

type runScriptState struct {
	createScriptTempDirOnce *sync.Once
	scriptTempDir           AbsPath
	system                  System
}

type preparedScriptCmd struct {
	cmd     *exec.Cmd
	cleanup func() error
}

func prepareScriptCmd(args runScriptArgs, state runScriptState) (_ preparedScriptCmd, err error) {
	// Create the script temporary directory, if needed.
	state.createScriptTempDirOnce.Do(func() {
		if !state.scriptTempDir.IsEmpty() {
			err = os.MkdirAll(state.scriptTempDir.String(), 0o700)
		}
	})
	if err != nil {
		return preparedScriptCmd{}, err
	}

	// Write the temporary script file. Put the randomness at the front of the
	// filename to preserve any file extension for Windows scripts.
	var f *os.File
	f, err = os.CreateTemp(state.scriptTempDir.String(), "*."+args.scriptName)
	if err != nil {
		return preparedScriptCmd{}, err
	}

	// Remove the temporary script file when we are done using it.
	cleanup := func() error {
		return os.RemoveAll(f.Name())
	}
	// If preparing the command fails after creating the temp file, clean it up here.
	// On success, cleanup is returned to the caller to run after the command exits.
	defer func() {
		if err != nil {
			chezmoierrors.CombineFunc(&err, cleanup)
		}
	}()

	// Make the script private before writing it in case it contains any
	// secrets.
	if runtime.GOOS != "windows" {
		if err := f.Chmod(0o700); err != nil {
			return preparedScriptCmd{}, err
		}
	}
	_, err = f.Write(args.data)
	err = chezmoierrors.Combine(err, f.Close())
	if err != nil {
		return preparedScriptCmd{}, err
	}

	cmd := args.interpreter.ExecCommand(f.Name())

	if args.setWorkingDir {
		cmd.Dir, err = getScriptWorkingDir(state.system, args.workingDir)
		if err != nil {
			return preparedScriptCmd{}, err
		}
	}

	cmd.Env = append(os.Environ(),
		"CHEZMOI_SOURCE_FILE="+args.sourceRelPath.String(),
	)

	return preparedScriptCmd{
		cmd:     cmd,
		cleanup: cleanup,
	}, nil
}

// RunScript implements System.RunScript.
func (s *RealSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) (err error) {
	args := runScriptArgs{
		scriptName:    scriptName.Base(),
		workingDir:    dir,
		setWorkingDir: true,
		data:          data,
		interpreter:   options.Interpreter,
		sourceRelPath: options.SourceRelPath,
	}

	state := runScriptState{
		createScriptTempDirOnce: &s.createScriptTempDirOnce,
		scriptTempDir:           s.scriptTempDir,
		system:                  s,
	}

	preparedScript, err := prepareScriptCmd(args, state)
	if err != nil {
		return err
	}
	defer chezmoierrors.CombineFunc(&err, preparedScript.cleanup)

	preparedScript.cmd.Stdin = os.Stdin
	preparedScript.cmd.Stdout = os.Stdout
	preparedScript.cmd.Stderr = os.Stderr

	return s.RunCmd(preparedScript.cmd)
}

// Stat implements System.Stat.
func (s *RealSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Stat(name.String())
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *RealSystem) UnderlyingFS() vfs.FS {
	return s.fileSystem
}

// getScriptWorkingDir returns the script's working directory.
//
// If this is a before_ script then the requested working directory may not
// actually exist yet, so search through the parent directory hierarchy until
// we find a suitable working directory.
func getScriptWorkingDir(s System, dir AbsPath) (string, error) {
	// This should always terminate because dir will eventually become ".", i.e.
	// the current directory.
	for {
		switch fileInfo, err := s.Stat(dir); {
		case err == nil && fileInfo.IsDir():
			// dir exists and is a directory. Use it.
			dirRawAbsPath, err := s.RawPath(dir)
			if err != nil {
				return "", err
			}
			return dirRawAbsPath.String(), nil
		case err == nil || errors.Is(err, fs.ErrNotExist):
			// Either dir does not exist, or it exists and is not a directory.
			dir = dir.Dir()
		default:
			// Some other error occurred.
			return "", err
		}
	}
}
