package chezmoi

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/coreos/go-semver/semver"
	vfs "github.com/twpayne/go-vfs/v2"
	"go.uber.org/multierr"
)

// A SourceState is a source state.
type SourceState struct {
	entries                 map[RelPath]SourceStateEntry
	system                  System
	sourceDirAbsPath        AbsPath
	destDirAbsPath          AbsPath
	umask                   os.FileMode
	encryption              Encryption
	ignore                  *patternSet
	minVersion              semver.Version
	defaultTemplateDataFunc func() map[string]interface{}
	userTemplateData        map[string]interface{}
	priorityTemplateData    map[string]interface{}
	templateData            map[string]interface{}
	templateFuncs           template.FuncMap
	templateOptions         []string
	templates               map[string]*template.Template
}

// A SourceStateOption sets an option on a source state.
type SourceStateOption func(*SourceState)

// WithDestDir sets the destination directory.
func WithDestDir(destDirAbsPath AbsPath) SourceStateOption {
	return func(s *SourceState) {
		s.destDirAbsPath = destDirAbsPath
	}
}

// WithEncryption sets the encryption.
func WithEncryption(encryption Encryption) SourceStateOption {
	return func(s *SourceState) {
		s.encryption = encryption
	}
}

// WithPriorityTemplateData adds priority template data.
func WithPriorityTemplateData(priorityTemplateData map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		RecursiveMerge(s.priorityTemplateData, priorityTemplateData)
	}
}

// WithSourceDir sets the source directory.
func WithSourceDir(sourceDirAbsPath AbsPath) SourceStateOption {
	return func(s *SourceState) {
		s.sourceDirAbsPath = sourceDirAbsPath
	}
}

// WithSystem sets the system.
func WithSystem(system System) SourceStateOption {
	return func(s *SourceState) {
		s.system = system
	}
}

// WithDefaultTemplateDataFunc sets the default template data function.
func WithDefaultTemplateDataFunc(defaultTemplateDataFunc func() map[string]interface{}) SourceStateOption {
	return func(s *SourceState) {
		s.defaultTemplateDataFunc = defaultTemplateDataFunc
	}
}

// WithTemplateFuncs sets the template functions.
func WithTemplateFuncs(templateFuncs template.FuncMap) SourceStateOption {
	return func(s *SourceState) {
		s.templateFuncs = templateFuncs
	}
}

// WithTemplateOptions sets the template options.
func WithTemplateOptions(templateOptions []string) SourceStateOption {
	return func(s *SourceState) {
		s.templateOptions = templateOptions
	}
}

type targetStateEntryFunc func(System, AbsPath) (TargetStateEntry, error)

