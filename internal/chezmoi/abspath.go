package chezmoi

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

var (
	DotAbsPath   = NewAbsPath(".")
	EmptyAbsPath = NewAbsPath("")
	RootAbsPath  = NewAbsPath("/")
)

// An AbsPath is an absolute path.
type AbsPath string

// NewAbsPath returns a new AbsPath.
func NewAbsPath(absPath string) AbsPath {
	return AbsPath(filepath.ToSlash(absPath))
}

// Append appends s to p.
func (p AbsPath) Append(s string) AbsPath {
	return NewAbsPath(string(p) + s)
}

// Base returns p's basename.
func (p AbsPath) Base() string {
	return path.Base(string(p))
}

// Bytes returns p as a []byte.
func (p AbsPath) Bytes() []byte {
	return []byte(p)
}

// Dir returns p's directory.
func (p AbsPath) Dir() AbsPath {
	return NewAbsPath(filepath.Dir(string(p))).ToSlash()
}

// IsEmpty returns if p is empty.
func (p AbsPath) IsEmpty() bool {
	return p == ""
}

// Ext returns p's extension.
func (p AbsPath) Ext() string {
	return path.Ext(string(p))
}

// HasDirPrefix returns if p has the given prefix.
func (p AbsPath) HasDirPrefix(prefix AbsPath) bool {
	pWithTrailingSlash := p.WithTrailingSlash()
	prefixWithTrailingSlash := prefix.WithTrailingSlash()
	if pWithTrailingSlash == prefixWithTrailingSlash {
		return true
	}
	return strings.HasPrefix(string(pWithTrailingSlash), string(prefixWithTrailingSlash))
}

// Join returns a new AbsPath with relPaths appended.
func (p AbsPath) Join(relPaths ...RelPath) AbsPath {
	relPathStrs := make([]string, len(relPaths)+1)
	relPathStrs[0] = string(p)
	for i, relPath := range relPaths {
		relPathStrs[i+1] = relPath.String()
	}
	return NewAbsPath(filepath.ToSlash(filepath.Join(relPathStrs...)))
}

// JoinString returns a new AbsPath with ss appended.
func (p AbsPath) JoinString(ss ...string) AbsPath {
	strs := make([]string, len(ss)+1)
	strs[0] = string(p)
	copy(strs[1:len(ss)+1], ss)
	return NewAbsPath(filepath.ToSlash(filepath.Join(strs...)))
}

// Len returns the length of p.
func (p AbsPath) Len() int {
	return len(p)
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
		*p = ""
		return nil
	}
	homeDirAbsPath, err := HomeDirAbsPath()
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
	return NewAbsPath(strings.TrimSuffix(dir, "/")), NewRelPath(file)
}

func (p AbsPath) String() string {
	return string(p)
}

// ToSlash calls [filepath.ToSlash] on p.
func (p AbsPath) ToSlash() AbsPath {
	return NewAbsPath(filepath.ToSlash(string(p)))
}

// TrimDirPrefix trims prefix from p.
func (p AbsPath) TrimDirPrefix(dirPrefixAbsPath AbsPath) (RelPath, error) {
	if p == dirPrefixAbsPath {
		return EmptyRelPath, nil
	}
	dirAbsPath := dirPrefixAbsPath.WithTrailingSlash()
	if !strings.HasPrefix(string(p), string(dirAbsPath)) {
		return EmptyRelPath, &NotInAbsDirError{
			pathAbsPath: p,
			dirAbsPath:  dirPrefixAbsPath,
		}
	}
	return NewRelPath(string(p[len(dirAbsPath):])), nil
}

// TrimSuffix returns p with the optional suffix removed.
func (p AbsPath) TrimSuffix(suffix string) AbsPath {
	return NewAbsPath(strings.TrimSuffix(string(p), suffix))
}

// Type implements github.com/spf13/pflag.Value.Type.
func (p AbsPath) Type() string {
	return "path"
}

// WithTrailingSlash returns p with a trailing slash.
func (p AbsPath) WithTrailingSlash() AbsPath {
	if strings.HasSuffix(string(p), "/") {
		return p
	}
	return p + "/"
}

// HomeDirAbsPath returns the user's home directory as an AbsPath.
func HomeDirAbsPath() (AbsPath, error) {
	userHomeDir, err := UserHomeDir()
	if err != nil {
		return EmptyAbsPath, err
	}
	absPath, err := NormalizePath(userHomeDir)
	if err != nil {
		return EmptyAbsPath, err
	}
	return absPath, nil
}

// StringToAbsPathHookFunc is a
// github.com/go-viper/mapstructure/v2.DecodeHookFunc that parses an AbsPath
// from a string.
func StringToAbsPathHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeFor[AbsPath]() {
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
