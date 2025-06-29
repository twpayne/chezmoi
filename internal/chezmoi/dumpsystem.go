package chezmoi

import (
	"io/fs"
	"os/exec"

	vfs "github.com/twpayne/go-vfs/v5"
)

// A dataType is a data type.
type dataType string

// dataTypes.
const (
	dataTypeCommand dataType = "command"
	dataTypeDir     dataType = "dir"
	dataTypeFile    dataType = "file"
	dataTypeScript  dataType = "script"
	dataTypeSymlink dataType = "symlink"
)

// A DumpSystem is a System that writes to a data file.
type DumpSystem struct {
	emptySystemMixin
	noUpdateSystemMixin

	data map[string]any
}

// A commandData contains data about a command.
type commandData struct {
	Type dataType `json:"type" yaml:"type"`
	Path string   `json:"path" yaml:"path"`
	Args []string `json:"args" yaml:"args"`
}

// A dirData contains data about a directory.
type dirData struct {
	Type dataType    `json:"type" yaml:"type"`
	Name AbsPath     `json:"name" yaml:"name"`
	Perm fs.FileMode `json:"perm" yaml:"perm"`
}

// A fileData contains data about a file.
type fileData struct {
	Type     dataType    `json:"type"     yaml:"type"`
	Name     AbsPath     `json:"name"     yaml:"name"`
	Contents string      `json:"contents" yaml:"contents"`
	Perm     fs.FileMode `json:"perm"     yaml:"perm"`
}

// A scriptData contains data about a script.
type scriptData struct {
	Type        dataType     `json:"type"                  yaml:"type"`
	Name        AbsPath      `json:"name"                  yaml:"name"`
	Contents    string       `json:"contents"              yaml:"contents"`
	Condition   string       `json:"condition"             yaml:"condition"`
	Interpreter *Interpreter `json:"interpreter,omitempty" yaml:"interpreter,omitempty"`
}

// A symlinkData contains data about a symlink.
type symlinkData struct {
	Type     dataType `json:"type"     yaml:"type"`
	Name     AbsPath  `json:"name"     yaml:"name"`
	Linkname string   `json:"linkname" yaml:"linkname"`
}

// NewDumpSystem returns a new DumpSystem that accumulates data.
func NewDumpSystem() *DumpSystem {
	return &DumpSystem{
		data: make(map[string]any),
	}
}

// Data returns s's data.
func (s *DumpSystem) Data() any {
	return s.data
}

// Mkdir implements System.Mkdir.
func (s *DumpSystem) Mkdir(dirname AbsPath, perm fs.FileMode) error {
	return s.setData(dirname.String(), &dirData{
		Type: dataTypeDir,
		Name: dirname,
		Perm: perm,
	})
}

// RunCmd implements System.RunCmd.
func (s *DumpSystem) RunCmd(cmd *exec.Cmd) error {
	if cmd.Dir == "" {
		return nil
	}
	return s.setData(cmd.Dir, &commandData{
		Type: dataTypeCommand,
		Path: cmd.Path,
		Args: cmd.Args,
	})
}

// RunScript implements System.RunScript.
func (s *DumpSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	scriptNameStr := scriptName.String()
	scriptData := &scriptData{
		Type:     dataTypeScript,
		Name:     NewAbsPath(scriptNameStr),
		Contents: string(data),
	}
	if options.Condition != ScriptConditionNone {
		scriptData.Condition = string(options.Condition)
	}
	if !options.Interpreter.None() {
		scriptData.Interpreter = options.Interpreter
	}
	return s.setData(scriptNameStr, scriptData)
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DumpSystem) UnderlyingFS() vfs.FS {
	return nil
}

// WriteFile implements System.WriteFile.
func (s *DumpSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	return s.setData(filename.String(), &fileData{
		Type:     dataTypeFile,
		Name:     filename,
		Contents: string(data),
		Perm:     perm,
	})
}

// WriteSymlink implements System.WriteSymlink.
func (s *DumpSystem) WriteSymlink(oldName string, newName AbsPath) error {
	return s.setData(newName.String(), &symlinkData{
		Type:     dataTypeSymlink,
		Name:     newName,
		Linkname: oldName,
	})
}

func (s *DumpSystem) setData(key string, value any) error {
	if _, ok := s.data[key]; ok {
		return fs.ErrExist
	}
	s.data[key] = value
	return nil
}
