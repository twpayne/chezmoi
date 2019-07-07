package chezmoi

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs"
)

const (
	ignoreName       = ".chezmoiignore"
	removeName       = ".chezmoiremove"
	templatesDirName = ".chezmoitemplates"
	versionName      = ".chezmoiversion"
)

// An AddOptions contains options for TargetState.Add.
type AddOptions struct {
	Empty    bool
	Encrypt  bool
	Exact    bool
	Follow   bool
	Template bool
}

// An ImportTAROptions contains options for TargetState.ImportTAR.
type ImportTAROptions struct {
	DestinationDir  string
	Exact           bool
	StripComponents int
}

// A PopulateOptions contains options for TargetState.Populate.
type PopulateOptions struct {
	ExecuteTemplates bool
}

// A TargetState represents the root target state.
type TargetState struct {
	DestDir       string
	TargetIgnore  *PatternSet
	TargetRemove  *PatternSet
	Umask         os.FileMode
	SourceDir     string
	Data          map[string]interface{}
	TemplateFuncs template.FuncMap
	Templates     map[string]*template.Template
	GPG           *GPG
	Entries       map[string]Entry
	MinVersion    *semver.Version
}

// NewTargetState creates a new TargetState.
func NewTargetState(destDir string, umask os.FileMode, sourceDir string, data map[string]interface{}, templateFuncs template.FuncMap, gpg *GPG) *TargetState {
	return &TargetState{
		DestDir:       destDir,
		TargetIgnore:  NewPatternSet(),
		TargetRemove:  NewPatternSet(),
		Umask:         umask,
		SourceDir:     sourceDir,
		Data:          data,
		TemplateFuncs: templateFuncs,
		GPG:           gpg,
		Entries:       make(map[string]Entry),
	}
}

// Add adds a new target to ts.
func (ts *TargetState) Add(fs vfs.FS, addOptions AddOptions, targetPath string, info os.FileInfo, mutator Mutator) error {
	contains, err := vfs.Contains(fs, targetPath, ts.DestDir)
	if err != nil {
		return err
	}
	if !contains {
		return fmt.Errorf("%s: outside target directory", targetPath)
	}
	targetName, err := filepath.Rel(ts.DestDir, targetPath)
	if err != nil {
		return err
	}
	if info == nil {
		var err error
		if addOptions.Follow {
			info, err = fs.Stat(targetPath)
		} else {
			info, err = fs.Lstat(targetPath)
		}
		if err != nil {
			return err
		}
	} else if addOptions.Follow && info.Mode()&os.ModeType == os.ModeSymlink {
		info, err = fs.Stat(targetPath)
		if err != nil {
			return err
		}
	}

	// Add the parent directories, if needed.
	parentDirSourceName := ""
	entries := ts.Entries
	if parentDirName := filepath.Dir(targetName); parentDirName != "." {
		parentEntry, err := ts.findEntry(parentDirName)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if parentEntry == nil {
			if err := ts.Add(fs, addOptions, filepath.Join(ts.DestDir, parentDirName), nil, mutator); err != nil {
				return err
			}
			parentEntry, err = ts.findEntry(parentDirName)
			if err != nil {
				return err
			}
		} else if _, ok := parentEntry.(*Dir); !ok {
			return fmt.Errorf("%s: not a directory", parentDirName)
		}
		parentDir := parentEntry.(*Dir)
		parentDirSourceName = parentDir.sourceName
		entries = parentDir.Entries
	}

	switch {
	case info.IsDir():
		perm := info.Mode().Perm()
		infos, err := fs.ReadDir(targetPath)
		if err != nil {
			return err
		}
		empty := len(infos) == 0
		return ts.addDir(targetName, entries, parentDirSourceName, addOptions.Exact, perm, empty, mutator)
	case info.Mode().IsRegular():
		if info.Size() == 0 && !addOptions.Empty {
			return nil
		}
		contents, err := fs.ReadFile(targetPath)
		if err != nil {
			return err
		}
		if addOptions.Template {
			contents = autoTemplate(contents, ts.Data)
		}
		if addOptions.Encrypt {
			contents, err = ts.GPG.Encrypt(targetPath, contents)
			if err != nil {
				return err
			}
		}
		return ts.addFile(targetName, entries, parentDirSourceName, info, addOptions.Encrypt, addOptions.Template, contents, mutator, fs)
	case info.Mode()&os.ModeType == os.ModeSymlink:
		linkname, err := fs.Readlink(targetPath)
		if err != nil {
			return err
		}
		return ts.addSymlink(targetName, entries, parentDirSourceName, linkname, mutator)
	default:
		return fmt.Errorf("%s: not a regular file, directory, or symlink", targetName)
	}
}

