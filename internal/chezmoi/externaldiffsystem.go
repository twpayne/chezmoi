package chezmoi

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"

	vfs "github.com/twpayne/go-vfs/v3"
)

// An ExternalDiffSystem is a DiffSystem that uses an external diff tool.
type ExternalDiffSystem struct {
	system         System
	command        string
	args           []string
	destDirAbsPath AbsPath
	tempDirAbsPath AbsPath
}

// NewExternalDiffSystem creates a new ExternalDiffSystem.
func NewExternalDiffSystem(system System, command string, args []string, destDirAbsPath AbsPath) *ExternalDiffSystem {
	return &ExternalDiffSystem{
		system:         system,
		command:        command,
		args:           args,
		destDirAbsPath: destDirAbsPath,
	}
}

// Close frees all resources held by s.
func (s *ExternalDiffSystem) Close() error {
	if s.tempDirAbsPath != "" {
		if err := os.RemoveAll(string(s.tempDirAbsPath)); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		s.tempDirAbsPath = ""
	}
	return nil
}

// Chmod implements System.Chmod.
func (s *ExternalDiffSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	return s.system.Chmod(name, mode)
}

// Glob implements System.Glob.
func (s *ExternalDiffSystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *ExternalDiffSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdCombinedOutput(cmd)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *ExternalDiffSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return s.system.IdempotentCmdOutput(cmd)
}

// Lstat implements System.Lstat.
func (s *ExternalDiffSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *ExternalDiffSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	return s.system.Mkdir(name, perm)
}

// RawPath implements System.RawPath.
func (s *ExternalDiffSystem) RawPath(path AbsPath) (AbsPath, error) {
	return s.system.RawPath(path)
}

// ReadDir implements System.ReadDir.
func (s *ExternalDiffSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.system.ReadDir(name)
}

// ReadFile implements System.ReadFile.
func (s *ExternalDiffSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.system.ReadFile(name)
}

// Readlink implements System.Readlink.
func (s *ExternalDiffSystem) Readlink(name AbsPath) (string, error) {
	return s.system.Readlink(name)
}

// RemoveAll implements System.RemoveAll.
func (s *ExternalDiffSystem) RemoveAll(name AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.RemoveAll(name)
}

// Rename implements System.Rename.
func (s *ExternalDiffSystem) Rename(oldpath, newpath AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.Rename(oldpath, newpath)
}

// RunCmd implements System.RunCmd.
func (s *ExternalDiffSystem) RunCmd(cmd *exec.Cmd) error {
	return s.system.RunCmd(cmd)
}

// RunIdempotentCmd implements System.RunIdempotentCmd.
func (s *ExternalDiffSystem) RunIdempotentCmd(cmd *exec.Cmd) error {
	return s.system.RunIdempotentCmd(cmd)
}

// RunScript implements System.RunScript.
func (s *ExternalDiffSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	// FIXME generate suitable inputs for s.command
	return s.system.RunScript(scriptname, dir, data, interpreter)
}

// Stat implements System.Stat.
func (s *ExternalDiffSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Stat(name)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *ExternalDiffSystem) UnderlyingFS() vfs.FS {
	return s.system.UnderlyingFS()
}

// WriteFile implements System.WriteFile.
func (s *ExternalDiffSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	targetRelPath, err := filename.TrimDirPrefix(s.destDirAbsPath)
	if err != nil {
		return err
	}
	tempDirAbsPath, err := s.tempDir()
	if err != nil {
		return err
	}
	targetAbsPath := tempDirAbsPath.Join(targetRelPath)
	if err := os.MkdirAll(string(targetAbsPath.Dir()), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(string(targetAbsPath), data, perm); err != nil {
		return err
	}
	return s.runDiffCommand(filename, targetAbsPath)
}

// WriteSymlink implements System.WriteSymlink.
func (s *ExternalDiffSystem) WriteSymlink(oldname string, newname AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.WriteSymlink(oldname, newname)
}

// tempDir creates a temporary directory for s if it does not already exist and
// returns its path.
func (s *ExternalDiffSystem) tempDir() (AbsPath, error) {
	if s.tempDirAbsPath == "" {
		tempDir, err := os.MkdirTemp("", "chezmoi-diff")
		if err != nil {
			return "", err
		}
		s.tempDirAbsPath = AbsPath(tempDir)
	}
	return s.tempDirAbsPath, nil
}

// runDiffCommand runs the external diff command.
func (s *ExternalDiffSystem) runDiffCommand(destAbsPath, targetAbsPath AbsPath) error {
	templateData := struct {
		Destination string
		Target      string
	}{
		Destination: string(destAbsPath),
		Target:      string(targetAbsPath),
	}

	args := make([]string, 0, len(s.args))
	// Work around a regression introduced in 2.1.5
	// (https://github.com/twpayne/chezmoi/pull/1328) in a user-friendly way.
	//
	// Prior to #1328, the diff.args config option was prepended to the default
	// order of files to the diff command. Post #1328, the diff.args config
	// option replaced all arguments to the diff command.
	//
	// Work around this by looking for any templates in diff.args. An arg is
	// considered a template if, after execution as as template, it is not equal
	// to the original arg.
	anyTemplateArgs := false
	for i, arg := range s.args {
		tmpl, err := template.New("diff.args[" + strconv.Itoa(i) + "]").Parse(arg)
		if err != nil {
			return err
		}

		var sb strings.Builder
		if err := tmpl.Execute(&sb, templateData); err != nil {
			return err
		}
		args = append(args, sb.String())

		// Detect template arguments.
		if arg != sb.String() {
			anyTemplateArgs = true
		}
	}

	// If there are no template arguments, then append the destination and
	// target paths as prior to #1328.
	if !anyTemplateArgs {
		args = append(args, templateData.Destination, templateData.Target)
	}

	//nolint:gosec
	cmd := exec.Command(s.command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return s.system.RunIdempotentCmd(cmd)
}