// NewSourceState creates a new source state with the given options.
func NewSourceState(options ...SourceStateOption) *SourceState {
	s := &SourceState{
		entries:              make(map[RelPath]SourceStateEntry),
		umask:                Umask,
		encryption:           NoEncryption{},
		ignore:               newPatternSet(),
		priorityTemplateData: make(map[string]interface{}),
		userTemplateData:     make(map[string]interface{}),
		templateOptions:      DefaultTemplateOptions,
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// AddOptions are options to SourceState.Add.
type AddOptions struct {
	AutoTemplate     bool
	Create           bool
	Empty            bool
	Encrypt          bool
	EncryptedSuffix  string
	Exact            bool
	Include          *EntryTypeSet
	RemoveDir        RelPath
	Template         bool
	TemplateSymlinks bool
}

// Add adds destAbsPathInfos to s.
func (s *SourceState) Add(sourceSystem System, persistentState PersistentState, destSystem System, destAbsPathInfos map[AbsPath]os.FileInfo, options *AddOptions) error {
	type update struct {
		destAbsPath    AbsPath
		entryState     *EntryState
		sourceRelPaths SourceRelPaths
	}

	destAbsPaths := make(AbsPaths, 0, len(destAbsPathInfos))
	for destAbsPath := range destAbsPathInfos {
		destAbsPaths = append(destAbsPaths, destAbsPath)
	}
	sort.Sort(destAbsPaths)

	updates := make([]update, 0, len(destAbsPathInfos))
	newSourceStateEntries := make(map[SourceRelPath]SourceStateEntry)
	newSourceStateEntriesByTargetRelPath := make(map[RelPath]SourceStateEntry)
	for _, destAbsPath := range destAbsPaths {
		destAbsPathInfo := destAbsPathInfos[destAbsPath]
		if !options.Include.IncludeFileInfo(destAbsPathInfo) {
			continue
		}
		targetRelPath := destAbsPath.MustTrimDirPrefix(s.destDirAbsPath)

		// Find the target's parent directory in the source state.
		var parentSourceRelPath SourceRelPath
		if targetParentRelPath := targetRelPath.Dir(); targetParentRelPath == "." {
			parentSourceRelPath = SourceRelPath{}
		} else if parentEntry, ok := newSourceStateEntriesByTargetRelPath[targetParentRelPath]; ok {
			parentSourceRelPath = parentEntry.SourceRelPath()
		} else if parentEntry, ok := s.entries[targetParentRelPath]; ok {
			parentSourceRelPath = parentEntry.SourceRelPath()
		} else {
			return fmt.Errorf("%s: parent directory not in source state", destAbsPath)
		}

		actualStateEntry, err := NewActualStateEntry(destSystem, destAbsPath, destAbsPathInfo, nil)
		if err != nil {
			return err
		}
		newSourceStateEntry, err := s.sourceStateEntry(actualStateEntry, destAbsPath, destAbsPathInfo, parentSourceRelPath, options)
		if err != nil {
			return err
		}
		if newSourceStateEntry == nil {
			continue
		}

		sourceEntryRelPath := newSourceStateEntry.SourceRelPath()

		entryState, err := actualStateEntry.EntryState()
		if err != nil {
			return err
		}
		update := update{
			destAbsPath:    destAbsPath,
			entryState:     entryState,
			sourceRelPaths: SourceRelPaths{sourceEntryRelPath},
		}

		if oldSourceStateEntry, ok := s.entries[targetRelPath]; ok {
			// If both the new and old source state entries are directories but the name has changed,
			// rename to avoid losing the directory's contents. Otherwise,
			// remove the old.
			oldSourceEntryRelPath := oldSourceStateEntry.SourceRelPath()
			if !oldSourceEntryRelPath.Empty() && oldSourceEntryRelPath != sourceEntryRelPath {
				_, newIsDir := newSourceStateEntry.(*SourceStateDir)
				_, oldIsDir := oldSourceStateEntry.(*SourceStateDir)
				if newIsDir && oldIsDir {
					newSourceStateEntry = &SourceStateRenameDir{
						oldSourceRelPath: oldSourceEntryRelPath,
						newSourceRelPath: sourceEntryRelPath,
					}
				} else {
					newSourceStateEntries[oldSourceEntryRelPath] = &SourceStateRemove{}
					update.sourceRelPaths = append(update.sourceRelPaths, oldSourceEntryRelPath)
				}
			}
		}

		newSourceStateEntries[sourceEntryRelPath] = newSourceStateEntry
		newSourceStateEntriesByTargetRelPath[targetRelPath] = newSourceStateEntry

		updates = append(updates, update)
	}

	entries := make(map[RelPath]SourceStateEntry)
	for sourceRelPath, sourceStateEntry := range newSourceStateEntries {
		entries[sourceRelPath.RelPath()] = sourceStateEntry
	}

	// Simulate removing a directory by creating SourceStateRemove entries for
	// all existing source state entries that are in options.RemoveDir and not
	// in the new source state.
	if options.RemoveDir != "" {
		for targetRelPath, sourceStateEntry := range s.entries {
			if !targetRelPath.HasDirPrefix(options.RemoveDir) {
				continue
			}
			if _, ok := newSourceStateEntriesByTargetRelPath[targetRelPath]; ok {
				continue
			}
			sourceRelPath := sourceStateEntry.SourceRelPath()
			entries[sourceRelPath.RelPath()] = &SourceStateRemove{}
			update := update{
				destAbsPath: s.destDirAbsPath.Join(targetRelPath),
				entryState: &EntryState{
					Type: EntryStateTypeRemove,
				},
				sourceRelPaths: []SourceRelPath{sourceRelPath},
			}
			updates = append(updates, update)
		}
	}

	targetSourceState := &SourceState{
		entries: entries,
	}

	for _, update := range updates {
		for _, sourceRelPath := range update.sourceRelPaths {
			if err := targetSourceState.Apply(sourceSystem, sourceSystem, NullPersistentState{}, s.sourceDirAbsPath, sourceRelPath.RelPath(), ApplyOptions{
				Include: options.Include,
				Umask:   s.umask,
			}); err != nil {
				return err
			}
		}
		if err := persistentStateSet(persistentState, entryStateBucket, []byte(update.destAbsPath), update.entryState); err != nil {
			return err
		}
	}

	return nil
}

// AddDestAbsPathInfos adds an os.FileInfo to destAbsPathInfos for destAbsPath
// and any of its parents which are not already known.
func (s *SourceState) AddDestAbsPathInfos(destAbsPathInfos map[AbsPath]os.FileInfo, system System, destAbsPath AbsPath, info os.FileInfo) error {
	for {
		if _, err := destAbsPath.TrimDirPrefix(s.destDirAbsPath); err != nil {
			return err
		}

		if _, ok := destAbsPathInfos[destAbsPath]; ok {
			return nil
		}

		if info == nil {
			var err error
			info, err = system.Lstat(destAbsPath)
			if err != nil {
				return err
			}
		}
		destAbsPathInfos[destAbsPath] = info

		parentAbsPath := destAbsPath.Dir()
		if parentAbsPath == s.destDirAbsPath {
			return nil
		}
		parentRelPath := parentAbsPath.MustTrimDirPrefix(s.destDirAbsPath)
		if _, ok := s.entries[parentRelPath]; ok {
			return nil
		}

		destAbsPath = parentAbsPath
		info = nil
	}
}

// A PreApplyFunc is called before a target is applied.
type PreApplyFunc func(targetRelPath RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *EntryState) error

// ApplyOptions are options to SourceState.ApplyAll and SourceState.ApplyOne.
type ApplyOptions struct {
	Include      *EntryTypeSet
	PreApplyFunc PreApplyFunc
	Umask        os.FileMode
}

// Apply updates targetRelPath in targetDir in destSystem to match s.
func (s *SourceState) Apply(targetSystem, destSystem System, persistentState PersistentState, targetDir AbsPath, targetRelPath RelPath, options ApplyOptions) error {
	sourceStateEntry := s.entries[targetRelPath]

	if !options.Include.IncludeEncrypted() {
		if sourceStateFile, ok := sourceStateEntry.(*SourceStateFile); ok && sourceStateFile.Attr.Encrypted {
			return nil
		}
	}

	destAbsPath := s.destDirAbsPath.Join(targetRelPath)
	targetStateEntry, err := sourceStateEntry.TargetStateEntry(destSystem, destAbsPath)
	if err != nil {
		return err
	}

	if options.Include != nil && !options.Include.IncludeTargetStateEntry(targetStateEntry) {
		return nil
	}

	targetAbsPath := targetDir.Join(targetRelPath)

	targetEntryState, err := targetStateEntry.EntryState(options.Umask)
	if err != nil {
		return err
	}

	switch skip, err := targetStateEntry.SkipApply(persistentState); {
	case err != nil:
		return err
	case skip:
		return nil
	}

	actualStateEntry, err := NewActualStateEntry(targetSystem, targetAbsPath, nil, nil)
	if err != nil {
		return err
	}

	if options.PreApplyFunc != nil {
		var lastWrittenEntryState *EntryState
		var entryState EntryState
		ok, err := persistentStateGet(persistentState, entryStateBucket, []byte(targetAbsPath), &entryState)
		if err != nil {
			return err
		}
		if ok {
			lastWrittenEntryState = &entryState
		}

		actualEntryState, err := actualStateEntry.EntryState()
		if err != nil {
			return err
		}

		if err := options.PreApplyFunc(targetRelPath, targetEntryState, lastWrittenEntryState, actualEntryState); err != nil {
			return err
		}
	}

	if changed, err := targetStateEntry.Apply(targetSystem, persistentState, actualStateEntry); err != nil {
		return err
	} else if !changed {
		return nil
	}

	return persistentStateSet(persistentState, entryStateBucket, []byte(targetAbsPath), targetEntryState)
}

// Encryption returns s's encryption.
func (s *SourceState) Encryption() Encryption {
	return s.encryption
}

// Entries returns s's source state entries.
func (s *SourceState) Entries() map[RelPath]SourceStateEntry {
	return s.entries
}

// Entry returns the source state entry for targetRelPath.
func (s *SourceState) Entry(targetRelPath RelPath) (SourceStateEntry, bool) {
	sourceStateEntry, ok := s.entries[targetRelPath]
	return sourceStateEntry, ok
}

// ExecuteTemplateData returns the result of executing template data.
func (s *SourceState) ExecuteTemplateData(name string, data []byte) ([]byte, error) {
	tmpl, err := template.New(name).
		Option(s.templateOptions...).
		Funcs(s.templateFuncs).
		Parse(string(data))
	if err != nil {
		return nil, err
	}
	for name, t := range s.templates {
		tmpl, err = tmpl.AddParseTree(name, t.Tree)
		if err != nil {
			return nil, err
		}
	}
	var sb strings.Builder
	if err = tmpl.ExecuteTemplate(&sb, name, s.TemplateData()); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

// Ignored returns if targetRelPath is ignored.
func (s *SourceState) Ignored(targetRelPath RelPath) bool {
	return s.ignore.match(string(targetRelPath))
}

// MinVersion returns the minimum version for which s is valid.
func (s *SourceState) MinVersion() semver.Version {
	return s.minVersion
}

// MustEntry returns the source state entry associated with targetRelPath, and
// panics if it does not exist.
func (s *SourceState) MustEntry(targetRelPath RelPath) SourceStateEntry {
	sourceStateEntry, ok := s.entries[targetRelPath]
	if !ok {
		panic(fmt.Sprintf("%s: not in source state", targetRelPath))
	}
	return sourceStateEntry
}

// Read reads the source state from the source directory.
func (s *SourceState) Read() error {
	info, err := s.system.Lstat(s.sourceDirAbsPath)
	switch {
	case os.IsNotExist(err):
		return nil
	case err != nil:
		return err
	case !info.IsDir():
		return fmt.Errorf("%s: not a directory", s.sourceDirAbsPath)
	}

	// Read all source entries.
	allSourceStateEntries := make(map[RelPath][]SourceStateEntry)
	if err := Walk(s.system, s.sourceDirAbsPath, func(sourceAbsPath AbsPath, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if sourceAbsPath == s.sourceDirAbsPath {
			return nil
		}
		sourceRelPath := SourceRelPath{
			relPath: sourceAbsPath.MustTrimDirPrefix(s.sourceDirAbsPath),
			isDir:   info.IsDir(),
		}

		parentSourceRelPath, sourceName := sourceRelPath.Split()
		// Follow symlinks in the source directory.
		if info.Mode()&os.ModeType == os.ModeSymlink {
			info, err = s.system.Stat(s.sourceDirAbsPath.Join(sourceRelPath.RelPath()))
			if err != nil {
				return err
			}
		}
		switch {
		case strings.HasPrefix(info.Name(), dataName):
			return s.addTemplateData(sourceAbsPath)
		case info.Name() == ignoreName:
			// .chezmoiignore is interpreted as a template. we walk the
			// filesystem in alphabetical order, so, luckily for us,
			// .chezmoidata will be read before .chezmoiignore, so data in
			// .chezmoidata is available to be used in .chezmoiignore. Unluckily
			// for us, .chezmoitemplates will be read afterwards so partial
			// templates will not be available in .chezmoiignore.
			return s.addPatterns(s.ignore, sourceAbsPath, parentSourceRelPath)
		case info.Name() == removeName:
			// The comment about .chezmoiignore and templates applies to
			// .chezmoiremove too.
			removePatterns := newPatternSet()
			if err := s.addPatterns(removePatterns, sourceAbsPath, sourceRelPath); err != nil {
				return err
			}
			matches, err := removePatterns.glob(s.system.UnderlyingFS(), string(s.destDirAbsPath)+"/")
			if err != nil {
				return err
			}
			n := 0
			for _, match := range matches {
				if !s.Ignored(RelPath(match)) {
					matches[n] = match
					n++
				}
			}
			targetParentRelPath := parentSourceRelPath.TargetRelPath(s.encryption.EncryptedSuffix())
			matches = matches[:n]
			for _, match := range matches {
				targetRelPath := targetParentRelPath.Join(RelPath(match))
				sourceStateEntry := &SourceStateRemove{
					targetRelPath: targetRelPath,
				}
				allSourceStateEntries[targetRelPath] = append(allSourceStateEntries[targetRelPath], sourceStateEntry)
			}
			return nil
		case info.Name() == templatesDirName:
			if err := s.addTemplatesDir(sourceAbsPath); err != nil {
				return err
			}
			return vfs.SkipDir
		case info.Name() == versionName:
			return s.addVersionFile(sourceAbsPath)
		case strings.HasPrefix(info.Name(), Prefix):
			fallthrough
		case strings.HasPrefix(info.Name(), ignorePrefix):
			if info.IsDir() {
				return vfs.SkipDir
			}
			return nil
		case info.IsDir():
			da := parseDirAttr(sourceName.String())
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).Join(RelPath(da.TargetName))
			if s.Ignored(targetRelPath) {
				return vfs.SkipDir
			}
			sourceStateEntry := s.newSourceStateDir(sourceRelPath, da)
			allSourceStateEntries[targetRelPath] = append(allSourceStateEntries[targetRelPath], sourceStateEntry)
			return nil
		case info.Mode().IsRegular():
			fa := parseFileAttr(sourceName.String(), s.encryption.EncryptedSuffix())
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).Join(RelPath(fa.TargetName))
			if s.Ignored(targetRelPath) {
				return nil
			}
			sourceStateEntry := s.newSourceStateFile(sourceRelPath, fa, targetRelPath)
			allSourceStateEntries[targetRelPath] = append(allSourceStateEntries[targetRelPath], sourceStateEntry)
			return nil
		default:
			return &errUnsupportedFileType{
				absPath: sourceAbsPath,
				mode:    info.Mode(),
			}
		}
	}); err != nil {
		return err
	}

	// Remove all ignored targets.
	for targetRelPath := range allSourceStateEntries {
		if s.Ignored(targetRelPath) {
			delete(allSourceStateEntries, targetRelPath)
		}
	}

	// Generate SourceStateRemoves for exact directories.
	for targetRelPath, sourceStateEntries := range allSourceStateEntries {
		if len(sourceStateEntries) != 1 {
			continue
		}

		switch sourceStateDir, ok := sourceStateEntries[0].(*SourceStateDir); {
		case !ok:
			continue
		case !sourceStateDir.Attr.Exact:
			continue
		}

		switch infos, err := s.system.ReadDir(s.destDirAbsPath.Join(targetRelPath)); {
		case err == nil:
			for _, info := range infos {
				name := info.Name()
				if name == "." || name == ".." {
					continue
				}
				destEntryRelPath := targetRelPath.Join(RelPath(name))
				if _, ok := allSourceStateEntries[destEntryRelPath]; ok {
					continue
				}
				if s.Ignored(destEntryRelPath) {
					continue
				}
				allSourceStateEntries[destEntryRelPath] = append(allSourceStateEntries[destEntryRelPath], &SourceStateRemove{
					targetRelPath: destEntryRelPath,
				})
			}
		case os.IsNotExist(err):
			// Do nothing.
		default:
			return err
		}
	}

	// Check for duplicate source entries with the same target name. Iterate
	// over the target names in order so that any error is deterministic.
	targetRelPaths := make(RelPaths, 0, len(allSourceStateEntries))
	for targetRelPath := range allSourceStateEntries {
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	sort.Sort(targetRelPaths)
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntries := allSourceStateEntries[targetRelPath]
		if len(sourceStateEntries) == 1 {
			continue
		}
		sourceRelPaths := make(SourceRelPaths, 0, len(sourceStateEntries))
		for _, sourceStateEntry := range sourceStateEntries {
			sourceRelPaths = append(sourceRelPaths, sourceStateEntry.SourceRelPath())
		}
		sort.Sort(sourceRelPaths)
		err = multierr.Append(err, &errDuplicateTarget{
			targetRelPath:  targetRelPath,
			sourceRelPaths: sourceRelPaths,
		})
	}
	if err != nil {
		return err
	}

	// Populate s.Entries with the unique source entry for each target.
	for targetRelPath, sourceEntries := range allSourceStateEntries {
		s.entries[targetRelPath] = sourceEntries[0]
	}

	return nil
}

// TargetRelPaths returns all of s's target relative paths in order.
func (s *SourceState) TargetRelPaths() RelPaths {
	targetRelPaths := make(RelPaths, 0, len(s.entries))
	for targetRelPath := range s.entries {
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	sort.Slice(targetRelPaths, func(i, j int) bool {
		orderI := s.entries[targetRelPaths[i]].Order()
		orderJ := s.entries[targetRelPaths[j]].Order()
		switch {
		case orderI < orderJ:
			return true
		case orderI == orderJ:
			return targetRelPaths[i] < targetRelPaths[j]
		default:
			return false
		}
	})
	return targetRelPaths
}

// TemplateData returns s's template data.
func (s *SourceState) TemplateData() map[string]interface{} {
	if s.templateData == nil {
		s.templateData = make(map[string]interface{})
		if s.defaultTemplateDataFunc != nil {
			RecursiveMerge(s.templateData, s.defaultTemplateDataFunc())
			s.defaultTemplateDataFunc = nil
		}
		RecursiveMerge(s.templateData, s.userTemplateData)
		RecursiveMerge(s.templateData, s.priorityTemplateData)
	}
	return s.templateData
}

// addPatterns executes the template at sourceAbsPath, interprets the result as
// a list of patterns, and adds all patterns found to patternSet.
func (s *SourceState) addPatterns(patternSet *patternSet, sourceAbsPath AbsPath, sourceRelPath SourceRelPath) error {
	data, err := s.executeTemplate(sourceAbsPath)
	if err != nil {
		return err
	}
	dir := sourceRelPath.Dir().TargetRelPath("")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var lineNumber int
	for scanner.Scan() {
		lineNumber++
		text := scanner.Text()
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
			text = mustTrimPrefix(text, "!")
		}
		pattern := string(dir.Join(RelPath(text)))
		if err := patternSet.add(pattern, include); err != nil {
			return fmt.Errorf("%s:%d: %w", sourceAbsPath, lineNumber, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	return nil
}

// addTemplateData adds all template data in sourceAbsPath to s.
func (s *SourceState) addTemplateData(sourceAbsPath AbsPath) error {
	_, name := sourceAbsPath.Split()
	suffix := mustTrimPrefix(string(name), dataName+".")
	format, ok := Formats[suffix]
	if !ok {
		return fmt.Errorf("%s: unknown format", sourceAbsPath)
	}
	data, err := s.system.ReadFile(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	var templateData map[string]interface{}
	if err := format.Unmarshal(data, &templateData); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	RecursiveMerge(s.userTemplateData, templateData)
	return nil
}

// addTemplatesDir adds all templates in templateDir to s.
func (s *SourceState) addTemplatesDir(templatesDirAbsPath AbsPath) error {
	return Walk(s.system, templatesDirAbsPath, func(templateAbsPath AbsPath, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		switch {
		case info.Mode().IsRegular():
			contents, err := s.system.ReadFile(templateAbsPath)
			if err != nil {
				return err
			}
			templateRelPath := templateAbsPath.MustTrimDirPrefix(templatesDirAbsPath)
			name := string(templateRelPath)
			tmpl, err := template.New(name).Option(s.templateOptions...).Funcs(s.templateFuncs).Parse(string(contents))
			if err != nil {
				return err
			}
			if s.templates == nil {
				s.templates = make(map[string]*template.Template)
			}
			s.templates[name] = tmpl
			return nil
		case info.IsDir():
			return nil
		default:
			return &errUnsupportedFileType{
				absPath: templateAbsPath,
				mode:    info.Mode(),
			}
		}
	})
}

// addVersionFile reads a .chezmoiversion file from source path and updates s's
// minimum version if it contains a more recent version than the current minimum
// version.
func (s *SourceState) addVersionFile(sourceAbsPath AbsPath) error {
	data, err := s.system.ReadFile(sourceAbsPath)
	if err != nil {
		return err
	}
	version, err := semver.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}
	if s.minVersion.LessThan(*version) {
		s.minVersion = *version
	}
	return nil
}

// applyAll updates targetDir in fs to match s.
func (s *SourceState) applyAll(targetSystem, destSystem System, persistentState PersistentState, targetDir AbsPath, options ApplyOptions) error {
	for _, targetRelPath := range s.TargetRelPaths() {
		switch err := s.Apply(targetSystem, destSystem, persistentState, targetDir, targetRelPath, options); {
		case errors.Is(err, Skip):
			continue
		case err != nil:
			return err
		}
	}
	return nil
}

// executeTemplate executes the template at path and returns the result.
func (s *SourceState) executeTemplate(templateAbsPath AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(templateAbsPath)
	if err != nil {
		return nil, err
	}
	return s.ExecuteTemplateData(string(templateAbsPath), data)
}

// newSourceStateDir returns a new SourceStateDir.
func (s *SourceState) newSourceStateDir(sourceRelPath SourceRelPath, dirAttr DirAttr) *SourceStateDir {
	targetStateDir := &TargetStateDir{
		perm: dirAttr.perm() &^ s.umask,
	}
	return &SourceStateDir{
		sourceRelPath:    sourceRelPath,
		Attr:             dirAttr,
		targetStateEntry: targetStateDir,
	}
}

func (s *SourceState) newCreateTargetStateEntryFunc(fileAttr FileAttr, sourceLazyContents *lazyContents) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contents, err := destSystem.ReadFile(destAbsPath)
		switch {
		case err == nil:
		case os.IsNotExist(err):
			contents, err = sourceLazyContents.Contents()
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
		return &TargetStateFile{
			lazyContents: newLazyContents(contents),
			empty:        true,
			perm:         fileAttr.perm() &^ s.umask,
		}, nil
	}
}

func (s *SourceState) newFileTargetStateEntryFunc(sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := func() ([]byte, error) {
			contents, err := sourceLazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fileAttr.Template {
				contents, err = s.ExecuteTemplateData(sourceRelPath.String(), contents)
				if err != nil {
					return nil, err
				}
			}
			return contents, nil
		}
		return &TargetStateFile{
			lazyContents: newLazyContentsFunc(contentsFunc),
			empty:        fileAttr.Empty,
			perm:         fileAttr.perm() &^ s.umask,
		}, nil
	}
}

func (s *SourceState) newModifyTargetStateEntryFunc(sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := func() (contents []byte, err error) {
			// Read the current contents of the target.
			var currentContents []byte
			currentContents, err = destSystem.ReadFile(destAbsPath)
			if err != nil && !os.IsNotExist(err) {
				return
			}

			// Compute the contents of the modifier.
			var modifierContents []byte
			modifierContents, err = sourceLazyContents.Contents()
			if err != nil {
				return
			}
			if fileAttr.Template {
				modifierContents, err = s.ExecuteTemplateData(sourceRelPath.String(), modifierContents)
				if err != nil {
					return
				}
			}

			// If the modifier is empty then return the current contents unchanged.
			if isEmpty(modifierContents) {
				contents = currentContents
				return
			}

			// Write the modifier to a temporary file.
			var tempFile *os.File
			tempFile, err = os.CreateTemp("", "*."+fileAttr.TargetName)
			if err != nil {
				return
			}
			defer func() {
				err = multierr.Append(err, os.RemoveAll(tempFile.Name()))
			}()
			if runtime.GOOS != "windows" {
				if err = tempFile.Chmod(0o700); err != nil {
					return
				}
			}
			_, err = tempFile.Write(modifierContents)
			err = multierr.Append(err, tempFile.Close())
			if err != nil {
				return
			}

			// Run the modifier on the current contents.
			//nolint:gosec
			cmd := exec.Command(tempFile.Name())
			cmd.Stdin = bytes.NewReader(currentContents)
			return destSystem.IdempotentCmdOutput(cmd)
		}
		return &TargetStateFile{
			lazyContents: newLazyContentsFunc(contentsFunc),
			perm:         fileAttr.perm() &^ s.umask,
		}, nil
	}
}

func (s *SourceState) newScriptTargetStateEntryFunc(sourceRelPath SourceRelPath, fileAttr FileAttr, targetRelPath RelPath, sourceLazyContents *lazyContents) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := func() ([]byte, error) {
			contents, err := sourceLazyContents.Contents()
			if err != nil {
				return nil, err
			}
			if fileAttr.Template {
				contents, err = s.ExecuteTemplateData(sourceRelPath.String(), contents)
				if err != nil {
					return nil, err
				}
			}
			return contents, nil
		}
		return &TargetStateScript{
			lazyContents: newLazyContentsFunc(contentsFunc),
			name:         targetRelPath,
			once:         fileAttr.Once,
		}, nil
	}
}

