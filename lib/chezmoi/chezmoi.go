package chezmoi

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
)

var (
	dirNameRegexp         = regexp.MustCompile(`\A(?P<private>private_)?(?P<dot>dot_)?(?P<name>.*)\z`)
	dirNameSubexpIndexes  = makeSubexpIndexes(dirNameRegexp)
	fileNameRegexp        = regexp.MustCompile(`\A(?P<private>private_)?(?P<executable>executable_)?(?P<dot>dot_)?(?P<name>.*?)(?P<template>\.tmpl)?\z`)
	fileNameSubexpIndexes = makeSubexpIndexes(fileNameRegexp)
)

// A FileState represents the target state of a file.
type FileState struct {
	Name     string
	Mode     os.FileMode
	Contents []byte
}

// A DirState represents the target state of a directory.
type DirState struct {
	Name  string
	Mode  os.FileMode
	Dirs  map[string]*DirState
	Files map[string]*FileState
}

// parseDirName parses a single directory name. It returns the target name,
// mode, and any error.
func parseDirName(dirName string) (string, os.FileMode, error) {
	m := dirNameRegexp.FindStringSubmatch(dirName)
	if m == nil {
		return "", os.FileMode(0), errors.Errorf("invalid directory name: %s", dirName)
	}
	name := m[dirNameSubexpIndexes["name"]]
	if m[dirNameSubexpIndexes["dot"]] != "" {
		name = "." + name
	}
	mode := os.FileMode(0777)
	if m[dirNameSubexpIndexes["private"]] != "" {
		mode &= 0700
	}
	return name, mode, nil
}

// parseFileName parses a single file name. It returns the target name, mode,
// whether the contents should be interpreted as a template, and any error.
func parseFileName(fileName string) (string, os.FileMode, bool, error) {
	m := fileNameRegexp.FindStringSubmatch(fileName)
	if m == nil {
		return "", os.FileMode(0), false, errors.Errorf("invalid file name: %s", fileName)
	}
	name := m[fileNameSubexpIndexes["name"]]
	if m[fileNameSubexpIndexes["dot"]] != "" {
		name = "." + name
	}
	mode := os.FileMode(0666)
	if m[fileNameSubexpIndexes["executable"]] != "" {
		mode |= 0111
	}
	if m[fileNameSubexpIndexes["private"]] != "" {
		mode &= 0700
	}
	isTemplate := m[fileNameSubexpIndexes["template"]] != ""
	return name, mode, isTemplate, nil
}

// parseDirNameComponents parses multiple directory name components. It returns
// the target directory names, target modes, and any error.
func parseDirNameComponents(components []string) ([]string, []os.FileMode, error) {
	dirNames := []string{}
	modes := []os.FileMode{}
	for _, component := range components {
		dirName, mode, err := parseDirName(component)
		if err != nil {
			return nil, nil, err
		}
		dirNames = append(dirNames, dirName)
		modes = append(modes, mode)
	}
	return dirNames, modes, nil
}

// parseFilePath parses a single file path. It returns the target directory
// names, the target filename, the target mode, whether the contents should be
// interpreted as a template, and any error.
func parseFilePath(path string) ([]string, string, os.FileMode, bool, error) {
	if path == "" {
		return nil, "", os.FileMode(0), false, errors.New("empty path")
	}
	components := splitPathList(path)
	dirNames, _, err := parseDirNameComponents(components[0 : len(components)-1])
	if err != nil {
		return nil, "", os.FileMode(0), false, err
	}
	fileName, mode, isTemplate, err := parseFileName(components[len(components)-1])
	if err != nil {
		return nil, "", os.FileMode(0), false, err
	}
	return dirNames, fileName, mode, isTemplate, nil
}

// newDirState returns a new directory state.
func newDirState(name string, mode os.FileMode) *DirState {
	return &DirState{
		Name:  name,
		Mode:  mode,
		Dirs:  make(map[string]*DirState),
		Files: make(map[string]*FileState),
	}
}

// newRootDirState returns a new root directory state.
func newRootDirState() *DirState {
	return newDirState("", os.FileMode(0))
}

// isRoot returns whether ds refers to the root.
func (ds *DirState) isRoot() bool {
	return ds.Name == "" && ds.Mode == os.FileMode(0)
}

