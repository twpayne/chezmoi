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
	"text/template"
	"time"

	"github.com/twpayne/go-vfs"
)

const (
	symlinkPrefix    = "symlink_"
	privatePrefix    = "private_"
	emptyPrefix      = "empty_"
	executablePrefix = "executable_"
	dotPrefix        = "dot_"
	templateSuffix   = ".tmpl"
)

// A templateFuncError is an error encountered while executing a template
// function.
type templateFuncError struct {
	name string
	err  error
}

// An Entry is either a Dir, a File, or a Symlink.
type Entry interface {
	Apply(fs vfs.FS, targetDir string, umask os.FileMode, actuator Actuator) error
	Evaluate() error
	SourceName() string
	TargetName() string
	archive(*tar.Writer, string, *tar.Header, os.FileMode) error
}

// A File represents the target state of a file.
type File struct {
	sourceName       string
	targetName       string
	Empty            bool
	Perm             os.FileMode
	Template         bool
	contents         []byte
	contentsErr      error
	evaluateContents func() ([]byte, error)
}

// A Dir represents the target state of a directory.
type Dir struct {
	sourceName string
	targetName string
	Perm       os.FileMode
	Entries    map[string]Entry
}

// A Symlink represents the target state of a symlink.
type Symlink struct {
	sourceName     string
	targetName     string
	Template       bool
	target         string
	targetErr      error
	evaluateTarget func() (string, error)
}

// A TargetState represents the root target state.
type TargetState struct {
	TargetDir string
	Umask     os.FileMode
	SourceDir string
	Data      map[string]interface{}
	Funcs     template.FuncMap
	Entries   map[string]Entry
}

// ParsedSourceDirName is a parsed source dir name.
type ParsedSourceDirName struct {
	DirName string
	Perm    os.FileMode
}

// A ParsedSourceFileName is a parsed source file name.
type ParsedSourceFileName struct {
	FileName string
	Mode     os.FileMode
	Empty    bool
	Template bool
}

type parsedSourceFilePath struct {
	ParsedSourceFileName
	dirNames []string
}

// newDir returns a new directory state.
func newDir(sourceName string, targetName string, perm os.FileMode) *Dir {
	return &Dir{
		sourceName: sourceName,
		targetName: targetName,
		Perm:       perm,
		Entries:    make(map[string]Entry),
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

// Apply ensures that targetDir in fs matches d.
func (d *Dir) Apply(fs vfs.FS, targetDir string, umask os.FileMode, actuator Actuator) error {
	targetPath := filepath.Join(targetDir, d.targetName)
	info, err := fs.Lstat(targetPath)
	switch {
	case err == nil && info.Mode().IsDir():
		if info.Mode()&os.ModePerm != d.Perm&^umask {
			if err := actuator.Chmod(targetPath, d.Perm&^umask); err != nil {
				return err
			}
		}
	case err == nil:
		if err := actuator.RemoveAll(targetPath); err != nil {
			return err
		}
		fallthrough
	case os.IsNotExist(err):
		if err := actuator.Mkdir(targetPath, d.Perm&^umask); err != nil {
			return err
		}
	default:
		return err
	}
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].Apply(fs, targetDir, umask, actuator); err != nil {
			return err
		}
	}
	return nil
}

// Evaluate evaluates all entries in d.
func (d *Dir) Evaluate() error {
	for _, entryName := range sortedEntryNames(d.Entries) {
		if err := d.Entries[entryName].Evaluate(); err != nil {
			return err
		}
	}
	return nil
}

// Private returns true if d is private.
func (d *Dir) Private() bool {
	return d.Perm&os.ModePerm&077 == 0
}

// SourceName implements Entry.SourceName.
func (d *Dir) SourceName() string {
	return d.sourceName
}

// TargetName implements Entry.TargetName.
func (d *Dir) TargetName() string {
	return d.targetName
}