func (s *SourceState) newSymlinkTargetStateEntryFunc(sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		linknameFunc := func() (string, error) {
			linknameBytes, err := sourceLazyContents.Contents()
			if err != nil {
				return "", err
			}
			if fileAttr.Template {
				linknameBytes, err = s.ExecuteTemplateData(sourceRelPath.String(), linknameBytes)
				if err != nil {
					return "", err
				}
			}
			return string(bytes.TrimSpace(linknameBytes)), nil
		}
		return &TargetStateSymlink{
			lazyLinkname: newLazyLinknameFunc(linknameFunc),
		}, nil
	}
}

// newSourceStateFile returns a new SourceStateFile.
func (s *SourceState) newSourceStateFile(sourceRelPath SourceRelPath, fileAttr FileAttr, targetRelPath RelPath) *SourceStateFile {
	sourceLazyContents := newLazyContentsFunc(func() ([]byte, error) {
		contents, err := s.system.ReadFile(s.sourceDirAbsPath.Join(sourceRelPath.RelPath()))
		if err != nil {
			return nil, err
		}
		if fileAttr.Encrypted {
			contents, err = s.encryption.Decrypt(contents)
			if err != nil {
				return nil, err
			}
		}
		return contents, nil
	})

	var targetStateEntryFunc targetStateEntryFunc
	switch fileAttr.Type {
	case SourceFileTypeCreate:
		targetStateEntryFunc = s.newCreateTargetStateEntryFunc(fileAttr, sourceLazyContents)
	case SourceFileTypeFile:
		targetStateEntryFunc = s.newFileTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	case SourceFileTypeModify:
		targetStateEntryFunc = s.newModifyTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	case SourceFileTypeScript:
		targetStateEntryFunc = s.newScriptTargetStateEntryFunc(sourceRelPath, fileAttr, targetRelPath, sourceLazyContents)
	case SourceFileTypeSymlink:
		targetStateEntryFunc = s.newSymlinkTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	default:
		panic(fmt.Sprintf("%d: unsupported type", fileAttr.Type))
	}

	return &SourceStateFile{
		lazyContents:         sourceLazyContents,
		sourceRelPath:        sourceRelPath,
		Attr:                 fileAttr,
		targetStateEntryFunc: targetStateEntryFunc,
	}
}

