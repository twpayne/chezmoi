package xdg

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// A BaseDirectorySpecification represents an XDG Base Directory Specification
// configuration. See
// https://standards.freedesktop.org/basedir-spec/basedir-spec-latest.html.
type BaseDirectorySpecification struct {
	ConfigHome string
	ConfigDirs []string
	DataHome   string
	DataDirs   []string
	CacheHome  string
	RuntimeDir string
	StateHome  string
}

// A GetenvFunc is a function that gets an environment variable, like os.Getenv.
type GetenvFunc func(string) string

// NewBaseDirectorySpecification returns a new BaseDirectorySpecification,
// configured from the user's environment variables.
func NewBaseDirectorySpecification() (*BaseDirectorySpecification, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewTestBaseDirectorySpecification(homeDir, os.Getenv), nil
}

// NewTestBaseDirectorySpecification returns a new BaseDirectorySpecification
// using homeDir and the getenv function. getenv can be nil, in which case
// default values are returned.
func NewTestBaseDirectorySpecification(homeDir string, getenv GetenvFunc) *BaseDirectorySpecification {
	if getenv == nil {
		getenv = func(string) string {
			return ""
		}
	}

	defaultConfigHome := filepath.Join(homeDir, ".config")
	configHome := firstNonEmpty(getenv("XDG_CONFIG_HOME"), defaultConfigHome)

	defaultConfigDirs := filepath.Join("/", "etc", "xdg")
	configDirs := append(
		[]string{configHome},
		filepath.SplitList(firstNonEmpty(getenv("XDG_CONFIG_DIRS"), defaultConfigDirs))...,
	)

	defaultDataHome := filepath.Join(homeDir, ".local", "share")
	dataHome := firstNonEmpty(getenv("XDG_DATA_HOME"), defaultDataHome)

	defaultDataDirs := strings.Join([]string{
		filepath.Join("/", "usr", "local", "share"),
		filepath.Join("/", "usr", "share"),
	}, string(filepath.ListSeparator))
	dataDirs := append([]string{dataHome}, filepath.SplitList(firstNonEmpty(getenv("XDG_DATA_DIRS"), defaultDataDirs))...)

	defaultCacheHome := filepath.Join(homeDir, ".cache")
	cacheHome := firstNonEmpty(getenv("XDG_CACHE_HOME"), defaultCacheHome)

	runtimeDir := getenv("XDG_RUNTIME_DIR")

	defaultStateHome := filepath.Join(homeDir, ".local", "state")
	stateHome := firstNonEmpty(getenv("XDG_STATE_HOME"), defaultStateHome)

	return &BaseDirectorySpecification{
		ConfigHome: configHome,
		ConfigDirs: configDirs,
		DataHome:   dataHome,
		DataDirs:   dataDirs,
		CacheHome:  cacheHome,
		RuntimeDir: runtimeDir,
		StateHome:  stateHome,
	}
}

// OpenConfigFile opens the first configuration file with the given name found,
// its full path, and any error. If no file can be found, the error will be
// fs.ErrNotExist.
//
// The file is opened with the open argument. If open is nil, os.Open is used.
func (b *BaseDirectorySpecification) OpenConfigFile(fsys fs.FS, nameComponents ...string) (fs.File, string, error) {
	return openFile(fsys, nameComponents, b.ConfigDirs)
}

// OpenDataFile opens the first data file with the given name found, its full
// path, and any error. If no file can be found, the error will be
// fs.ErrNotExist.
//
// The file is opened with the open argument. If open is nil, os.Open is used.
func (b *BaseDirectorySpecification) OpenDataFile(fsys fs.FS, nameComponents ...string) (fs.File, string, error) {
	return openFile(fsys, nameComponents, b.DataDirs)
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func openFile(fsys fs.FS, nameComponents, dirs []string) (fs.File, string, error) {
	for _, dir := range dirs {
		path := filepath.Join(append([]string{dir}, nameComponents...)...)
		f, err := fsys.Open(path)
		switch {
		case err == nil:
			return f, path, nil
		case errors.Is(err, fs.ErrNotExist):
			continue
		default:
			return nil, path, err
		}
	}
	return nil, "", fs.ErrNotExist
}