// Apply ensures that ts.DestDir in fs matches ts.
func (ts *TargetState) Apply(fs vfs.FS, mutator Mutator, applyOptions *ApplyOptions) error {
	if applyOptions.Remove {
		// Build a set of targets to remove.
		targetsToRemove := make(map[string]struct{})
		includes := make([]string, 0, len(ts.TargetRemove.includes))
		for include := range ts.TargetRemove.includes {
			includes = append(includes, include)
		}
		for _, include := range includes {
			matches, err := fs.Glob(filepath.Join(ts.DestDir, include))
			if err != nil {
				return err
			}
			for _, match := range matches {
				relPath := strings.TrimPrefix(match, ts.DestDir+string(filepath.Separator))
				// Don't remove targets that are ignored.
				if ts.TargetIgnore.Match(relPath) {
					continue
				}
				// Don't remove targets that are excluded from remove.
				if !ts.TargetRemove.Match(relPath) {
					continue
				}
				targetsToRemove[match] = struct{}{}
			}
		}

		// FIXME check that the set of targets to remove does not intersect wth
		// the list of all entries.

		// Remove targets in reverse order so we remove children before their
		// parents.
		sortedTargetsToRemove := make([]string, 0, len(targetsToRemove))
		for target := range targetsToRemove {
			sortedTargetsToRemove = append(sortedTargetsToRemove, target)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(sortedTargetsToRemove)))
		for _, target := range sortedTargetsToRemove {
			if err := mutator.RemoveAll(target); err != nil {
				return err
			}
		}
	}

	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].Apply(fs, mutator, applyOptions); err != nil {
			return err
		}
	}
	return nil
}

// Archive writes ts to w.
func (ts *TargetState) Archive(w *tar.Writer, umask os.FileMode) error {
	headerTemplate, err := ts.getTarHeaderTemplate()
	if err != nil {
		return err
	}

	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].archive(w, ts.TargetIgnore.Match, headerTemplate, umask); err != nil {
			return err
		}
	}
	return nil
}

// ConcreteValue returns a value suitable for serialization.
func (ts *TargetState) ConcreteValue(recursive bool) (interface{}, error) {
	var entryConcreteValues []interface{}
	for _, entryName := range sortedEntryNames(ts.Entries) {
		entryConcreteValue, err := ts.Entries[entryName].ConcreteValue(ts.DestDir, ts.TargetIgnore.Match, ts.SourceDir, ts.Umask, recursive)
		if err != nil {
			return nil, err
		}
		if entryConcreteValue != nil {
			entryConcreteValues = append(entryConcreteValues, entryConcreteValue)
		}
	}
	return entryConcreteValues, nil
}

// Evaluate evaluates all of the entries in ts.
func (ts *TargetState) Evaluate() error {
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].Evaluate(ts.TargetIgnore.Match); err != nil {
			return err
		}
	}
	return nil
}

// Get returns the state of the given target, or nil if no such target is found.
func (ts *TargetState) Get(fs vfs.Stater, target string) (Entry, error) {
	contains, err := vfs.Contains(fs, target, ts.DestDir)
	if err != nil {
		return nil, err
	}
	if !contains {
		return nil, fmt.Errorf("%s: outside target directory", target)
	}
	targetName, err := filepath.Rel(ts.DestDir, target)
	if err != nil {
		return nil, err
	}
	return ts.findEntry(targetName)
}