func (s *SourceState) newSourceStateDirEntry(info os.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions) (SourceStateEntry, error) {
	dirAttr := DirAttr{
		TargetName: info.Name(),
		Exact:      options.Exact,
		Private:    isPrivate(info),
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelDirPath(RelPath(dirAttr.SourceName())))
	return &SourceStateDir{
		Attr:          dirAttr,
		sourceRelPath: sourceRelPath,
		targetStateEntry: &TargetStateDir{
			perm: 0o777 &^ s.umask,
		},
	}, nil
}

func (s *SourceState) newSourceStateFileEntryFromFile(actualStateFile *ActualStateFile, info os.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions) (SourceStateEntry, error) {
	fileAttr := FileAttr{
		TargetName: info.Name(),
		Empty:      options.Empty,
		Encrypted:  options.Encrypt,
		Executable: isExecutable(info),
		Private:    isPrivate(info),
		Template:   options.Template || options.AutoTemplate,
	}
	if options.Create {
		fileAttr.Type = SourceFileTypeCreate
	} else {
		fileAttr.Type = SourceFileTypeFile
	}
	contents, err := actualStateFile.Contents()
	if err != nil {
		return nil, err
	}
	if options.AutoTemplate {
		contents = autoTemplate(contents, s.TemplateData())
	}
	if len(contents) == 0 && !options.Empty {
		return nil, nil
	}
	if options.Encrypt {
		contents, err = s.encryption.Encrypt(contents)
		if err != nil {
			return nil, err
		}
	}
	lazyContents := newLazyContents(contents)
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(RelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))))
	return &SourceStateFile{
		Attr:          fileAttr,
		sourceRelPath: sourceRelPath,
		lazyContents:  lazyContents,
		targetStateEntry: &TargetStateFile{
			lazyContents: lazyContents,
			empty:        options.Empty,
			perm:         0o666 &^ s.umask,
		},
	}, nil
}

