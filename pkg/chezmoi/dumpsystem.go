package chezmoi

import (
	"io/fs"
	"os/exec"

	vfs "github.com/twpayne/go-vfs/v4"
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
	data map[string]interface{}
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
	Type     dataType    `json:"type" yaml:"type"`
	Name     AbsPath     `json:"name" yaml:"name"`
	Contents string      `json:"contents" yaml:"contents"`
	Perm     fs.FileMode `json:"perm" yaml:"perm"`
}

// A scriptData contains data about a script.
type scriptData struct {
	Type        dataType     `json:"type" yaml:"type"`
	Name        AbsPath      `json:"name" yaml:"name"`
	Contents    string       `json:"contents" yaml:"contents"`
	Interpreter *Interpreter `json:"interpreter,omitempty" yaml:"interpreter,omitempty"`
}

// A symlinkData contains data about a symlink.
type symlinkData struct {
	Type     dataType `json:"type" yaml:"type"`
	Name     AbsPath  `json:"name" yaml:"name"`
	Linkname string   `json:"linkname" yaml:"linkname"`
}

// NewDumpSystem returns a new DumpSystem that accumulates data.
func NewDumpSystem() *DumpSystem {
	return &DumpSystem{
		data: make(map[string]interface{}),
	}
}

// Data returns s's data.
func (s *DumpSystem) Data() interface{} {
	return s.data
}

// Mkdir implements System.Mkdir.
func (s *DumpSystem) Mkdir(dirname AbsPath, perm fs.FileMode) error {
	if _, exists := s.data[dirname.String()]; exists {
		return fs.ErrExist
	}
	s.data[dirname.String()] = &dirData{
		Type: dataTypeDir,
		Name: dirname,
		Perm: perm,
	}
	return nil
}

// RunCmd implements System.RunCmd.
func (s *DumpSystem) RunCmd(cmd *exec.Cmd) error {
	if cmd.Dir == "" {
		return nil
	}
	s.data[cmd.Dir] = &commandData{
		Type: dataTypeCommand,
		Path: cmd.Path,
		Args: cmd.Args,
	}
	return nil
}

// RunScript implements System.RunScript.
func (s *DumpSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) error {
	scriptnameStr := scriptname.String()
	if _, exists := s.data[scriptnameStr]; exists {
		return fs.ErrExist
	}
	scriptData := &scriptData{
		Type:     dataTypeScript,
		Name:     NewAbsPath(scriptnameStr),
		Contents: string(data),
	}
	if !interpreter.None() {
		scriptData.Interpreter = interpreter
	}
	s.data[scriptnameStr] = scriptData
	return nil
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DumpSystem) UnderlyingFS() vfs.FS {
	return nil
}

// WriteFile implements System.WriteFile.
func (s *DumpSystem) WriteFile(filename AbsPath, data []byte, perm fs.FileMode) error {
	filenameStr := filename.String()
	if _, exists := s.data[filenameStr]; exists {
		return fs.ErrExist
	}
	s.data[filenameStr] = &fileData{
		Type:     dataTypeFile,
		Name:     filename,
		Contents: string(data),
		Perm:     perm,
	}
	return nil
}

// WriteSymlink implements System.WriteSymlink.
func (s *DumpSystem) WriteSymlink(oldname string, newname AbsPath) error {
	newnameStr := newname.String()
	if _, exists := s.data[newnameStr]; exists {
		return fs.ErrExist
	}
	s.data[newnameStr] = &symlinkData{
		Type:     dataTypeSymlink,
		Name:     newname,
		Linkname: oldname,
	}
	return nil
}