// archive writes f to w.
func (f *File) archive(w *tar.Writer, fileName string, headerTemplate *tar.Header, umask os.FileMode) error {
	contents, err := f.Contents()
	if err != nil {
		return err
	}
	if len(contents) == 0 && !f.Empty {
		return nil
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeReg
	header.Name = fileName
	header.Size = int64(len(contents))
	header.Mode = int64(f.Perm &^ umask)
	if err := w.WriteHeader(&header); err != nil {
		return nil
	}
	_, err = w.Write(contents)
	return err
}

// Apply ensures that the state of targetPath in fs matches f.
func (f *File) Apply(fs vfs.FS, targetDir string, umask os.FileMode, actuator Actuator) error {
	contents, err := f.Contents()
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, f.targetName)
	info, err := fs.Lstat(targetPath)
	var currData []byte
	switch {
	case err == nil && info.Mode().IsRegular():
		if len(contents) == 0 && !f.Empty {
			return actuator.RemoveAll(targetPath)
		}
		currData, err = fs.ReadFile(targetPath)
		if err != nil {
			return err
		}
		if !bytes.Equal(currData, contents) {
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
	if len(contents) == 0 && !f.Empty {
		return nil
	}
	return actuator.WriteFile(targetPath, contents, f.Perm&^umask, currData)
}

// Evaluate evaluates f's contents.
func (f *File) Evaluate() error {
	_, err := f.Contents()
	return err
}

// Contents returns f's contents.
func (f *File) Contents() ([]byte, error) {
	if f.evaluateContents != nil {
		f.contents, f.contentsErr = f.evaluateContents()
		f.evaluateContents = nil
	}
	return f.contents, f.contentsErr
}

// Executable returns true is f is executable.
func (f *File) Executable() bool {
	return f.Perm&0111 != 0
}

// Private returns true if f is private.
func (f *File) Private() bool {
	return f.Perm&os.ModePerm&077 == 0
}

// SourceName implements Entry.SourceName.
func (f *File) SourceName() string {
	return f.sourceName
}

// TargetName implements Entry.TargetName.
func (f *File) TargetName() string {
	return f.targetName
}

// archive writes s to w.
func (s *Symlink) archive(w *tar.Writer, dirName string, headerTemplate *tar.Header, umask os.FileMode) error {
	target, err := s.Target()
	if err != nil {
		return err
	}
	header := *headerTemplate
	header.Typeflag = tar.TypeSymlink
	header.Linkname = target
	return w.WriteHeader(&header)
}

// Apply ensures that the state of s's target in fs matches s.
func (s *Symlink) Apply(fs vfs.FS, targetDir string, umask os.FileMode, actuator Actuator) error {
	target, err := s.Target()
	if err != nil {
		return err
	}
	targetPath := filepath.Join(targetDir, s.targetName)
	info, err := fs.Lstat(targetPath)
	switch {
	case err == nil && info.Mode()&os.ModeType == os.ModeSymlink:
		currentTarget, err := fs.Readlink(targetPath)
		if err != nil {
			return err
		}
		if currentTarget == target {
			return nil
		}
	case err == nil:
	case os.IsNotExist(err):
	default:
		return err
	}
	return actuator.WriteSymlink(target, targetPath)
}

// Evaluate evaluates s's target.
func (s *Symlink) Evaluate() error {
	_, err := s.Target()
	return err
}

func (s *Symlink) SourceName() string {
	return s.sourceName
}

// Target returns f's contents.
func (s *Symlink) Target() (string, error) {
	if s.evaluateTarget != nil {
		s.target, s.targetErr = s.evaluateTarget()
		s.evaluateTarget = nil
	}
	return s.target, s.targetErr
}

// TargetName implements Entry.TargetName.
func (s *Symlink) TargetName() string {
	return s.targetName
}

// NewTargetState creates a new TargetState.
func NewTargetState(targetDir string, umask os.FileMode, sourceDir string, data map[string]interface{}, funcs template.FuncMap) *TargetState {
	return &TargetState{
		TargetDir: targetDir,
		Umask:     umask,
		SourceDir: sourceDir,
		Data:      data,
		Funcs:     funcs,
		Entries:   make(map[string]Entry),
	}
}

// Add adds a new target to ts.
func (ts *TargetState) Add(fs vfs.FS, targetPath string, info os.FileInfo, addEmpty, addTemplate bool, actuator Actuator) error {
	if !filepath.HasPrefix(targetPath, ts.TargetDir) {
		return fmt.Errorf("%s: outside target directory", targetPath)
	}
	targetName, err := filepath.Rel(ts.TargetDir, targetPath)
	if err != nil {
		return err
	}
	if info == nil {
		var err error
		info, err = fs.Lstat(targetPath)
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
		sourceName := ParsedSourceFileName{
			FileName: name,
			Mode:     info.Mode() & os.ModePerm,
			Empty:    info.Size() == 0,
			Template: addTemplate,
		}.SourceFileName()
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		data, err := fs.ReadFile(targetPath)
		if err != nil {
			return err
		}
		if addTemplate {
			data = autoTemplate(data, ts.Data)
		}
		if err := actuator.WriteFile(filepath.Join(ts.SourceDir, sourceName), data, 0666&^ts.Umask, nil); err != nil {
			return err
		}
		entries[name] = &File{
			sourceName: sourceName,
			targetName: targetName,
			Empty:      len(data) == 0,
			Perm:       info.Mode() & os.ModePerm,
			Template:   addTemplate,
			contents:   data,
		}
	case info.Mode().IsDir():
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*Dir); !ok {
				return fmt.Errorf("%s: already added and not a directory", targetName)
			}
			return nil // entry already exists
		}
		sourceName := ParsedSourceDirName{
			DirName: name,
			Perm:    info.Mode() & os.ModePerm,
		}.SourceDirName()
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		if err := actuator.Mkdir(filepath.Join(ts.SourceDir, sourceName), 0777&^ts.Umask); err != nil {
			return err
		}
		// If the directory is empty, add a .keep file so the directory is
		// managed by git. Chezmoi will ignore the .keep file as it begins with
		// a dot.
		infos, err := fs.ReadDir(targetPath)
		if err != nil {
			return err
		}
		if len(infos) == 0 {
			if err := actuator.WriteFile(filepath.Join(ts.SourceDir, sourceName, ".keep"), nil, 0666&^ts.Umask, nil); err != nil {
				return err
			}
		}
		entries[name] = newDir(sourceName, targetName, info.Mode()&os.ModePerm)
	case info.Mode()&os.ModeType == os.ModeSymlink:
		if entry, ok := entries[name]; ok {
			if _, ok := entry.(*Symlink); !ok {
				return fmt.Errorf("%s: already added and not a symlink", targetName)
			}
			return nil // entry already exists
		}
		sourceName := ParsedSourceFileName{
			FileName: name,
			Mode:     os.ModeSymlink,
		}.SourceFileName()
		if dirSourceName != "" {
			sourceName = filepath.Join(dirSourceName, sourceName)
		}
		data, err := fs.Readlink(targetPath)
		if err != nil {
			return err
		}
		if err := actuator.WriteFile(filepath.Join(ts.SourceDir, sourceName), []byte(data), 0666&^ts.Umask, nil); err != nil {
			return err
		}
		entries[name] = &Symlink{
			sourceName: sourceName,
			targetName: targetName,
			target:     data,
		}
	default:
		return fmt.Errorf("%s: not a regular file or directory", targetName)
	}
	return nil
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
		if err := ts.Entries[entryName].Apply(fs, ts.TargetDir, ts.Umask, actuator); err != nil {
			return err
		}
	}
	return nil
}

