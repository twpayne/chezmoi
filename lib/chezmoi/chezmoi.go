package chezmoi

// FIXME rename FileState to File
// FIXME rename RootState to Root
// FIXME add Symlink

import (
	"archive/tar"
	"bytes"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/absfs/afero"
	"github.com/pkg/errors"
)

const (
	privatePrefix    = "private_"
	emptyPrefix      = "empty_"
	executablePrefix = "executable_"
	dotPrefix        = "dot_"
	templateSuffix   = ".tmpl"
)

// An Entry is either a Dir or a FileState.
type Entry interface {
	SourceName() string
	addAllEntries(map[string]Entry, string)
	apply(afero.Fs, string, os.FileMode, Actuator) error
	archive(*tar.Writer, string, *tar.Header, os.FileMode) error
}

// A FileState represents the target state of a file.
type FileState struct {
	sourceName string
	Empty      bool
	Mode       os.FileMode
	Contents   []byte
}

// A Dir represents the target state of a directory.
type Dir struct {
	sourceName string
	Mode       os.FileMode
	Entries    map[string]Entry
}

// A RootState represents the root target state.
type RootState struct {
	TargetDir string
	Umask     os.FileMode
	SourceDir string
	Data      map[string]interface{}
	Entries   map[string]Entry
}

// newDir returns a new directory state.
func newDir(sourceName string, mode os.FileMode) *Dir {
	return &Dir{
		sourceName: sourceName,
		Mode:       mode,
		Entries:    make(map[string]Entry),
	}
}

// addAllEntries adds d and all of the entries in d to result.
func (d *Dir) addAllEntries(result map[string]Entry, name string) {
	result[name] = d
	for entryName, entry := range d.Entries {
		entry.addAllEntries(result, filepath.Join(name, entryName))
	}
}

// archive writes d to w.
func (d *Dir) archive(w *tar.Writer, dirName string, headerTemplate *tar.Header, umask os.FileMode) error {
	header := *headerTemplate
	header.Typeflag = tar.TypeDir
	header.Name = dirName
	header.Mode = int64(d.Mode &^ umask & os.ModePerm)
	if err := w.WriteHeader(&header); err != nil {
		return err
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].archive(w, filepath.Join(dirName, entryName), headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}

// apply ensures that targetDir in fs matches d.
func (d *Dir) apply(fs afero.Fs, targetDir string, umask os.FileMode, actuator Actuator) error {
	fi, err := fs.Stat(targetDir)
	switch {
	case err == nil && fi.Mode().IsDir():
		if fi.Mode()&os.ModePerm != d.Mode&^umask {
			if err := actuator.Chmod(targetDir, d.Mode&^umask); err != nil {
				return err
			}
		}
	case err == nil:
		if err := actuator.RemoveAll(targetDir); err != nil {
			return err
		}
		fallthrough
	case os.IsNotExist(err):
		if err := actuator.Mkdir(targetDir, d.Mode&^umask); err != nil {
			return err
		}
	default:
		return err
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].apply(fs, filepath.Join(targetDir, entryName), umask, actuator); err != nil {
			return err
		}
	}
	return nil
}

// SourceName implements Entry.SourceName.
func (d *Dir) SourceName() string {
	return d.sourceName
}

// addAllEntries adds fs to result.
func (fs *FileState) addAllEntries(result map[string]Entry, name string) {
	result[name] = fs
}

