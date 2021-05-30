package chezmoi

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	vfs "github.com/twpayne/go-vfs/v3"
	"go.uber.org/multierr"
)

// An RealSystem is a System that writes to a filesystem and executes scripts.
type RealSystem struct {
	fileSystem  vfs.FS
	interpeters map[string]Interpreter
}

// NewRealSystem returns a System that acts on fs.
func NewRealSystem(fileSystem vfs.FS, interpreters map[string]Interpreter) *RealSystem {
	return &RealSystem{
		fileSystem:  fileSystem,
		interpeters: interpreters,
	}
}

// Chmod implements System.Chmod.
func (s *RealSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	return nil
}

// Readlink implements System.Readlink.
func (s *RealSystem) Readlink(name AbsPath) (string, error) {
	linkname, err := s.fileSystem.Readlink(string(name))
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(linkname), nil
}

// RunScript implements System.RunScript.
func (s *RealSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte) (err error) {
	// Write the temporary script file. Put the randomness at the front of the
	// filename to preserve any file extension for Windows scripts.
	f, err := os.CreateTemp("", "*."+scriptname.Base())
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, os.RemoveAll(f.Name()))
	}()

	_, err = f.Write(data)
	err = multierr.Append(err, f.Close())
	if err != nil {
		return
	}

	// By default, execute the script directly.
	//nolint:gosec
	cmd := exec.Command(f.Name())
	// Determine whether an interpreter is defined for the script's extension,
	// and, if so use it. This allows us to distinguish between scripts can be
	// executed directly by Windows (e.g. Batch scripts) and scripts which need
	// an interpreter (e.g. PowerShell scripts).
	ext := strings.ToLower(strings.TrimPrefix(scriptname.Ext(), "."))
	if interpreter, ok := s.interpeters[ext]; ok && interpreter.Command != "" {
		cmd = exec.Command(interpreter.Command, append(interpreter.Args, f.Name())...)
	}
	cmd.Dir, err = s.getScriptWorkingDir(dir)
	if err != nil {
		return
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = s.RunCmd(cmd)
	return
}

// WriteFile implements System.WriteFile.
func (s *RealSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	return s.fileSystem.WriteFile(string(filename), data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *RealSystem) WriteSymlink(oldname string, newname AbsPath) error {
	if err := s.fileSystem.RemoveAll(string(newname)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return s.fileSystem.Symlink(filepath.FromSlash(oldname), string(newname))
}
