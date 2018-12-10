package chezmoi

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	vfs "github.com/twpayne/go-vfs"
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
	err error
}

// An Entry is either a Dir, a File, or a Symlink.
type Entry interface {
	Apply(fs vfs.FS, targetDir string, umask os.FileMode, mutator Mutator) error
	ConcreteValue(targetDir, sourceDir string, recursive bool) (interface{}, error)
	Evaluate() error
	SourceName() string
	TargetName() string
	archive(w *tar.Writer, headerTemplate *tar.Header, umask os.FileMode) error
}

// A Symlink represents the target state of a symlink.
type Symlink struct {
	sourceName       string
	targetName       string
	Template         bool
	linkName         string
	linkNameErr      error
	evaluateLinkName func() (string, error)
}

type symlinkConcreteValue struct {
	Type       string `json:"type" yaml:"type"`
	SourcePath string `json:"sourcePath" yaml:"sourcePath"`
	TargetPath string `json:"targetPath" yaml:"targetPath"`
	Template   bool   `json:"template" yaml:"template"`
	LinkName   string `json:"linkName" yaml:"linkName"`
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

// archive writes s to w.
func (s *Symlink) archive(w *tar.Writer, headerTemplate *tar.Header, umask os.FileMode) error {
	linkName, err := s.LinkName()
	if err != nil {
		return err
	}
	header := *headerTemplate
	header.Name = s.targetName
	header.Typeflag = tar.TypeSymlink
	header.Linkname = linkName
	return w.WriteHeader(&header)
}

// Apply ensures that the state of s's target in fs matches s.
func (s *Symlink) Apply(fs vfs.FS, targetDir string, umask os.FileMode, mutator Mutator) error {
	target, err := s.LinkName()
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
	return mutator.WriteSymlink(target, targetPath)
}

// ConcreteValue implements Entry.ConcreteValue.
func (s *Symlink) ConcreteValue(targetDir, sourceDir string, recursive bool) (interface{}, error) {
	linkName, err := s.LinkName()
	if err != nil {
		return nil, err
	}
	return &symlinkConcreteValue{
		Type:       "symlink",
		SourcePath: filepath.Join(sourceDir, s.SourceName()),
		TargetPath: filepath.Join(targetDir, s.TargetName()),
		Template:   s.Template,
		LinkName:   linkName,
	}, nil
}

// Evaluate evaluates s's target.
func (s *Symlink) Evaluate() error {
	_, err := s.LinkName()
	return err
}

// SourceName implements Entry.SourceName.
func (s *Symlink) SourceName() string {
	return s.sourceName
}

// LinkName returns s's link name.
func (s *Symlink) LinkName() (string, error) {
	if s.evaluateLinkName != nil {
		s.linkName, s.linkNameErr = s.evaluateLinkName()
		s.evaluateLinkName = nil
	}
	return s.linkName, s.linkNameErr
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
func (ts *TargetState) Add(fs vfs.FS, targetPath string, info os.FileInfo, addEmpty, addTemplate bool, mutator Mutator) error {
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
			if err := ts.Add(fs, filepath.Join(ts.TargetDir, parentDirName), nil, false, false, mutator); err != nil {
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
		if err := mutator.WriteFile(filepath.Join(ts.SourceDir, sourceName), data, 0666&^ts.Umask, nil); err != nil {
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
		if err := mutator.Mkdir(filepath.Join(ts.SourceDir, sourceName), 0777&^ts.Umask); err != nil {
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
			if err := mutator.WriteFile(filepath.Join(ts.SourceDir, sourceName, ".keep"), nil, 0666&^ts.Umask, nil); err != nil {
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
		if err := mutator.WriteFile(filepath.Join(ts.SourceDir, sourceName), []byte(data), 0666&^ts.Umask, nil); err != nil {
			return err
		}
		entries[name] = &Symlink{
			sourceName: sourceName,
			targetName: targetName,
			linkName:   data,
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
		if err := ts.Entries[entryName].archive(w, &headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}

// Apply ensures that ts.TargetDir in fs matches ts.
func (ts *TargetState) Apply(fs vfs.FS, mutator Mutator) error {
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].Apply(fs, ts.TargetDir, ts.Umask, mutator); err != nil {
			return err
		}
	}
	return nil
}

// ConcreteValue returns a value suitable for serialization.
func (ts *TargetState) ConcreteValue(recursive bool) (interface{}, error) {
	var entryConcreteValues []interface{}
	for _, entryName := range sortedEntryNames(ts.Entries) {
		entryConcreteValue, err := ts.Entries[entryName].ConcreteValue(ts.TargetDir, ts.SourceDir, recursive)
		if err != nil {
			return nil, err
		}
		entryConcreteValues = append(entryConcreteValues, entryConcreteValue)
	}
	return entryConcreteValues, nil
}

// Evaluate evaluates all of the entries in ts.
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

func (ts *TargetState) AddArchive(r *tar.Reader, destinationDir string, stripComponents int, mutator Mutator) error {
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeSymlink:
			if err := ts.addArchiveHeader(r, header, destinationDir, stripComponents, mutator); err != nil {
				return err
			}
		case tar.TypeXGlobalHeader:
		default:
			return fmt.Errorf("%s: unspported typeflag '%c'", header.Name, header.Typeflag)
		}
	}
	return nil
}

// Populate walks fs from ts.SourceDir to populate ts.
func (ts *TargetState) Populate(fs vfs.FS) error {
	return vfs.Walk(fs, ts.SourceDir, func(path string, info os.FileInfo, _ error) error {
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
				evaluateLinkName := func() (string, error) {
					data, err := fs.ReadFile(path)
					return string(data), err
				}
				if psfp.Template {
					evaluateLinkName = func() (string, error) {
						data, err := ts.executeTemplate(fs, path)
						return string(data), err
					}
				}
				entry = &Symlink{
					sourceName:       relPath,
					targetName:       targetName,
					Template:         psfp.Template,
					evaluateLinkName: evaluateLinkName,
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

func (ts *TargetState) addArchiveHeader(r *tar.Reader, header *tar.Header, destinationDir string, stripComponents int, mutator Mutator) error {
	targetPath := header.Name
	if stripComponents > 0 {
		targetPath = filepath.Join(strings.Split(targetPath, string(os.PathSeparator))[stripComponents:]...)
	}
	if destinationDir != "" {
		targetPath = filepath.Join(destinationDir, targetPath)
	} else {
		targetPath = filepath.Join(ts.TargetDir, targetPath)
	}
	targetName, err := filepath.Rel(ts.TargetDir, targetPath)
	if err != nil {
		return err
	}
	parentDirSourceName := ""
	entries := ts.Entries
	if parentDirName := filepath.Dir(targetName); parentDirName != "." {
		parentEntry, err := ts.findEntry(parentDirName)
		if err != nil {
			return err
		}
		parentDir, ok := parentEntry.(*Dir)
		if !ok {
			return fmt.Errorf("%s: parent is not a directory", targetName)
		}
		parentDirSourceName = parentDir.sourceName
		entries = parentDir.Entries
	}
	name := filepath.Base(targetName)
	switch header.Typeflag {
	case tar.TypeReg:
		var existingFile *File
		var existingContents []byte
		if entry, ok := entries[name]; ok {
			existingFile, ok = entry.(*File)
			if !ok {
				return fmt.Errorf("%s: already added and not a regular file", targetName)
			}
			existingContents, err = existingFile.Contents()
			if err != nil {
				return err
			}
		}
		perm := os.FileMode(header.Mode) & os.ModePerm
		empty := header.Size == 0
		sourceName := ParsedSourceFileName{
			FileName: name,
			Mode:     perm,
			Empty:    empty,
		}.SourceFileName()
		if parentDirSourceName != "" {
			sourceName = filepath.Join(parentDirSourceName, sourceName)
		}
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		file := &File{
			sourceName: sourceName,
			targetName: targetName,
			Empty:      empty,
			Perm:       perm,
			Template:   false,
			contents:   contents,
		}
		if existingFile != nil {
			if bytes.Equal(existingFile.contents, file.contents) {
				if existingFile.sourceName == file.sourceName {
					return nil
				}
				return mutator.Rename(filepath.Join(ts.SourceDir, existingFile.sourceName), filepath.Join(ts.SourceDir, file.sourceName))
			}
			if err := mutator.RemoveAll(filepath.Join(ts.SourceDir, existingFile.sourceName)); err != nil {
				return err
			}
		}
		entries[name] = file
		return mutator.WriteFile(filepath.Join(ts.SourceDir, sourceName), contents, 0666&^ts.Umask, existingContents)
	case tar.TypeDir:
		var existingDir *Dir
		if entry, ok := entries[name]; ok {
			existingDir, ok = entry.(*Dir)
			if !ok {
				return fmt.Errorf("%s: already added and not a directory", targetName)
			}
		}
		perm := os.FileMode(header.Mode) & os.ModePerm
		sourceName := ParsedSourceDirName{
			DirName: name,
			Perm:    perm,
		}.SourceDirName()
		if parentDirSourceName != "" {
			sourceName = filepath.Join(parentDirSourceName, sourceName)
		}
		dir := newDir(sourceName, targetName, perm)
		if existingDir != nil {
			if existingDir.sourceName == dir.sourceName {
				return nil
			}
			return mutator.Rename(filepath.Join(ts.SourceDir, existingDir.sourceName), filepath.Join(ts.SourceDir, dir.sourceName))
		}
		// FIXME Add a .keep file if the directory is empty
		entries[name] = dir
		return mutator.Mkdir(filepath.Join(ts.SourceDir, sourceName), 0777&^ts.Umask)
	case tar.TypeSymlink:
		var existingSymlink *Symlink
		var existingLinkName string
		if entry, ok := entries[name]; ok {
			existingSymlink, ok = entry.(*Symlink)
			if !ok {
				return fmt.Errorf("%s: already added and not a symlink", targetName)
			}
			existingLinkName, err = existingSymlink.LinkName()
			if err != nil {
				return err
			}
		}
		sourceName := ParsedSourceFileName{
			FileName: name,
			Mode:     os.ModeSymlink,
		}.SourceFileName()
		if parentDirSourceName != "" {
			sourceName = filepath.Join(parentDirSourceName, sourceName)
		}
		symlink := &Symlink{
			sourceName: sourceName,
			targetName: targetName,
			linkName:   header.Linkname,
		}
		if existingSymlink != nil {
			if existingSymlink.linkName == symlink.linkName {
				if existingSymlink.sourceName == symlink.sourceName {
					return nil
				}
				return mutator.Rename(filepath.Join(ts.SourceDir, existingSymlink.sourceName), filepath.Join(ts.SourceDir, symlink.sourceName))
			}
			if err := mutator.RemoveAll(filepath.Join(ts.SourceDir, existingSymlink.sourceName)); err != nil {
				return err
			}
		}
		entries[name] = symlink
		return mutator.WriteFile(filepath.Join(ts.SourceDir, symlink.sourceName), []byte(symlink.linkName), 0666&^ts.Umask, []byte(existingLinkName))
	default:
		return fmt.Errorf("%s: unspported typeflag '%c'", header.Name, header.Typeflag)
	}
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
