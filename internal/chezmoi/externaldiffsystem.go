package chezmoi

import (
	"bytes"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	vfs "github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/internal/chezmoilog"
)

// An ExternalDiffSystem is a DiffSystem that uses an external diff tool.
type ExternalDiffSystem struct {
	system         System
	command        string
	args           []string
	destDirAbsPath AbsPath
	tempDirAbsPath AbsPath
	filter         *EntryTypeFilter
	reverse        bool
	scriptContents bool
	textConvFunc   TextConvFunc
}

// ExternalDiffSystemOptions are options for NewExternalDiffSystem.
type ExternalDiffSystemOptions struct {
	Filter         *EntryTypeFilter
	Reverse        bool
	ScriptContents bool
	TextConvFunc   TextConvFunc
}

// NewExternalDiffSystem creates a new ExternalDiffSystem.
func NewExternalDiffSystem(
	system System,
	command string,
	args []string,
	destDirAbsPath AbsPath,
	options *ExternalDiffSystemOptions,
) *ExternalDiffSystem {
	return &ExternalDiffSystem{
		system:         system,
		command:        command,
		args:           args,
		destDirAbsPath: destDirAbsPath,
		filter:         options.Filter,
		reverse:        options.Reverse,
		scriptContents: options.ScriptContents,
		textConvFunc:   options.TextConvFunc,
	}
}

// Close frees all resources held by s.
func (s *ExternalDiffSystem) Close() error {
	if !s.tempDirAbsPath.IsEmpty() {
		if err := os.RemoveAll(s.tempDirAbsPath.String()); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		s.tempDirAbsPath = EmptyAbsPath
	}
	return nil
}

// Chmod implements System.Chmod.
func (s *ExternalDiffSystem) Chmod(name AbsPath, mode fs.FileMode) error {
	// FIXME generate suitable inputs for s.command
	return s.system.Chmod(name, mode)
}

// Chtimes implements System.Chtimes.
func (s *ExternalDiffSystem) Chtimes(name AbsPath, atime, mtime time.Time) error {
	return s.system.Chtimes(name, atime, mtime)
}

// Glob implements System.Glob.
func (s *ExternalDiffSystem) Glob(pattern string) ([]string, error) {
	return s.system.Glob(pattern)
}

// Link implements System.Link.
func (s *ExternalDiffSystem) Link(oldName, newName AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.Link(oldName, newName)
}

// Lstat implements System.Lstat.
func (s *ExternalDiffSystem) Lstat(name AbsPath) (fs.FileInfo, error) {
	return s.system.Lstat(name)
}

// Mkdir implements System.Mkdir.
func (s *ExternalDiffSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeDirs) {
		targetRelPath, err := name.TrimDirPrefix(s.destDirAbsPath)
		if err != nil {
			return err
		}
		tempDirAbsPath, err := s.tempDir()
		if err != nil {
			return err
		}
		targetAbsPath := tempDirAbsPath.Join(targetRelPath)
		if err := os.MkdirAll(targetAbsPath.String(), perm); err != nil {
			return err
		}
		if err := s.runDiffCommand(devNullAbsPath, targetAbsPath); err != nil {
			return err
		}
	}
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

// Remove implements System.Remove.
func (s *ExternalDiffSystem) Remove(name AbsPath) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeRemove) {
		switch fileInfo, err := s.system.Lstat(name); {
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		case err != nil:
			return err
		case s.filter.IncludeFileInfo(fileInfo):
			if err := s.runDiffCommand(name, devNullAbsPath); err != nil {
				return err
			}
		}
	}
	return s.system.Remove(name)
}

// RemoveAll implements System.RemoveAll.
func (s *ExternalDiffSystem) RemoveAll(name AbsPath) error {
	if s.filter.IncludeEntryTypeBits(EntryTypeRemove) {
		switch fileInfo, err := s.system.Lstat(name); {
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		case err != nil:
			return err
		case s.filter.IncludeFileInfo(fileInfo):
			if err := s.runDiffCommand(name, devNullAbsPath); err != nil {
				return err
			}
		}
	}
	return s.system.RemoveAll(name)
}

// Rename implements System.Rename.
func (s *ExternalDiffSystem) Rename(oldPath, newPath AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.Rename(oldPath, newPath)
}

// RunCmd implements System.RunCmd.
func (s *ExternalDiffSystem) RunCmd(cmd *exec.Cmd) error {
	return s.system.RunCmd(cmd)
}