// ImportTAR imports a tar archive.
func (ts *TargetState) ImportTAR(r *tar.Reader, importTAROptions ImportTAROptions, mutator Mutator, fs PrivacyStater) error {
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeSymlink:
			if err := ts.importHeader(r, importTAROptions, header, mutator, fs); err != nil {
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
func (ts *TargetState) Populate(fs vfs.FS, options *PopulateOptions) error {
	return vfs.Walk(fs, ts.SourceDir, func(path string, info os.FileInfo, _ error) error {
		relPath, err := filepath.Rel(ts.SourceDir, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		// Treat all files and directories beginning with "." specially.
		if _, name := filepath.Split(relPath); strings.HasPrefix(name, ".") {
			switch {
			case info.Name() == ignoreName:
				dns := dirNames(parseDirNameComponents(splitPathList(relPath)))
				return ts.addPatterns(fs, ts.TargetIgnore, path, filepath.Join(dns...))
			case info.Name() == removeName:
				dns := dirNames(parseDirNameComponents(splitPathList(relPath)))
				return ts.addPatterns(fs, ts.TargetRemove, path, filepath.Join(dns...))
			case info.Name() == templatesDirName:
				if err := ts.addTemplatesDir(fs, path); err != nil {
					return err
				}
				return filepath.SkipDir
			case info.Name() == versionName:
				data, err := fs.ReadFile(path)
				if err != nil {
					return err
				}
				version, err := semver.NewVersion(strings.TrimSpace(string(data)))
				if err != nil {
					return err
				}
				if ts.MinVersion == nil || ts.MinVersion.LessThan(*version) {
					ts.MinVersion = version
				}
				return nil
			case info.IsDir():
				// Don't recurse into ignored subdirectories.
				return filepath.SkipDir
			}
			// Ignore all other files and directories.
			return nil
		}
		switch {
		case info.IsDir():
			components := splitPathList(relPath)
			das := parseDirNameComponents(components)
			dns := dirNames(das)
			targetName := filepath.Join(dns...)
			entries, err := ts.findEntries(dns[:len(dns)-1])
			if err != nil {
				return err
			}
			da := das[len(das)-1]
			entries[da.Name] = newDir(relPath, targetName, da.Exact, da.Perm)
		case info.Mode().IsRegular():
			psfp := parseSourceFilePath(relPath)
			dns := dirNames(psfp.dirAttributes)
			entries, err := ts.findEntries(dns)
			if err != nil {
				return err
			}
			switch {
			case psfp.fileAttributes != nil && psfp.fileAttributes.Mode&os.ModeType == 0 || psfp.scriptAttributes != nil:
				readFile := func() ([]byte, error) {
					return fs.ReadFile(path)
				}
				evaluateContents := readFile
				if psfp.fileAttributes != nil && psfp.fileAttributes.Encrypted {
					prevEvaluateContents := evaluateContents
					evaluateContents = func() ([]byte, error) {
						ciphertext, err := prevEvaluateContents()
						if err != nil {
							return nil, err
						}
						return ts.GPG.Decrypt(path, ciphertext)
					}
				}
				if psfp.fileAttributes != nil && psfp.fileAttributes.Template || psfp.scriptAttributes != nil && psfp.scriptAttributes.Template {
					if options == nil || options.ExecuteTemplates {
						prevEvaluateContents := evaluateContents
						evaluateContents = func() ([]byte, error) {
							data, err := prevEvaluateContents()
							if err != nil {
								return nil, err
							}
							return ts.executeTemplateData(path, data)
						}
					}
				}
				switch {
				case psfp.fileAttributes != nil:
					entry := &File{
						sourceName:       relPath,
						targetName:       filepath.Join(append(dns, psfp.fileAttributes.Name)...),
						Empty:            psfp.fileAttributes.Empty,
						Encrypted:        psfp.fileAttributes.Encrypted,
						Perm:             psfp.fileAttributes.Mode.Perm(),
						Template:         psfp.fileAttributes.Template,
						evaluateContents: evaluateContents,
					}
					entries[psfp.fileAttributes.Name] = entry
				case psfp.scriptAttributes != nil:
					entry := &Script{
						sourceName:       relPath,
						targetName:       filepath.Join(append(dns, psfp.scriptAttributes.Name)...),
						Once:             psfp.scriptAttributes.Once,
						Template:         psfp.scriptAttributes.Template,
						evaluateContents: evaluateContents,
					}
					entries[psfp.scriptAttributes.Name] = entry
				}
			case psfp.fileAttributes != nil && psfp.fileAttributes.Mode&os.ModeType == os.ModeSymlink:
				evaluateLinkname := func() (string, error) {
					data, err := fs.ReadFile(path)
					return string(data), err
				}
				if psfp.fileAttributes.Template {
					evaluateLinkname = func() (string, error) {
						data, err := ts.executeTemplate(fs, path)
						return string(data), err
					}
				}
				entry := &Symlink{
					sourceName:       relPath,
					targetName:       filepath.Join(append(dns, psfp.fileAttributes.Name)...),
					Template:         psfp.fileAttributes.Template,
					evaluateLinkname: evaluateLinkname,
				}
				entries[psfp.fileAttributes.Name] = entry
			default:
				return fmt.Errorf("%s: unsupported file type", path)
			}
		default:
			return fmt.Errorf("%s: unsupported file type", path)
		}
		return nil
	})
}

func (ts *TargetState) addDir(targetName string, entries map[string]Entry, parentDirSourceName string, exact bool, perm os.FileMode, empty bool, mutator Mutator) error {
	name := filepath.Base(targetName)
	if entry, ok := entries[name]; ok {
		if _, ok = entry.(*Dir); !ok {
			return fmt.Errorf("%s: already added and not a directory", targetName)
		}
		return nil
	}
	sourceName := DirAttributes{
		Name:  name,
		Exact: exact,
		Perm:  perm,
	}.SourceName()
	if parentDirSourceName != "" {
		sourceName = filepath.Join(parentDirSourceName, sourceName)
	}
	dir := newDir(sourceName, targetName, exact, perm)
	if err := mutator.Mkdir(filepath.Join(ts.SourceDir, sourceName), 0777&^ts.Umask); err != nil {
		return err
	}
	// If the directory is empty, add a .keep file so the directory is
	// managed by git. Chezmoi will ignore the .keep file as it begins with
	// a dot.
	if empty {
		if err := mutator.WriteFile(filepath.Join(ts.SourceDir, sourceName, ".keep"), nil, 0666&^ts.Umask, nil); err != nil {
			return err
		}
	}
	entries[name] = dir
	return nil
}

func (ts *TargetState) addFile(targetName string, entries map[string]Entry, parentDirSourceName string, info os.FileInfo, encrypted, template bool, contents []byte, mutator Mutator, fs PrivacyStater) error {
	name := filepath.Base(targetName)
	var existingFile *File
	var existingContents []byte
	if entry, ok := entries[name]; ok {
		existingFile, ok = entry.(*File)
		if !ok {
			return fmt.Errorf("%s: already added and not a regular file", targetName)
		}
		var err error
		existingContents, err = existingFile.Contents()
		if err != nil {
			return err
		}
	}

	perm := info.Mode().Perm()
	destFile := filepath.Join(ts.DestDir, name)
	if IsPrivate(fs, destFile, ts.Umask) {
		// since Windows doesn't really have the concept of "groups", the
		// group permission bits might be set even on a file that should
		// be considered private.  This will clear them.  Posix-style platforms
		// remain unaffected because IsPrivate will only return true if those
		// bits weren't set in the first place
		perm &^= 0077
	}

	empty := info.Size() == 0
	sourceName := FileAttributes{
		Name:      name,
		Mode:      perm,
		Empty:     empty,
		Encrypted: encrypted,
		Template:  template,
	}.SourceName()
	if parentDirSourceName != "" {
		sourceName = filepath.Join(parentDirSourceName, sourceName)
	}
	file := &File{
		sourceName: sourceName,
		targetName: targetName,
		Empty:      empty,
		Encrypted:  encrypted,
		Perm:       perm,
		Template:   template,
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
}

func (ts *TargetState) addPatterns(fs vfs.FS, ps *PatternSet, path, relPath string) error {
	data, err := ts.executeTemplate(fs, path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(relPath)
	s := bufio.NewScanner(bytes.NewReader(data))
	for s.Scan() {
		text := s.Text()
		if index := strings.IndexRune(text, '#'); index != -1 {
			text = text[:index]
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		include := true
		if strings.HasPrefix(text, "!") {
			include = false
			text = strings.TrimPrefix(text, "!")
		}
		pattern := filepath.Join(dir, text)
		if err := ps.Add(pattern, include); err != nil {
			return fmt.Errorf("%s: %v", path, err)
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("%s: %v", path, err)
	}
	return nil
}

func (ts *TargetState) addSymlink(targetName string, entries map[string]Entry, parentDirSourceName string, linkname string, mutator Mutator) error {
	name := filepath.Base(targetName)
	var existingSymlink *Symlink
	var existingLinkname string
	if entry, ok := entries[name]; ok {
		existingSymlink, ok = entry.(*Symlink)
		if !ok {
			return fmt.Errorf("%s: already added and not a symlink", targetName)
		}
		var err error
		existingLinkname, err = existingSymlink.Linkname()
		if err != nil {
			return err
		}
	}
	sourceName := FileAttributes{
		Name: name,
		Mode: os.ModeSymlink,
	}.SourceName()
	if parentDirSourceName != "" {
		sourceName = filepath.Join(parentDirSourceName, sourceName)
	}
	symlink := &Symlink{
		sourceName: sourceName,
		targetName: targetName,
		linkname:   linkname,
	}
	if existingSymlink != nil {
		if existingSymlink.linkname == symlink.linkname {
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
	return mutator.WriteFile(filepath.Join(ts.SourceDir, symlink.sourceName), []byte(symlink.linkname), 0666&^ts.Umask, []byte(existingLinkname))
}

func (ts *TargetState) addTemplatesDir(fs vfs.FS, path string) error {
	return vfs.Walk(fs, path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			contents, err := fs.ReadFile(path)
			if err != nil {
				return err
			}
			name := filepath.Base(path)
			tmpl, err := template.New(name).Parse(string(contents))
			if err != nil {
				return err
			}
			if ts.Templates == nil {
				ts.Templates = make(map[string]*template.Template)
			}
			ts.Templates[name] = tmpl
			return nil
		case info.IsDir():
			return nil
		default:
			return fmt.Errorf("unsupported file in %s: %s", templatesDirName, path)
		}
	})
}

func (ts *TargetState) executeTemplate(fs vfs.FS, path string) ([]byte, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ts.executeTemplateData(path, data)
}

func (ts *TargetState) executeTemplateData(name string, data []byte) ([]byte, error) {
	tmpl, err := template.New(name).Option("missingkey=error").Funcs(ts.TemplateFuncs).Parse(string(data))
	if err != nil {
		return nil, err
	}
	for name, t := range ts.Templates {
		tmpl, err = tmpl.AddParseTree(name, t.Tree)
		if err != nil {
			return nil, err
		}
	}
	output := &bytes.Buffer{}
	if err = tmpl.Execute(output, ts.Data); err != nil {
		return nil, err
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

func (ts *TargetState) importHeader(r io.Reader, importTAROptions ImportTAROptions, header *tar.Header, mutator Mutator, fs PrivacyStater) error {
	targetPath := header.Name
	if importTAROptions.StripComponents > 0 {
		targetPath = filepath.Join(strings.Split(targetPath, string(os.PathSeparator))[importTAROptions.StripComponents:]...)
	}
	if importTAROptions.DestinationDir != "" {
		targetPath = filepath.Join(importTAROptions.DestinationDir, targetPath)
	} else {
		targetPath = filepath.Join(ts.DestDir, targetPath)
	}
	targetName, err := filepath.Rel(ts.DestDir, targetPath)
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
	switch header.Typeflag {
	case tar.TypeDir:
		perm := os.FileMode(header.Mode).Perm()
		empty := false // FIXME don't assume directory is empty
		return ts.addDir(targetName, entries, parentDirSourceName, importTAROptions.Exact, perm, empty, mutator)
	case tar.TypeReg:
		info := header.FileInfo()
		contents, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		return ts.addFile(targetName, entries, parentDirSourceName, info, false, false, contents, mutator, fs)
	case tar.TypeSymlink:
		linkname := header.Linkname
		return ts.addSymlink(targetName, entries, parentDirSourceName, linkname, mutator)
	default:
		return fmt.Errorf("%s: unspported typeflag '%c'", header.Name, header.Typeflag)
	}
}
