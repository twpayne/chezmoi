package chezmoi

import (
	"archive/tar"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs"
)

const (
	ignoreName  = ".chezmoiignore"
	versionName = ".chezmoiversion"
)

// An AddOptions contains options for TargetState.Add.
type AddOptions struct {
	Empty    bool
	Encrypt  bool
	Exact    bool
	Template bool
}

// An ImportTAROptions contains options for TargetState.ImportTAR.
type ImportTAROptions struct {
	DestinationDir  string
	Exact           bool
	StripComponents int
}

// A TargetState represents the root target state.
type TargetState struct {
	DestDir       string
	TargetIgnore  *PatternSet
	Umask         os.FileMode
	SourceDir     string
	Data          map[string]interface{}
	TemplateFuncs template.FuncMap
	GPGRecipient  string
	Entries       map[string]Entry
	MinVersion    *semver.Version
	Scripts       map[string]*Script
}

// NewTargetState creates a new TargetState.
func NewTargetState(destDir string, umask os.FileMode, sourceDir string, data map[string]interface{}, templateFuncs template.FuncMap, gpgRecipient string) *TargetState {
	return &TargetState{
		DestDir:       destDir,
		TargetIgnore:  NewPatternSet(),
		Umask:         umask,
		SourceDir:     sourceDir,
		Data:          data,
		TemplateFuncs: templateFuncs,
		GPGRecipient:  gpgRecipient,
		Entries:       make(map[string]Entry),
		Scripts:       make(map[string]*Script),
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
		info, err = fs.Lstat(targetPath)
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
			contents, err = autoTemplate(contents, ts.Data)
			if err != nil {
				return err
			}
		}
		if addOptions.Encrypt {
			contents, err = ts.Encrypt(contents)
			if err != nil {
				return err
			}
		}
		return ts.addFile(targetName, entries, parentDirSourceName, info, addOptions.Encrypt, addOptions.Template, contents, mutator)
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
func (ts *TargetState) Apply(fs vfs.FS, mutator Mutator) error {
	for _, entryName := range sortedEntryNames(ts.Entries) {
		if err := ts.Entries[entryName].Apply(fs, ts.DestDir, ts.TargetIgnore.Match, ts.Umask, mutator); err != nil {
			return err
		}
	}

	return nil
}

// ApplyScripts that scripts get executed as required
func (ts *TargetState) ApplyScripts(fs vfs.FS, force bool, dryRun bool) error {
	for _, scriptName := range sortedScriptNames(ts.Scripts) {
		if err := ts.Scripts[scriptName].Apply(ts.DestDir, dryRun); err != nil {
			return err
		}
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
		if err := ts.Entries[entryName].archive(w, ts.TargetIgnore.Match, &headerTemplate, umask); err != nil {
			return err
		}
	}
	for _, scriptName := range sortedScriptNames(ts.Scripts) {
		if err := ts.Scripts[scriptName].Evaluate(); err != nil {
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

// Decrypt decrypts ciphertext.
func (ts *TargetState) Decrypt(ciphertext []byte) ([]byte, error) {
	cmd := exec.Command("gpg", "--decrypt")
	cmd.Stdin = bytes.NewReader(ciphertext)
	return cmd.Output()
}

// Encrypt encrypts plaintext for ts's recipient.
func (ts *TargetState) Encrypt(plaintext []byte) ([]byte, error) {
	args := []string{"--armor", "--encrypt"}
	if ts.GPGRecipient != "" {
		args = append(args, "--recipient", ts.GPGRecipient)
	}
	cmd := exec.Command("gpg", args...)
	cmd.Stdin = bytes.NewReader(plaintext)
	return cmd.Output()
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
func (ts *TargetState) Get(fs vfs.FS, target string) (Entry, error) {
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
func (ts *TargetState) ImportTAR(r *tar.Reader, importTAROptions ImportTAROptions, mutator Mutator) error {
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir, tar.TypeReg, tar.TypeSymlink:
			if err := ts.importHeader(r, importTAROptions, header, mutator); err != nil {
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
	if err := ts.populateScripts(fs); err != nil {
		return err
	}
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
				return ts.addSourceIgnore(fs, path, filepath.Join(dns...))
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

			targetName := filepath.Join(append(dns, psfp.Name)...)
			var entry Entry
			switch psfp.Mode & os.ModeType {
			case 0:
				readFile := func() ([]byte, error) {
					return fs.ReadFile(path)
				}
				evaluateContents := readFile
				if psfp.Encrypted {
					prevEvaluateContents := evaluateContents
					evaluateContents = func() ([]byte, error) {
						ciphertext, err := prevEvaluateContents()
						if err != nil {
							return nil, err
						}
						return ts.Decrypt(ciphertext)
					}
				}
				if psfp.Template {
					prevEvaluateContents := evaluateContents
					evaluateContents = func() ([]byte, error) {
						data, err := prevEvaluateContents()
						if err != nil {
							return nil, err
						}
						return ts.executeTemplateData(path, data)
					}
				}
				entry = &File{
					sourceName:       relPath,
					targetName:       targetName,
					Empty:            psfp.Empty,
					Encrypted:        psfp.Encrypted,
					Perm:             psfp.Mode.Perm(),
					Template:         psfp.Template,
					evaluateContents: evaluateContents,
				}
			case os.ModeSymlink:
				evaluateLinkname := func() (string, error) {
					data, err := fs.ReadFile(path)
					return string(data), err
				}
				if psfp.Template {
					evaluateLinkname = func() (string, error) {
						data, err := ts.executeTemplate(fs, path)
						return string(data), err
					}
				}
				entry = &Symlink{
					sourceName:       relPath,
					targetName:       targetName,
					Template:         psfp.Template,
					evaluateLinkname: evaluateLinkname,
				}
			default:
				return fmt.Errorf("%v: unsupported mode 0%o", path, psfp.Mode&os.ModeType)
			}
			entries[psfp.Name] = entry
		default:
			return fmt.Errorf("%s: unsupported file type", path)
		}
		return nil
	})
}

func (ts *TargetState) populateScripts(fs vfs.FS) error {
	scriptsDir := path.Join(ts.SourceDir, scriptsDir)
	files, err := fs.ReadDir(scriptsDir)
	if os.IsNotExist(err) {
		return nil
	}
	for _, f := range files {
		if !f.IsDir() && f.Mode()&0111 != 0 {
			fp := path.Join(scriptsDir, f.Name())
			psfp := parseSourceFilePath(fp)
			evaluateContents := func() ([]byte, error) {
				return fs.ReadFile(fp)
			}
			if psfp.Template {
				evaluateContents = func() ([]byte, error) {
					return ts.executeTemplate(fs, fp)
				}
			}
			s := &Script{
				Name:             StripTemplateExtension(f.Name()),
				SourcePath:       fp,
				evaluateContents: evaluateContents,
			}
			ts.Scripts[s.Name] = s
		}
	}
	return nil
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

func (ts *TargetState) addFile(targetName string, entries map[string]Entry, parentDirSourceName string, info os.FileInfo, encrypted, template bool, contents []byte, mutator Mutator) error {
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

func (ts *TargetState) addSourceIgnore(fs vfs.FS, path, relPath string) error {
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
		if err := ts.TargetIgnore.Add(pattern, include); err != nil {
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

func (ts *TargetState) executeTemplate(fs vfs.FS, path string) ([]byte, error) {
	data, err := fs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ts.executeTemplateData(path, data)
}

func (ts *TargetState) executeTemplateData(name string, data []byte) (_ []byte, err error) {
	tmpl, err := template.New(name).Option("missingkey=error").Funcs(ts.TemplateFuncs).Parse(string(data))
	if err != nil {
		return nil, err
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

func (ts *TargetState) importHeader(r io.Reader, importTAROptions ImportTAROptions, header *tar.Header, mutator Mutator) error {
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
		return ts.addFile(targetName, entries, parentDirSourceName, info, false, false, contents, mutator)
	case tar.TypeSymlink:
		linkname := header.Linkname
		return ts.addSymlink(targetName, entries, parentDirSourceName, linkname, mutator)
	default:
		return fmt.Errorf("%s: unspported typeflag '%c'", header.Name, header.Typeflag)
	}
}