// Evaluates all of the entries in ts.
func (ts *TargetState) Evaluate() error {
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].Evaluate(); err != nil {
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
			psfp := parseSourceFilePath(relPath)
			entries, err := ts.findEntries(psfp.dirNames)
			if err != nil {
				return err
			}

			targetName := filepath.Join(append(psfp.dirNames, psfp.FileName)...)
			var entry Entry
			switch psfp.Mode & os.ModeType {
			case 0:
				evaluateContents := func() ([]byte, error) {
					return fs.ReadFile(path)
				}
				if psfp.Template {
					evaluateContents = func() ([]byte, error) {
						return ts.executeTemplate(fs, path)
					}
				}
				entry = &File{
					sourceName:       relPath,
					targetName:       targetName,
					Empty:            psfp.Empty,
					Perm:             psfp.Mode & os.ModePerm,
					Template:         psfp.Template,
					evaluateContents: evaluateContents,
				}
			case os.ModeSymlink:
				evaluateTarget := func() (string, error) {
					data, err := fs.ReadFile(path)
					return string(data), err
				}
				if psfp.Template {
					evaluateTarget = func() (string, error) {
						data, err := ts.executeTemplate(fs, path)
						return string(data), err
					}
				}
				entry = &Symlink{
					sourceName:     relPath,
					targetName:     targetName,
					Template:       psfp.Template,
					evaluateTarget: evaluateTarget,
				}
			default:
				return fmt.Errorf("%v: unsupported mode: %d", path, psfp.Mode&os.ModeType)
			}
			entries[psfp.FileName] = entry
		case info.Mode().IsDir():
			components := splitPathList(relPath)
			dirNames, perms := parseDirNameComponents(components)
			targetName := filepath.Join(dirNames...)
			entries, err := ts.findEntries(dirNames[:len(dirNames)-1])
			if err != nil {
				return err
			}
			dirName := dirNames[len(dirNames)-1]
			perm := perms[len(perms)-1]
			entries[dirName] = newDir(relPath, targetName, perm)
		default:
			return fmt.Errorf("unsupported file type: %s", path)
		}
		return nil
	})
}

func (ts *TargetState) executeTemplate(fs vfs.FS, path string) ([]byte, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ts.executeTemplateData(path, data)
}