// Apply ensures that targetDir in fs matches ds.
func (ds *DirState) Apply(fs afero.Fs, targetDir string) error {
	if !ds.isRoot() {
		if _, dirName := filepath.Split(targetDir); dirName != ds.Name {
			return errors.Errorf("name mismatch: got %s, want %s", dirName, ds.Name)
		}
		fi, err := fs.Stat(targetDir)
		switch {
		case err == nil && fi.Mode().IsDir():
			if fi.Mode()&os.ModePerm != ds.Mode {
				if err := fs.Chmod(targetDir, ds.Mode); err != nil {
					return err
				}
			}
		case err == nil:
			if err := fs.RemoveAll(targetDir); err != nil {
				return err
			}
			fallthrough
		case os.IsNotExist(err):
			if err := fs.Mkdir(targetDir, ds.Mode); err != nil {
				return err
			}
		default:
			return err
		}
	}
	for fileName, fileState := range ds.Files {
		if err := fileState.Apply(fs, filepath.Join(targetDir, fileName)); err != nil {
			return err
		}
	}
	for dirName, dirState := range ds.Dirs {
		if err := dirState.Apply(fs, filepath.Join(targetDir, dirName)); err != nil {
			return err
		}
	}
	return nil
}

// ReadSourceDirState walks fs from sourceDir creating a target directory
// state. Any templates found are executed with data.
func ReadSourceDirState(fs afero.Fs, sourceDir string, data interface{}) (*DirState, error) {
	rootDS := newRootDirState()
	if err := afero.Walk(fs, sourceDir, func(path string, fi os.FileInfo, err error) error {
		if path == sourceDir {
			return nil
		}
		relPath := strings.TrimPrefix(path, sourceDir)
		switch {
		case fi.Mode().IsRegular():
			dirNames, fileName, mode, isTemplate, err := parseFilePath(relPath)
			if err != nil {
				return errors.Wrap(err, path)
			}
			ds := rootDS
			for _, dirName := range dirNames {
				ds = ds.Dirs[dirName]
			}
			r, err := fs.Open(path)
			if err != nil {
				return err
			}
			defer r.Close()
			contents, err := ioutil.ReadAll(r)
			if err != nil {
				return errors.Wrap(err, path)
			}
			if isTemplate {
				tmpl, err := template.New(path).Parse(string(contents))
				if err != nil {
					return errors.Wrap(err, path)
				}
				output := &bytes.Buffer{}
				if err := tmpl.Execute(output, data); err != nil {
					return errors.Wrap(err, path)
				}
				contents = output.Bytes()
			}
			ds.Files[fileName] = &FileState{
				Name:     fileName,
				Mode:     mode,
				Contents: contents,
			}
		case fi.Mode().IsDir():
			components := splitPathList(relPath)
			dirNames, modes, err := parseDirNameComponents(components)
			if err != nil {
				return errors.Wrap(err, path)
			}
			ds := rootDS
			for i := 0; i < len(dirNames)-1; i++ {
				ds = ds.Dirs[dirNames[i]]
			}
			dirName := dirNames[len(dirNames)-1]
			mode := modes[len(modes)-1]
			ds.Dirs[dirName] = newDirState(dirName, mode)
		default:
			return errors.Errorf("unsupported file type: %s", path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return rootDS, nil
}

// Apply ensures that state of targetPath in fs matches fileState.
func (fileState *FileState) Apply(fs afero.Fs, targetPath string) error {
	if _, fileName := filepath.Split(targetPath); fileName != fileState.Name {
		return errors.Errorf("name mismatch: got %s, want %s", targetPath, fileState.Name)
	}
	fi, err := fs.Stat(targetPath)
	switch {
	case err == nil && fi.Mode().IsRegular() && fi.Mode()&os.ModePerm == fileState.Mode:
		f, err := fs.Open(targetPath)
		if err != nil {
			return err
		}
		defer f.Close()
		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return errors.Wrap(err, targetPath)
		}
		if reflect.DeepEqual(contents, fileState.Contents) {
			return nil
		}
	case err == nil:
		if err := fs.RemoveAll(targetPath); err != nil {
			return err
		}
	case os.IsNotExist(err):
	default:
		return err
	}
	// FIXME atomically replace
	return afero.WriteFile(fs, targetPath, fileState.Contents, fileState.Mode)
}

func makeSubexpIndexes(re *regexp.Regexp) map[string]int {
	result := make(map[string]int)
	for index, name := range re.SubexpNames() {
		result[name] = index
	}
	return result
}

func splitPathList(path string) []string {
	components := strings.Split(path, string(filepath.Separator))
	if components[0] == "" {
		return components[1:len(components)]
	}
	return components
}
