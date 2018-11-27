package chezmoi

// FIXME add Symlink

import (
	"archive/tar"
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/twpayne/go-vfs"
)

const (
	privatePrefix    = "private_"
	emptyPrefix      = "empty_"
	executablePrefix = "executable_"
	dotPrefix        = "dot_"
	templateSuffix   = ".tmpl"
)

// An Entry is either a Dir or a File.
type Entry interface {
	SourceName() string
	addAllEntries(map[string]Entry, string)
	apply(vfs.FS, string, os.FileMode, Actuator) error
	archive(*tar.Writer, string, *tar.Header, os.FileMode) error
}

// A File represents the target state of a file.
type File struct {
	sourceName string
	Empty      bool
	Perm       os.FileMode
	Contents   []byte
}

// A Dir represents the target state of a directory.
type Dir struct {
	sourceName string
	Perm       os.FileMode
	Entries    map[string]Entry
}

// A TargetState represents the root target state.
type TargetState struct {
	TargetDir string
	Umask     os.FileMode
	SourceDir string
	Data      map[string]interface{}
	Entries   map[string]Entry
}

// newDir returns a new directory state.
func newDir(sourceName string, perm os.FileMode) *Dir {
	return &Dir{
		sourceName: sourceName,
		Perm:       perm,
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
	header.Mode = int64(d.Perm &^ umask & os.ModePerm)
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
func (d *Dir) apply(fs vfs.FS, targetDir string, umask os.FileMode, actuator Actuator) error {
	info, err := fs.Stat(targetDir)
	switch {
	case err == nil && info.Mode().IsDir():
		if info.Mode()&os.ModePerm != d.Perm&^umask {
			if err := actuator.Chmod(targetDir, d.Perm&^umask); err != nil {
				return err
			}
		}
	case err == nil:
		if err := actuator.RemoveAll(targetDir); err != nil {
			return err
		}
		fallthrough
	case os.IsNotExist(err):
		if err := actuator.Mkdir(targetDir, d.Perm&^umask); err != nil {
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

// addAllEntries adds f to result.
func (f *File) addAllEntries(result map[string]Entry, name string) {
	result[name] = f
}

// archive writes f to w.
func (f *File) archive(w *tar.Writer, fileName string, headerTemplate *tar.Header, umask os.FileMode) error {
	if len(f.Contents) == 0 && !f.Empty {
		return nil
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = fileName
	header.Size = int64(len(f.Contents))
	header.Mode = int64(f.Perm &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return nil
	}
	_, err := w.Write(f.Contents)
	return err
}

// apply ensures that state of targetPath in fs matches f.
func (f *File) apply(fs vfs.FS, targetPath string, umask os.FileMode, actuator Actuator) error {
	info, err := fs.Stat(targetPath)
	var currentContents []byte
	switch {
	case err == nil && info.Mode().IsRegular():
		if len(f.Contents) == 0 && !f.Empty {
			return actuator.RemoveAll(targetPath)
		}
		currentContents, err = fs.ReadFile(targetPath)
		if err != nil {
			return err
		}
		if !bytes.Equal(currentContents, f.Contents) {
			break
		}
		if info.Mode()&os.ModePerm != f.Perm&^umask {
			if err := actuator.Chmod(targetPath, f.Perm&^umask); err != nil {
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
	if len(f.Contents) == 0 && !f.Empty {
		return nil
	}
	return actuator.WriteFile(targetPath, f.Contents, f.Perm&^umask, currentContents)
}

// SourceName implements Entry.SourceName.
func (f *File) SourceName() string {
	return f.sourceName
}

// NewTargetState creates a new TargetState.
func NewTargetState(targetDir string, umask os.FileMode, sourceDir string, data map[string]interface{}) *TargetState {
	return &TargetState{
		TargetDir: targetDir,
		Umask:     umask,
		SourceDir: sourceDir,
		Data:      data,
		Entries:   make(map[string]Entry),
	}
}

// Add adds a new target to ts.
func (ts *TargetState) Add(fs vfs.FS, target string, info os.FileInfo, addEmpty, addTemplate bool, actuator Actuator) error {
	if !filepath.HasPrefix(target, ts.TargetDir) {
		return fmt.Errorf("%s: outside target directory", target)
	}
	targetName, err := filepath.Rel(ts.TargetDir, target)
	if err != nil {
		return err
	}
	if info == nil {
		var err error
		info, err = fs.Stat(target)
		if err != nil {
			return err
		}
	}

	// Add the parent directories, if needed.
	dirSourceName := ""
	entries := ts.Entries
	if parentDirName := filepath.Dir(targetName); parentDirName != "." {
		parentEntry, err := ts.findEntry(parentDirName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if parentEntry == nil {
			if err := ts.Add(fs, filepath.Join(ts.TargetDir, parentDirName), nil, false, false, actuator); err != nil {
				return err
			}
			parentEntry, err = ts.findEntry(parentDirName)
			if err != nil {
				return err
			}
		} else if _, ok := parentEntry.(*Dir); !ok {
			return fmt.Errorf("%s: not a directory", parentDirName)
		}
		dir := parentEntry.(*Dir)
		dirSourceName = dir.sourceName
		entries = dir.Entries
	}

	name := filepath.Base(targetName)
	switch {
	case info.Mode().IsRegular():
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*File); !ok {
				return fmt.Errorf("%s: already added and not a regular file", targetName)
			}
			return nil // entry already exists
		}
		if info.Size() == 0 && !addEmpty {
			return nil
		}
		sourceName := makeFileName(name, info.Mode(), info.Size() == 0, addTemplate)
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		contents, err := fs.ReadFile(target)
		if err != nil {
			return err
		}
		if addTemplate {
			contents = autoTemplate(contents, ts.Data)
		}
		if err := actuator.WriteFile(filepath.Join(ts.SourceDir, sourceName), contents, 0666&^ts.Umask, nil); err != nil {
			return err
		}
		entries[name] = &File{
			sourceName: sourceName,
			Empty:      len(contents) == 0,
			Perm:       info.Mode() & os.ModePerm,
			Contents:   contents,
		}
	case info.Mode().IsDir():
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*Dir); !ok {
				return fmt.Errorf("%s: already added and not a directory", targetName)
			}
			return nil // entry already exists
		}
		sourceName := makeDirName(name, info.Mode())
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		if err := actuator.Mkdir(filepath.Join(ts.SourceDir, sourceName), 0777&^ts.Umask); err != nil {
			return err
		}
		// If the directory is empty, add a .keep file so the directory is
		// managed by git. Chezmoi will ignore the .keep file as it begins with
		// a dot.
		if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat.Nlink == 2 {
			if err := actuator.WriteFile(filepath.Join(ts.SourceDir, sourceName, ".keep"), nil, 0666&^ts.Umask, nil); err != nil {
				return err
			}
		}
		entries[name] = newDir(sourceName, info.Mode()&os.ModePerm)
	default:
		return fmt.Errorf("%s: not a regular file or directory", targetName)
	}
	return nil
}

// AllEntries returns all the Entries in ts.
func (ts *TargetState) AllEntries() map[string]Entry {
	result := make(map[string]Entry)
	for entryName, entry := range ts.Entries {
		entry.addAllEntries(result, entryName)
	}
	return result
}

// Archive writes ts to w.
func (ts *TargetState) Archive(w *tar.Writer, umask os.FileMode) error {
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
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].archive(w, entryName, &headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}

// Apply ensures that ts.TargetDir in fs matches ts.
func (ts *TargetState) Apply(fs vfs.FS, actuator Actuator) error {
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].apply(fs, filepath.Join(ts.TargetDir, entryName), ts.Umask, actuator); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the state of the given target, or nil if no such target is found.
func (ts *TargetState) Get(target string) (Entry, error) {
	if !filepath.HasPrefix(target, ts.TargetDir) {
		return nil, fmt.Errorf("%s: outside target directory", target)
	}
	targetName, err := filepath.Rel(ts.TargetDir, target)
	if err != nil {
		return nil, err
	}
	return ts.findEntry(targetName)
}

// Populate walks fs from ts.SourceDir to populate ts.
func (ts *TargetState) Populate(fs vfs.FS) error {
	return vfs.Walk(fs, ts.SourceDir, func(path string, info os.FileInfo, err error) error {
		relPath, err := filepath.Rel(ts.SourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		// Ignore all files and directories beginning with "."
		if _, name := filepath.Split(relPath); strings.HasPrefix(name, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case info.Mode().IsRegular():
			dirNames, fileName, mode, isEmpty, isTemplate := parseFilePath(relPath)
			entries, err := ts.findEntries(dirNames)
			if err != nil {
				return err
			}
			contents, err := fs.ReadFile(path)
			if err != nil {
				return err
			}
			if isTemplate {
				tmpl, err := template.New(path).Parse(string(contents))
				if err != nil {
					return fmt.Errorf("%s: %v", path, err)
				}
				output := &bytes.Buffer{}
				if err := tmpl.Execute(output, ts.Data); err != nil {
					return fmt.Errorf("%s: %v", path, err)
				}
				contents = output.Bytes()
			}
			entries[fileName] = &File{
				sourceName: relPath,
				Empty:      isEmpty,
				Perm:       mode,
				Contents:   contents,
			}
		case info.Mode().IsDir():
			components := splitPathList(relPath)
			dirNames, modes := parseDirNameComponents(components)
			entries, err := ts.findEntries(dirNames[:len(dirNames)-1])
			if err != nil {
				return err
			}
			dirName := dirNames[len(dirNames)-1]
			mode := modes[len(modes)-1]
			entries[dirName] = newDir(relPath, mode)
		default:
			return fmt.Errorf("unsupported file type: %s", path)
		}
		return nil
	})
}

func (ts *TargetState) findEntries(dirNames []string) (map[string]Entry, error) {
	entries := ts.Entries
	for i, dirName := range dirNames {
		if entry, ok := entries[dirName]; !ok {
			return nil, os.ErrNotExist
		} else if dir, ok := entry.(*Dir); ok {
			entries = dir.Entries
		} else {
			return nil, fmt.Errorf("%s: not a directory", filepath.Join(dirNames[:i+1]...))
		}
	}
	return entries, nil
}

func (ts *TargetState) findEntry(name string) (Entry, error) {
	names := splitPathList(name)
	entries, err := ts.findEntries(names[:len(names)-1])
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