func (ts *TargetState) executeTemplateData(name string, data []byte) (_ []byte, err error) {
	tmpl, err := template.New(name).Option("missingkey=error").Funcs(ts.Funcs).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	defer func() {
		if r := recover(); r != nil {
			if tfe, ok := r.(templateFuncError); ok {
				err = tfe.err
			} else {
				panic(r)
			}
		}
	}()
	output := &bytes.Buffer{}
	if err = tmpl.Execute(output, ts.Data); err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	return output.Bytes(), nil
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

// ReturnTemplateFuncError causes template execution to return an error.
func ReturnTemplateFuncError(err error) {
	panic(templateFuncError{
		err: err,
	})
}

// ParseSourceDirName parses a single directory name.
func ParseSourceDirName(dirName string) ParsedSourceDirName {
	perm := os.FileMode(0777)
	if strings.HasPrefix(dirName, privatePrefix) {
		dirName = strings.TrimPrefix(dirName, privatePrefix)
		perm &= 0700
	}
	if strings.HasPrefix(dirName, dotPrefix) {
		dirName = "." + strings.TrimPrefix(dirName, dotPrefix)
	}
	return ParsedSourceDirName{
		DirName: dirName,
		Perm:    perm,
	}
}

// SourceDirName returns psdn's source dir name.
func (psdn ParsedSourceDirName) SourceDirName() string {
	sourceDirName := ""
	if psdn.Perm&os.FileMode(077) == os.FileMode(0) {
		sourceDirName = privatePrefix
	}
	if strings.HasPrefix(psdn.DirName, ".") {
		sourceDirName += dotPrefix + strings.TrimPrefix(psdn.DirName, ".")
	} else {
		sourceDirName += psdn.DirName
	}
	return sourceDirName
}

// ParseSourceFileName parses a source file name.
func ParseSourceFileName(fileName string) ParsedSourceFileName {
	mode := os.FileMode(0666)
	empty := false
	template := false
	if strings.HasPrefix(fileName, symlinkPrefix) {
		fileName = strings.TrimPrefix(fileName, symlinkPrefix)
		mode |= os.ModeSymlink
	} else {
		private := false
		if strings.HasPrefix(fileName, privatePrefix) {
			fileName = strings.TrimPrefix(fileName, privatePrefix)
			private = true
		}
		if strings.HasPrefix(fileName, emptyPrefix) {
			fileName = strings.TrimPrefix(fileName, emptyPrefix)
			empty = true
		}
		if strings.HasPrefix(fileName, executablePrefix) {
			fileName = strings.TrimPrefix(fileName, executablePrefix)
			mode |= 0111
		}
		if private {
			mode &= 0700
		}
	}
	if strings.HasPrefix(fileName, dotPrefix) {
		fileName = "." + strings.TrimPrefix(fileName, dotPrefix)
	}
	if strings.HasSuffix(fileName, templateSuffix) {
		fileName = strings.TrimSuffix(fileName, templateSuffix)
		template = true
	}
	return ParsedSourceFileName{
		FileName: fileName,
		Mode:     mode,
		Empty:    empty,
		Template: template,
	}
}

// SourceFileName returns psfn's source file name.
func (psfn ParsedSourceFileName) SourceFileName() string {
	fileName := ""
	switch psfn.Mode & os.ModeType {
	case 0:
		if psfn.Mode&os.ModePerm&os.FileMode(077) == os.FileMode(0) {
			fileName = privatePrefix
		}
		if psfn.Empty {
			fileName += emptyPrefix
		}
		if psfn.Mode&os.ModePerm&os.FileMode(0111) != os.FileMode(0) {
			fileName += executablePrefix
		}
	case os.ModeSymlink:
		fileName = symlinkPrefix
	default:
		panic(fmt.Sprintf("%+v: unsupported type", psfn)) // FIXME return error instead of panicing
	}
	if strings.HasPrefix(psfn.FileName, ".") {
		fileName += dotPrefix + strings.TrimPrefix(psfn.FileName, ".")
	} else {
		fileName += psfn.FileName
	}
	if psfn.Template {
		fileName += templateSuffix
	}
	return fileName
}

// parseDirNameComponents parses multiple directory name components. It returns
// the target directory names, target permissions, and any error.
func parseDirNameComponents(components []string) ([]string, []os.FileMode) {
	dirNames := []string{}
	perms := []os.FileMode{}
	for _, component := range components {
		psdn := ParseSourceDirName(component)
		dirNames = append(dirNames, psdn.DirName)
		perms = append(perms, psdn.Perm)
	}
	return dirNames, perms
}

// parseSourceFilePath parses a single source file path.
func parseSourceFilePath(path string) parsedSourceFilePath {
	components := splitPathList(path)
	dirNames, _ := parseDirNameComponents(components[0 : len(components)-1])
	psfn := ParseSourceFileName(components[len(components)-1])
	return parsedSourceFilePath{
		ParsedSourceFileName: psfn,
		dirNames:             dirNames,
	}
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
