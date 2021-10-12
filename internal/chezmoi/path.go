package chezmoi

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

var (
	DotAbsPath   = NewAbsPath(".")
	EmptyAbsPath = NewAbsPath("")
	RootAbsPath  = NewAbsPath("/")
)

// An AbsPath is an absolute path.
type AbsPath struct {
	absPath string
}

// NewAbsPath returns a new AbsPath.
func NewAbsPath(absPath string) AbsPath {
	return AbsPath{
		absPath: absPath,
	}
}

// Base returns p's basename.
func (p AbsPath) Base() string {
	return path.Base(p.absPath)
}

// Bytes returns p as a []byte.
func (p AbsPath) Bytes() []byte {
	return []byte(p.absPath)
}

// Dir returns p's directory.
func (p AbsPath) Dir() AbsPath {
	return NewAbsPath(path.Dir(p.absPath))
}

// Empty returns if p is empty.
func (p AbsPath) Empty() bool {
	return p.absPath == ""
}

// Ext returns p's extension.
func (p AbsPath) Ext() string {
	return path.Ext(p.absPath)
}

// Join appends elems to p.
func (p AbsPath) Join(elems ...RelPath) AbsPath {
	elemStrs := make([]string, 0, len(elems)+1)
	elemStrs = append(elemStrs, p.absPath)
	for _, elem := range elems {
		elemStrs = append(elemStrs, string(elem))
	}
	return NewAbsPath(path.Join(elemStrs...))
}

// Len returns the length of p.
func (p AbsPath) Len() int {
	return len(p.absPath)
}

// MarshalText implements encoding.TextMarshaler.
func (p AbsPath) MarshalText() ([]byte, error) {
	return []byte(p.absPath), nil
}

// MustTrimDirPrefix is like TrimPrefix but panics on any error.
func (p AbsPath) MustTrimDirPrefix(dirPrefix AbsPath) RelPath {
	relPath, err := p.TrimDirPrefix(dirPrefix)
	if err != nil {
		panic(err)
	}
	return relPath
}

// Set implements github.com/spf13/pflag.Value.Set.
func (p *AbsPath) Set(s string) error {
	if s == "" {
		p.absPath = ""
		return nil
	}
	homeDirAbsPath, err := homeDirAbsPath()
	if err != nil {
		return err
	}
	absPath, err := NewAbsPathFromExtPath(s, homeDirAbsPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}

// Split returns p's directory and file.
func (p AbsPath) Split() (AbsPath, RelPath) {
	dir, file := path.Split(p.String())
	return NewAbsPath(dir), RelPath(file)
}

func (p AbsPath) String() string {
	return p.absPath
}

// ToSlash calls filepath.ToSlash on p.
func (p AbsPath) ToSlash() AbsPath {
	return NewAbsPath(filepath.ToSlash(p.absPath))
}

// TrimDirPrefix trims prefix from p.
func (p AbsPath) TrimDirPrefix(dirPrefixAbsPath AbsPath) (RelPath, error) {
	if p == dirPrefixAbsPath {
		return "", nil
	}
	dirAbsPath := dirPrefixAbsPath
	if dirAbsPath.absPath != "/" {
		dirAbsPath.absPath += "/"
	}
	if !strings.HasPrefix(p.absPath, dirAbsPath.absPath) {
		return "", &notInAbsDirError{
			pathAbsPath: p,
			dirAbsPath:  dirPrefixAbsPath,
		}
	}
	return RelPath(p.absPath[len(dirAbsPath.absPath):]), nil
}

// Type implements github.com/spf13/pflag.Value.Type.
func (p AbsPath) Type() string {
	return "path"
}

// UnmarshalText implements encoding.UnmarshalText.
func (p *AbsPath) UnmarshalText(text []byte) error {
	return p.Set(string(text))
}

// A RelPath is a relative path.
type RelPath string

// Base returns p's base name.
func (p RelPath) Base() string {
	return path.Base(string(p))
}

// Dir returns p's directory.
func (p RelPath) Dir() RelPath {
	return RelPath(path.Dir(string(p)))
}

// Ext returns p's extension.
func (p RelPath) Ext() string {
	return path.Ext(string(p))
}

// HasDirPrefix returns true if p has dir prefix dirPrefix.
func (p RelPath) HasDirPrefix(dirPrefix RelPath) bool {
	return strings.HasPrefix(string(p), string(dirPrefix)+"/")
}

// Join appends elems to p.
func (p RelPath) Join(elems ...RelPath) RelPath {
	elemStrs := make([]string, 0, len(elems)+1)
	elemStrs = append(elemStrs, string(p))
	for _, elem := range elems {
		elemStrs = append(elemStrs, string(elem))
	}
	return RelPath(path.Join(elemStrs...))
}

// Split returns p's directory and path.
func (p RelPath) Split() (RelPath, RelPath) {
	dir, file := path.Split(string(p))
	return RelPath(dir), RelPath(file)
}

func (p RelPath) String() string {
	return string(p)
}

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return "", &notInRelDirError{
			pathRelPath: p,
			dirRelPath:  dirPrefix,
		}
	}
	return p[len(dirPrefix)+1:], nil
}

// StringToAbsPathHookFunc is a github.com/mitchellh/mapstructure.DecodeHookFunc
// that parses an AbsPath from a string.
func StringToAbsPathHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if to != reflect.TypeOf(EmptyAbsPath) {
			return data, nil
		}
		s, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string, got a %T", data)
		}
		var absPath AbsPath
		if err := absPath.Set(s); err != nil {
			return nil, err
		}
		return absPath, nil
	}
}

// homeDirAbsPath returns the user's home directory as an AbsPath.
func homeDirAbsPath() (AbsPath, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return EmptyAbsPath, err
	}
	absPath, err := NormalizePath(userHomeDir)
	if err != nil {
		return EmptyAbsPath, err
	}
	return absPath, nil
}
