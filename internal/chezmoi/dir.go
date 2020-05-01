package chezmoi

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"

	vfs "github.com/twpayne/go-vfs"
)

// DirAttributes holds attributes parsed from a source directory name.
type DirAttributes struct {
	Name  string
	Exact bool
	Perm  os.FileMode
}

// A Dir represents the target state of a directory.
type Dir struct {
	sourceName string
	targetName string
	Exact      bool
	Perm       os.FileMode
	Entries    map[string]Entry
}

type dirConcreteValue struct {
	Type       string        `json:"type" yaml:"type"`
	SourcePath string        `json:"sourcePath" yaml:"sourcePath"`
	TargetPath string        `json:"targetPath" yaml:"targetPath"`
	Exact      bool          `json:"exact" yaml:"exact"`
	Perm       int           `json:"perm" yaml:"perm"`
	Entries    []interface{} `json:"entries" yaml:"entries"`
}

// ParseDirAttributes parses a single directory name.
func ParseDirAttributes(sourceName string) DirAttributes {
	name := sourceName
	perm := os.FileMode(0o777)
	exact := false
	if strings.HasPrefix(name, exactPrefix) {
		name = strings.TrimPrefix(name, exactPrefix)
		exact = true
	}
	if strings.HasPrefix(name, privatePrefix) {
		name = strings.TrimPrefix(name, privatePrefix)
		perm &= 0o700
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	return DirAttributes{
		Name:  name,
		Exact: exact,
		Perm:  perm,
	}
}

// SourceName returns da's source name.
func (da DirAttributes) SourceName() string {
	sourceName := ""
	if da.Exact {
		sourceName += exactPrefix
	}
	if da.Perm&os.FileMode(0o77) == os.FileMode(0) {
		sourceName += privatePrefix
	}
	if strings.HasPrefix(da.Name, ".") {
		sourceName += dotPrefix + strings.TrimPrefix(da.Name, ".")
	} else {
		sourceName += da.Name
	}
	return sourceName
}

// newDir returns a new directory state.
func newDir(sourceName, targetName string, exact bool, perm os.FileMode) *Dir {
	return &Dir{
		sourceName: sourceName,
		targetName: targetName,
		Exact:      exact,
		Perm:       perm,
		Entries:    make(map[string]Entry),
	}
}

// AppendAllEntries appends all Entries in d to allEntries.
func (d *Dir) AppendAllEntries(allEntries []Entry) []Entry {
	allEntries = append(allEntries, d)
	for _, entry := range d.Entries {
		allEntries = entry.AppendAllEntries(allEntries)
	}
	return allEntries
}

// Apply ensures that destDir in fs matches d.
func (d *Dir) Apply(fs vfs.FS, mutator Mutator, follow bool, applyOptions *ApplyOptions) error {
	if applyOptions.Ignore(d.targetName) {
		return nil
	}
	targetPath := filepath.Join(applyOptions.DestDir, d.targetName)
	var info os.FileInfo
	var err error
	if follow {
		info, err = fs.Stat(targetPath)
	} else {
		info, err = fs.Lstat(targetPath)
	}
	switch {
	case err == nil && info.IsDir():
		if info.Mode().Perm() != d.Perm&^applyOptions.Umask {
			if err := mutator.Chmod(targetPath, d.Perm&^applyOptions.Umask); err != nil {
				return err
			}
		}
	case err == nil:
		if err := mutator.RemoveAll(targetPath); err != nil {
			return err
		}
		fallthrough
	case os.IsNotExist(err):
		if err := mutator.Mkdir(targetPath, d.Perm&^applyOptions.Umask); err != nil {
			return err
		}
	default:
		return err
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].Apply(fs, mutator, follow, applyOptions); err != nil {
			return err
		}
	}
	if d.Exact {
		infos, err := fs.ReadDir(targetPath)
		if err != nil {
			return err
		}
		for _, info := range infos {
			name := info.Name()
			if _, ok := d.Entries[name]; !ok {
				if applyOptions.Ignore(filepath.Join(d.targetName, name)) {
					continue
				}
				if err := mutator.RemoveAll(filepath.Join(targetPath, name)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// ConcreteValue implements Entry.ConcreteValue.
func (d *Dir) ConcreteValue(ignore func(string) bool, sourceDir string, umask os.FileMode, recursive bool) (interface{}, error) {
	if ignore(d.targetName) {
		return nil, nil
	}
	var entryConcreteValues []interface{}
	if recursive {
		for _, entryName := range sortedEntryNames(d.Entries) {
			entryConcreteValue, err := d.Entries[entryName].ConcreteValue(ignore, sourceDir, umask, recursive)
			if err != nil {
				return nil, err
			}
			if entryConcreteValue != nil {
				entryConcreteValues = append(entryConcreteValues, entryConcreteValue)
			}
		}
	}
	return &dirConcreteValue{
		Type:       "dir",
		SourcePath: filepath.Join(sourceDir, d.SourceName()),
		TargetPath: d.TargetName(),
		Exact:      d.Exact,
		Perm:       int(d.Perm &^ umask),
		Entries:    entryConcreteValues,
	}, nil
}

// Evaluate evaluates all entries in d.
func (d *Dir) Evaluate(ignore func(string) bool) error {
	if ignore(d.targetName) {
		return nil
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].Evaluate(ignore); err != nil {
			return err
		}
	}
	return nil
}

// Private returns true if d is private.
func (d *Dir) Private() bool {
	return d.Perm&0o77 == 0
}

// SourceName implements Entry.SourceName.
func (d *Dir) SourceName() string {
	return d.sourceName
}

// TargetName implements Entry.TargetName.
func (d *Dir) TargetName() string {
	return d.targetName
}

// archive writes d to w.
func (d *Dir) archive(w *tar.Writer, ignore func(string) bool, headerTemplate *tar.Header, umask os.FileMode) error {
	if ignore(d.targetName) {
		return nil
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeDir
	header.Name = d.targetName
	header.Mode = int64(d.Perm &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return err
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].archive(w, ignore, headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}
