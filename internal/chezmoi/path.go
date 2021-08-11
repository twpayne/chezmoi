package chezmoi

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// An AbsPath is an absolute path.
type AbsPath string

// Base returns p's basename.
func (p AbsPath) Base() string {
	return path.Base(string(p))
}

// Dir returns p's directory.
func (p AbsPath) Dir() AbsPath {
	return AbsPath(path.Dir(string(p)))
}

// Join appends elems to p.
func (p AbsPath) Join(elems ...RelPath) AbsPath {
	elemStrs := make([]string, 0, len(elems)+1)
	elemStrs = append(elemStrs, string(p))
	for _, elem := range elems {
		elemStrs = append(elemStrs, string(elem))
	}
	return AbsPath(path.Join(elemStrs...))
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
	dir, file := path.Split(string(p))
	return AbsPath(dir), RelPath(file)
}

func (p AbsPath) String() string {
	return string(p)
}

// TrimDirPrefix trims prefix from p.
func (p AbsPath) TrimDirPrefix(dirPrefixAbsPath AbsPath) (RelPath, error) {
	dirAbsPath := dirPrefixAbsPath
	if dirAbsPath != "/" {
		dirAbsPath += "/"
	}
	if !strings.HasPrefix(string(p), string(dirAbsPath)) {
		return "", &errNotInAbsDir{
			pathAbsPath: p,
			dirAbsPath:  dirPrefixAbsPath,
		}
	}
	return RelPath(p[len(dirAbsPath):]), nil
}

// Type implements github.com/spf13/pflag.Value.Type.
func (p AbsPath) Type() string {
	return "path"
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

// TrimDirPrefix trims prefix from p.
func (p RelPath) TrimDirPrefix(dirPrefix RelPath) (RelPath, error) {
	if !p.HasDirPrefix(dirPrefix) {
		return "", &errNotInRelDir{
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
		if to != reflect.TypeOf(AbsPath("")) {
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
		return AbsPath(""), err
	}
	absPath, err := NormalizePath(userHomeDir)
	if err != nil {
		return AbsPath(""), err
	}
	return absPath, nil
}