// RunScript implements System.RunScript.
func (s *ExternalDiffSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	bits := EntryTypeScripts
	if options.Condition == ScriptConditionAlways {
		bits |= EntryTypeAlways
	}
	if s.filter.IncludeEntryTypeBits(bits) {
		tempDirAbsPath, err := s.tempDir()
		if err != nil {
			return err
		}
		targetAbsPath := tempDirAbsPath.Join(scriptName)
		if err := os.MkdirAll(targetAbsPath.Dir().String(), 0o700); err != nil {
			return err
		}
		toData := data
		if !s.scriptContents {
			toData = nil
		}
		if err := os.WriteFile(targetAbsPath.String(), toData, 0o700); err != nil {
			return err
		}
		if err := s.runDiffCommand(devNullAbsPath, targetAbsPath); err != nil {
			return err
		}
	}
	return s.system.RunScript(scriptName, dir, data, options)
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
	if s.filter.IncludeEntryTypeBits(EntryTypeFiles) {
		targetRelPath, err := filename.TrimDirPrefix(s.destDirAbsPath)
		if err != nil {
			return err
		}
		tempDirAbsPath, err := s.tempDir()
		if err != nil {
			return err
		}

		// If filename does not exist, replace it with /dev/null to avoid
		// passing the name of a non-existent file to the external diff command.
		// Otherwise, if the file exists and a textconv filter is configured,
		// run the filter and update fromAbsPath to point to the converted data.
		fromAbsPath := filename
		switch fileInfo, err := os.Lstat(fromAbsPath.String()); {
		case errors.Is(err, fs.ErrNotExist):
			fromAbsPath = devNullAbsPath
		case err != nil:
			return err
		case s.textConvFunc != nil:
			// Maybe convert the from data with textconv.
			fromData, err := os.ReadFile(fromAbsPath.String())
			if err != nil {
				return err
			}
			switch convertedFromData, converted, err := s.textConvFunc(fromAbsPath.String(), fromData); {
			case err != nil:
				return err
			case converted:
				tempFromAbsPath := tempDirAbsPath.Join(NewRelPath("a"), targetRelPath)
				if err := os.MkdirAll(tempFromAbsPath.Dir().String(), 0o700); err != nil {
					return err
				}
				if err := os.WriteFile(tempFromAbsPath.String(), convertedFromData, fileInfo.Mode().Perm()); err != nil {
					return err
				}
				fromAbsPath = tempFromAbsPath
			}
		}

		// Write the target contents to a file in a temporary directory.
		toAbsPath := tempDirAbsPath.Join(targetRelPath)
		toData := data
		if s.textConvFunc != nil {
			// Maybe convert the to data with textconv.
			switch convertedToData, converted, err := s.textConvFunc(filename.String(), toData); {
			case err != nil:
				return err
			case converted:
				toAbsPath = tempDirAbsPath.Join(NewRelPath("b"), targetRelPath)
				toData = convertedToData
			}
		}
		if err := os.MkdirAll(toAbsPath.Dir().String(), 0o700); err != nil {
			return err
		}
		if err := os.WriteFile(toAbsPath.String(), toData, perm); err != nil {
			return err
		}

		// Run the external diff command.
		if err := s.runDiffCommand(fromAbsPath, toAbsPath); err != nil {
			return err
		}
	}
	return s.system.WriteFile(filename, data, perm)
}

// WriteSymlink implements System.WriteSymlink.
func (s *ExternalDiffSystem) WriteSymlink(oldName string, newName AbsPath) error {
	// FIXME generate suitable inputs for s.command
	return s.system.WriteSymlink(oldName, newName)
}

// tempDir creates a temporary directory for s if it does not already exist and
// returns its path.
func (s *ExternalDiffSystem) tempDir() (AbsPath, error) {
	if s.tempDirAbsPath.IsEmpty() {
		tempDir, err := os.MkdirTemp("", "chezmoi-diff")
		if err != nil {
			return EmptyAbsPath, err
		}
		s.tempDirAbsPath = NewAbsPath(tempDir)
	}
	return s.tempDirAbsPath, nil
}

// runDiffCommand runs the external diff command.
func (s *ExternalDiffSystem) runDiffCommand(destAbsPath, targetAbsPath AbsPath) error {
	templateData := struct {
		Destination string
		Target      string
	}{
		Destination: destAbsPath.String(),
		Target:      targetAbsPath.String(),
	}

	if s.reverse {
		templateData.Destination, templateData.Target = templateData.Target, templateData.Destination
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
	// considered a template if, after execution as a template, it is not equal
	// to the original arg.
	anyTemplateArgs := false
	for i, arg := range s.args {
		tmpl, err := template.New("diff.args[" + strconv.Itoa(i) + "]").Parse(arg)
		if err != nil {
			return err
		}

		builder := strings.Builder{}
		if err := tmpl.Execute(&builder, templateData); err != nil {
			return err
		}
		args = append(args, builder.String())

		// Detect template arguments.
		if arg != builder.String() {
			anyTemplateArgs = true
		}
	}

	// If there are no template arguments, then append the destination and
	// target paths as prior to #1328.
	if !anyTemplateArgs {
		args = append(args, templateData.Destination, templateData.Target)
	}

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := chezmoilog.LogCmdRun(slog.Default(), cmd)

	// Swallow exit status 1 errors if the files differ as diff commands
	// traditionally exit with code 1 in this case.
	if exitError := (&exec.ExitError{}); errors.As(err, &exitError) && exitError.ExitCode() == 1 {
		destData, err2 := s.ReadFile(destAbsPath)
		switch {
		case errors.Is(err2, fs.ErrNotExist):
			// Do nothing.
		case err2 != nil:
			return errors.Join(err, err2)
		}
		targetData, err2 := s.ReadFile(targetAbsPath)
		switch {
		case errors.Is(err2, fs.ErrNotExist):
			// Do nothing.
		case err2 != nil:
			return errors.Join(err, err2)
		}
		if !bytes.Equal(destData, targetData) {
			return nil
		}
	}
	return err
}