// archive writes fs to w.
func (fs *FileState) archive(w *tar.Writer, fileName string, headerTemplate *tar.Header, umask os.FileMode) error {
	if len(fs.Contents) == 0 && !fs.Empty {
		return nil
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = fileName
	header.Size = int64(len(fs.Contents))
	header.Mode = int64(fs.Mode &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return nil
	}
	_, err := w.Write(fs.Contents)
	return err
}

// apply ensures that state of targetPath in fs matches fileState.
func (fs *FileState) apply(fileSystem afero.Fs, targetPath string, umask os.FileMode, actuator Actuator) error {
	fi, err := fileSystem.Stat(targetPath)
	var currentContents []byte
	switch {
	case err == nil && fi.Mode().IsRegular():
		if len(fs.Contents) == 0 && !fs.Empty {
			return actuator.RemoveAll(targetPath)
		}
		currentContents, err = afero.ReadFile(fileSystem, targetPath)
		if err != nil {
			return err
		}
		if !bytes.Equal(currentContents, fs.Contents) {
			break
		}
		if fi.Mode()&os.ModePerm != fs.Mode&^umask {
			if err := actuator.Chmod(targetPath, fs.Mode&^umask); err != nil {
				return err
			}
		}
		return nil
	case err == nil:
		if err := actuator.RemoveAll(targetPath); err != nil {
			return err
		}
	case os.IsNotExist(err):
	default:
		return err
	}
	if len(fs.Contents) == 0 && !fs.Empty {
		return nil
	}
	return actuator.WriteFile(targetPath, fs.Contents, fs.Mode&^umask, currentContents)
}

// SourceName implements Entry.SourceName.
func (fs *FileState) SourceName() string {
	return fs.sourceName
}

// NewRootState creates a new RootState.
func NewRootState(targetDir string, umask os.FileMode, sourceDir string, data map[string]interface{}) *RootState {
	return &RootState{
		TargetDir: targetDir,
		Umask:     umask,
		SourceDir: sourceDir,
		Data:      data,
		Entries:   make(map[string]Entry),
	}
}

// Add adds a new target.
func (rs *RootState) Add(fs afero.Fs, target string, fi os.FileInfo, addEmpty, addTemplate bool, actuator Actuator) error {
	if !filepath.HasPrefix(target, rs.TargetDir) {
		return errors.Errorf("%s: outside target directory", target)
	}
	targetName, err := filepath.Rel(rs.TargetDir, target)
	if err != nil {
		return err
	}
	if fi == nil {
		var err error
		fi, err = fs.Stat(target)
		if err != nil {
			return err
		}
	}

	// Add the parent directories, if needed.
	dirSourceName := ""
	entries := rs.Entries
	if parentDirName := filepath.Dir(targetName); parentDirName != "." {
		parentEntry, err := rs.findEntry(parentDirName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if parentEntry == nil {
			if err := rs.Add(fs, filepath.Join(rs.TargetDir, parentDirName), nil, false, false, actuator); err != nil {
				return err
			}
			parentEntry, err = rs.findEntry(parentDirName)
			if err != nil {
				return err
			}
		} else if _, ok := parentEntry.(*Dir); !ok {
			return errors.Errorf("%s: not a directory", parentDirName)
		}
		dir := parentEntry.(*Dir)
		dirSourceName = dir.sourceName
		entries = dir.Entries
	}

	name := filepath.Base(targetName)
	switch {
	case fi.Mode().IsRegular():
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*FileState); !ok {
				return errors.Errorf("%s: already added and not a regular file", targetName)
			}
			return nil // entry already exists
		}
		if fi.Size() == 0 && !addEmpty {
			return nil
		}
		sourceName := makeFileName(name, fi.Mode(), fi.Size() == 0, addTemplate)
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		contents, err := afero.ReadFile(fs, target)
		if err != nil {
			return err
		}
		if addTemplate {
			contents = autoTemplate(contents, rs.Data)
		}
		if err := actuator.WriteFile(filepath.Join(rs.SourceDir, sourceName), contents, 0666&^rs.Umask, nil); err != nil {
			return err
		}
		entries[name] = &FileState{
			sourceName: sourceName,
			Empty:      len(contents) == 0,
			Mode:       fi.Mode(),
			Contents:   contents,
		}
	case fi.Mode().IsDir():
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*Dir); !ok {
				return errors.Errorf("%s: already added and not a directory", targetName)
			}
			return nil // entry already exists
		}
		sourceName := makeDirName(name, fi.Mode())
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		if err := actuator.Mkdir(filepath.Join(rs.SourceDir, sourceName), 0777&^rs.Umask); err != nil {
			return err
		}
		// If the directory is empty, add a .keep file so the directory is
		// managed by git. Chezmoi will ignore the .keep file as it begins with
		// a dot.
		if stat, ok := fi.Sys().(*syscall.Stat_t); ok && stat.Nlink == 2 {
			if err := actuator.WriteFile(filepath.Join(rs.SourceDir, sourceName, ".keep"), nil, 0666&^rs.Umask, nil); err != nil {
				return err
			}
		}
		entries[name] = newDir(sourceName, fi.Mode())
	default:
		return errors.Errorf("%s: not a regular file or directory", targetName)
	}
	return nil
}

