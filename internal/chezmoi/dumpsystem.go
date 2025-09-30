package chezmoi

import (
	"io/fs"
	"os/exec"

	vfs "github.com/twpayne/go-vfs/v5"
)

// A DumpSystemDataType is a data type in a dump system.
type DumpSystemDataType string

// Dump system data types.
const (
	DumpSystemDataTypeCommand DumpSystemDataType = "command"
	DumpSystemDataTypeDir     DumpSystemDataType = "dir"
	DumpSystemDataTypeFile    DumpSystemDataType = "file"
	DumpSystemDataTypeScript  DumpSystemDataType = "script"
	DumpSystemDataTypeSymlink DumpSystemDataType = "symlink"
)

// A DumpSystem is a System that writes to a data file.
type DumpSystem struct {
	emptySystemMixin
	noUpdateSystemMixin

	Data map[string]any
}

// A DumpSystemCommandData contains data about a command.
type DumpSystemCommandData struct {
	Type DumpSystemDataType `json:"type" yaml:"type"`
	Path string             `json:"path" yaml:"path"`
	Args []string           `json:"args" yaml:"args"`
}

// A DumpSystemDirData contains data about a directory.
type DumpSystemDirData struct {
	Type DumpSystemDataType `json:"type" yaml:"type"`
	Name AbsPath            `json:"name" yaml:"name"`
	Perm fs.FileMode        `json:"perm" yaml:"perm"`
}

// A DumpSystemFileData contains data about a file.
type DumpSystemFileData struct {
	Type     DumpSystemDataType `json:"type"     yaml:"type"`
	Name     AbsPath            `json:"name"     yaml:"name"`
	Contents string             `json:"contents" yaml:"contents"`
	Perm     fs.FileMode        `json:"perm"     yaml:"perm"`
}

// A DumpSystemScriptData contains data about a script.
type DumpSystemScriptData struct {
	Type        DumpSystemDataType `json:"type"                  yaml:"type"`
	Name        AbsPath            `json:"name"                  yaml:"name"`
	Contents    string             `json:"contents"              yaml:"contents"`
	Condition   string             `json:"condition"             yaml:"condition"`
	Interpreter *Interpreter       `json:"interpreter,omitempty" yaml:"interpreter,omitempty"`
}

// A DumpSystemSymlinkData contains data about a symlink.
type DumpSystemSymlinkData struct {
	Type     DumpSystemDataType `json:"type"     yaml:"type"`
	Name     AbsPath            `json:"name"     yaml:"name"`
	Linkname string             `json:"linkname" yaml:"linkname"`
}

// NewDumpSystem returns a new DumpSystem that accumulates data.
func NewDumpSystem() *DumpSystem {
	return &DumpSystem{
		Data: make(map[string]any),
	}
}

// Mkdir implements System.Mkdir.
func (s *DumpSystem) Mkdir(dirname AbsPath, perm fs.FileMode) error {
	return s.setData(dirname.String(), &DumpSystemDirData{
		Type: DumpSystemDataTypeDir,
		Name: dirname,
		Perm: perm,
	})
}

// RunCmd implements System.RunCmd.
func (s *DumpSystem) RunCmd(cmd *exec.Cmd) error {
	if cmd.Dir == "" {
		return nil
	}
	return s.setData(cmd.Dir, &DumpSystemCommandData{
		Type: DumpSystemDataTypeCommand,
		Path: cmd.Path,
		Args: cmd.Args,
	})
}

// RunScript implements System.RunScript.
func (s *DumpSystem) RunScript(scriptName RelPath, dir AbsPath, data []byte, options RunScriptOptions) error {
	scriptNameStr := scriptName.String()
	scriptData := &DumpSystemScriptData{
		Type:     DumpSystemDataTypeScript,
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
	return s.setData(filename.String(), &DumpSystemFileData{
		Type:     DumpSystemDataTypeFile,
		Name:     filename,
		Contents: string(data),
		Perm:     perm,
	})
}

// WriteSymlink implements System.WriteSymlink.
func (s *DumpSystem) WriteSymlink(oldName string, newName AbsPath) error {
	return s.setData(newName.String(), &DumpSystemSymlinkData{
		Type:     DumpSystemDataTypeSymlink,
		Name:     newName,
		Linkname: oldName,
	})
}

func (s *DumpSystem) setData(key string, value any) error {
	if _, ok := s.Data[key]; ok {
		return fs.ErrExist
	}
	s.Data[key] = value
	return nil
}
