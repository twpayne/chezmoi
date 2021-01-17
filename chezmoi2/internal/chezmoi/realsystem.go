package chezmoi

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/rs/zerolog/log"
	vfs "github.com/twpayne/go-vfs"
	"go.uber.org/multierr"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoilog"
)

// Glob implements System.Glob.
func (s *RealSystem) Glob(pattern string) ([]string, error) {
	return doublestar.GlobOS(doubleStarOS{FS: s.UnderlyingFS()}, pattern)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *RealSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return chezmoilog.LogCmdOutput(log.Logger, cmd)
}

// Lstat implements System.Lstat.
func (s *RealSystem) Lstat(filename AbsPath) (os.FileInfo, error) {
	return s.fs.Lstat(string(filename))
}

// Mkdir implements System.Mkdir.
func (s *RealSystem) Mkdir(name AbsPath, perm os.FileMode) error {
	return s.fs.Mkdir(string(name), perm)
}

// PathSeparator implements doublestar.OS.PathSeparator.
func (s *RealSystem) PathSeparator() rune {
	return '/'
}

// RawPath implements System.RawPath.
func (s *RealSystem) RawPath(absPath AbsPath) (AbsPath, error) {
	rawAbsPath, err := s.fs.RawPath(string(absPath))
	if err != nil {
		return "", err
	}
	return AbsPath(rawAbsPath), nil
}

// ReadDir implements System.ReadDir.
func (s *RealSystem) ReadDir(dirname AbsPath) ([]os.FileInfo, error) {
	return s.fs.ReadDir(string(dirname))
}

// ReadFile implements System.ReadFile.
func (s *RealSystem) ReadFile(filename AbsPath) ([]byte, error) {
	return s.fs.ReadFile(string(filename))
}

// RemoveAll implements System.RemoveAll.
func (s *RealSystem) RemoveAll(name AbsPath) error {
	return s.fs.RemoveAll(string(name))
}

// Rename implements System.Rename.
func (s *RealSystem) Rename(oldpath, newpath AbsPath) error {
	return s.fs.Rename(string(oldpath), string(newpath))
}

// RunCmd implements System.RunCmd.
func (s *RealSystem) RunCmd(cmd *exec.Cmd) error {
	return chezmoilog.LogCmdRun(log.Logger, cmd)
}

// RunScript implements System.RunScript.
func (s *RealSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte) (err error) {
	// Write the temporary script file. Put the randomness at the front of the
	// filename to preserve any file extension for Windows scripts.
	f, err := ioutil.TempFile("", "*."+scriptname.Base())
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, os.RemoveAll(f.Name()))
	}()

	// Make the script private before writing it in case it contains any
	// secrets.
	if runtime.GOOS != "windows" {
		if err = f.Chmod(0o700); err != nil {
			return
		}
	}
	_, err = f.Write(data)
	err = multierr.Append(err, f.Close())
	if err != nil {
		return
	}

	// Run the temporary script file.
	//nolint:gosec
	cmd := exec.Command(f.Name())
	cmd.Dir = string(dir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = s.RunCmd(cmd)
	return
}

// Stat implements System.Stat.
func (s *RealSystem) Stat(name AbsPath) (os.FileInfo, error) {
	return s.fs.Stat(string(name))
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *RealSystem) UnderlyingFS() vfs.FS {
	return s.fs
}