// AllEntries returns all the Entries in rs.
func (rs *RootState) AllEntries() map[string]Entry {
	result := make(map[string]Entry)
	for entryName, entry := range rs.Entries {
		entry.addAllEntries(result, entryName)
	}
	return result
}

// Archive writes rs to w.
func (rs *RootState) Archive(w *tar.Writer, umask os.FileMode) error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(currentUser.Gid)
	if err != nil {
		return err
	}
	group, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return err
	}
	now := time.Now()
	headerTemplate := tar.Header{
		Uid:        uid,
		Gid:        gid,
		Uname:      currentUser.Username,
		Gname:      group.Name,
		ModTime:    now,
		AccessTime: now,
		ChangeTime: now,
	}
	for _, entryName := range sortedEntryNames(rs.Entries) {
		if err := rs.Entries[entryName].archive(w, entryName, &headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}

// Apply ensures that targetDir in fs matches d.
func (rs *RootState) Apply(fs afero.Fs, actuator Actuator) error {
	for _, entryName := range sortedEntryNames(rs.Entries) {
		if err := rs.Entries[entryName].apply(fs, filepath.Join(rs.TargetDir, entryName), rs.Umask, actuator); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the state of the given target, or nil if no such target is found.
func (rs *RootState) Get(target string) (Entry, error) {
	if !filepath.HasPrefix(target, rs.TargetDir) {
		return nil, errors.Errorf("%s: outside target directory", target)
	}
	targetName, err := filepath.Rel(rs.TargetDir, target)
	if err != nil {
		return nil, err
	}
	return rs.findEntry(targetName)
}

// Populate walks fs from the source directory creating a target directory
// state.
func (rs *RootState) Populate(fs afero.Fs) error {
	return afero.Walk(fs, rs.SourceDir, func(path string, fi os.FileInfo, err error) error {
		relPath, err := filepath.Rel(rs.SourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		// Ignore all files and directories beginning with "."
		if _, name := filepath.Split(relPath); strings.HasPrefix(name, ".") {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case fi.Mode().IsRegular():
			dirNames, fileName, mode, isEmpty, isTemplate := parseFilePath(relPath)
			entries, err := rs.findEntries(dirNames)
			if err != nil {
				return err
			}
			contents, err := afero.ReadFile(fs, path)
			if err != nil {
				return err
			}
			if isTemplate {
				tmpl, err := template.New(path).Parse(string(contents))
				if err != nil {
					return errors.Wrap(err, path)
				}
				output := &bytes.Buffer{}
				if err := tmpl.Execute(output, rs.Data); err != nil {
					return errors.Wrap(err, path)
				}
				contents = output.Bytes()
			}
			entries[fileName] = &FileState{
				sourceName: relPath,
				Empty:      isEmpty,
				Mode:       mode,
				Contents:   contents,
			}
		case fi.Mode().IsDir():
			components := splitPathList(relPath)
			dirNames, modes := parseDirNameComponents(components)
			entries, err := rs.findEntries(dirNames[:len(dirNames)-1])
			if err != nil {
				return err
			}
			dirName := dirNames[len(dirNames)-1]
			mode := modes[len(modes)-1]
			entries[dirName] = newDir(relPath, mode)
		default:
			return errors.Errorf("unsupported file type: %s", path)
		}
		return nil
	})
}

func (rs *RootState) findEntries(dirNames []string) (map[string]Entry, error) {
	entries := rs.Entries
	for i, dirName := range dirNames {
		if entry, ok := entries[dirName]; !ok {
			return nil, os.ErrNotExist
		} else if dir, ok := entry.(*Dir); ok {
			entries = dir.Entries
		} else {
			return nil, errors.Errorf("%s: not a directory", filepath.Join(dirNames[:i+1]...))
		}
	}
	return entries, nil
}

func (rs *RootState) findEntry(name string) (Entry, error) {
	names := splitPathList(name)
	entries, err := rs.findEntries(names[:len(names)-1])
	if err != nil {
		return nil, err
	}
	return entries[names[len(names)-1]], nil
}

func makeDirName(name string, mode os.FileMode) string {
	dirName := ""
	if mode&os.FileMode(077) == os.FileMode(0) {
		dirName = privatePrefix
	}
	if strings.HasPrefix(name, ".") {
		dirName += dotPrefix + strings.TrimPrefix(name, ".")
	} else {
		dirName += name
	}
	return dirName
}

func makeFileName(name string, mode os.FileMode, isEmpty bool, isTemplate bool) string {
	fileName := ""
	if mode&os.FileMode(077) == os.FileMode(0) {
		fileName = privatePrefix
	}
	if isEmpty {
		fileName += emptyPrefix
	}
	if mode&os.FileMode(0111) != os.FileMode(0) {
		fileName += executablePrefix
	}
	if strings.HasPrefix(name, ".") {
		fileName += dotPrefix + strings.TrimPrefix(name, ".")
	} else {
		fileName += name
	}
	if isTemplate {
		fileName += templateSuffix
	}
	return fileName
}

// parseDirName parses a single directory name. It returns the target name,
// mode.
func parseDirName(dirName string) (string, os.FileMode) {
	name := dirName
	mode := os.FileMode(0777)
	if strings.HasPrefix(name, privatePrefix) {
		name = strings.TrimPrefix(name, privatePrefix)
		mode &= 0700
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	return name, mode
}

// parseFileName parses a single file name. It returns the target name, mode,
// whether the contents should be interpreted as a template, and any error.
func parseFileName(fileName string) (string, os.FileMode, bool, bool) {
	name := fileName
	mode := os.FileMode(0666)
	isPrivate := false
	isEmpty := false
	isTemplate := false
	if strings.HasPrefix(name, privatePrefix) {
		name = strings.TrimPrefix(name, privatePrefix)
		isPrivate = true
	}
	if strings.HasPrefix(name, emptyPrefix) {
		name = strings.TrimPrefix(name, emptyPrefix)
		isEmpty = true
	}
	if strings.HasPrefix(name, executablePrefix) {
		name = strings.TrimPrefix(name, executablePrefix)
		mode |= 0111
	}
	if strings.HasPrefix(name, dotPrefix) {
		name = "." + strings.TrimPrefix(name, dotPrefix)
	}
	if strings.HasSuffix(name, templateSuffix) {
		name = strings.TrimSuffix(name, templateSuffix)
		isTemplate = true
	}
	if isPrivate {
		mode &= 0700
	}
	return name, mode, isEmpty, isTemplate
}

// parseDirNameComponents parses multiple directory name components. It returns
// the target directory names, target modes, and any error.
func parseDirNameComponents(components []string) ([]string, []os.FileMode) {
	dirNames := []string{}
	modes := []os.FileMode{}
	for _, component := range components {
		dirName, mode := parseDirName(component)
		dirNames = append(dirNames, dirName)
		modes = append(modes, mode)
	}
	return dirNames, modes
}

// parseFilePath parses a single file path. It returns the target directory
// names, the target filename, the target mode, whether the contents should be
// interpreted as a template, and any error.
func parseFilePath(path string) ([]string, string, os.FileMode, bool, bool) {
	components := splitPathList(path)
	dirNames, _ := parseDirNameComponents(components[0 : len(components)-1])
	fileName, mode, isEmpty, isTemplate := parseFileName(components[len(components)-1])
	return dirNames, fileName, mode, isEmpty, isTemplate
}

// sortedEntryNames returns a sorted slice of all entry names.
func sortedEntryNames(entries map[string]Entry) []string {
	entryNames := []string{}
	for entryName := range entries {
		entryNames = append(entryNames, entryName)
	}
	sort.Strings(entryNames)
	return entryNames
}

func splitPathList(path string) []string {
	if strings.HasPrefix(path, string(filepath.Separator)) {
		path = strings.TrimPrefix(path, string(filepath.Separator))
	}
	return strings.Split(path, string(filepath.Separator))
}
