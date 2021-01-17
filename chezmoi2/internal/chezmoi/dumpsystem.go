package chezmoi

import (
	"os"

	vfs "github.com/twpayne/go-vfs"
)

// A dataType is a data type.
type dataType string

// dataTypes.
const (
	dataTypeDir     dataType = "dir"
	dataTypeFile    dataType = "file"
	dataTypeScript  dataType = "script"
	dataTypeSymlink dataType = "symlink"
)

// A DumpSystem is a System that writes to a data file.
type DumpSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
	data map[AbsPath]interface{}
}

// A dirData contains data about a directory.
type dirData struct {
	Type dataType    `json:"type" toml:"type" yaml:"type"`
	Name AbsPath     `json:"name" toml:"name" yaml:"name"`
	Perm os.FileMode `json:"perm" toml:"perm" yaml:"perm"`
}

// A fileData contains data about a file.
type fileData struct {
	Type     dataType    `json:"type" toml:"type" yaml:"type"`
	Name     AbsPath     `json:"name" toml:"name" yaml:"name"`
	Contents string      `json:"contents" toml:"contents" yaml:"contents"`
	Perm     os.FileMode `json:"perm" toml:"perm" yaml:"perm"`
}

// A scriptData contains data about a script.
type scriptData struct {
	Type     dataType `json:"type" toml:"type" yaml:"type"`
	Name     AbsPath  `json:"name" toml:"name" yaml:"name"`
	Contents string   `json:"contents" toml:"contents" yaml:"contents"`
}

// A symlinkData contains data about a symlink.
type symlinkData struct {
	Type     dataType `json:"type" toml:"type" yaml:"type"`
	Name     AbsPath  `json:"name" toml:"name" yaml:"name"`
	Linkname string   `json:"linkname" toml:"linkname" yaml:"linkname"`
}

// NewDumpSystem returns a new DumpSystem that accumulates data.
func NewDumpSystem() *DumpSystem {
	return &DumpSystem{
		data: make(map[AbsPath]interface{}),
	}
}

// Data returns s's data.
func (s *DumpSystem) Data() interface{} {
	return s.data
}

// Mkdir implements System.Mkdir.
func (s *DumpSystem) Mkdir(dirname AbsPath, perm os.FileMode) error {
	if _, exists := s.data[dirname]; exists {
		return os.ErrExist
	}
	s.data[dirname] = &dirData{
		Type: dataTypeDir,
		Name: dirname,
		Perm: perm,
	}
	return nil
}

// RunScript implements System.RunScript.
func (s *DumpSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte) error {
	scriptnameAbsPath := AbsPath(scriptname)
	if _, exists := s.data[scriptnameAbsPath]; exists {
		return os.ErrExist
	}
	s.data[scriptnameAbsPath] = &scriptData{
		Type:     dataTypeScript,
		Name:     scriptnameAbsPath,
		Contents: string(data),
	}
	return nil
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *DumpSystem) UnderlyingFS() vfs.FS {
	return nil
}

// WriteFile implements System.WriteFile.
func (s *DumpSystem) WriteFile(filename AbsPath, data []byte, perm os.FileMode) error {
	if _, exists := s.data[filename]; exists {
		return os.ErrExist
	}
	s.data[filename] = &fileData{
		Type:     dataTypeFile,
		Name:     filename,
		Contents: string(data),
		Perm:     perm,
	}
	return nil
}

// WriteSymlink implements System.WriteSymlink.
func (s *DumpSystem) WriteSymlink(oldname string, newname AbsPath) error {
	if _, exists := s.data[newname]; exists {
		return os.ErrExist
	}
	s.data[newname] = &symlinkData{
		Type:     dataTypeSymlink,
		Name:     newname,
		Linkname: oldname,
	}
	return nil
}