func (s *SourceState) newSourceStateFileEntryFromSymlink(actualStateSymlink *ActualStateSymlink, info os.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions) (SourceStateEntry, error) {
	linkname, err := actualStateSymlink.Linkname()
	if err != nil {
		return nil, err
	}
	contents := []byte(linkname)
	template := false
	switch {
	case options.AutoTemplate:
		contents = autoTemplate(contents, s.TemplateData())
		fallthrough
	case options.Template:
		template = true
	case !options.Template && options.TemplateSymlinks:
		switch {
		case strings.HasPrefix(linkname, string(s.sourceDirAbsPath)+"/"):
			contents = []byte("{{ .chezmoi.sourceDir }}/" + linkname[len(s.sourceDirAbsPath)+1:])
			template = true
		case strings.HasPrefix(linkname, string(s.destDirAbsPath)+"/"):
			contents = []byte("{{ .chezmoi.homeDir }}/" + linkname[len(s.destDirAbsPath)+1:])
			template = true
		}
	}
	contents = append(contents, '\n')
	lazyContents := newLazyContents(contents)
	fileAttr := FileAttr{
		TargetName: info.Name(),
		Type:       SourceFileTypeSymlink,
		Template:   template,
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(RelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))))
	return &SourceStateFile{
		Attr:          fileAttr,
		sourceRelPath: sourceRelPath,
		lazyContents:  lazyContents,
		targetStateEntry: &TargetStateFile{
			lazyContents: lazyContents,
			perm:         0o666 &^ s.umask,
		},
	}, nil
}

// sourceStateEntry returns a new SourceStateEntry based on actualStateEntry.
func (s *SourceState) sourceStateEntry(actualStateEntry ActualStateEntry, destAbsPath AbsPath, info os.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions) (SourceStateEntry, error) {
	switch actualStateEntry := actualStateEntry.(type) {
	case *ActualStateAbsent:
		return nil, fmt.Errorf("%s: not found", destAbsPath)
	case *ActualStateDir:
		return s.newSourceStateDirEntry(info, parentSourceRelPath, options)
	case *ActualStateFile:
		return s.newSourceStateFileEntryFromFile(actualStateEntry, info, parentSourceRelPath, options)
	case *ActualStateSymlink:
		return s.newSourceStateFileEntryFromSymlink(actualStateEntry, info, parentSourceRelPath, options)
	default:
		panic(fmt.Sprintf("%T: unsupported type", actualStateEntry))
	}
}
