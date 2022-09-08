package chezmoi

// FIXME implement externals in chezmoi source state format

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	vfs "github.com/twpayne/go-vfs/v4"
	"go.uber.org/multierr"
	"golang.org/x/exp/maps"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

// An ExternalType is a type of external source.
type ExternalType string

// ExternalTypes.
const (
	ExternalTypeArchive ExternalType = "archive"
	ExternalTypeFile    ExternalType = "file"
	ExternalTypeGitRepo ExternalType = "git-repo"
)

// An External is an external source.
type External struct {
	Type       ExternalType `json:"type" toml:"type" yaml:"type"`
	Encrypted  bool         `json:"encrypted" toml:"encrypted" yaml:"encrypted"`
	Exact      bool         `json:"exact" toml:"exact" yaml:"exact"`
	Executable bool         `json:"executable" toml:"executable" yaml:"executable"`
	Clone      struct {
		Args []string `json:"args" toml:"args" yaml:"args"`
	} `json:"clone" toml:"clone" yaml:"clone"`
	Exclude []string `json:"exclude" toml:"exclude" yaml:"exclude"`
	Filter  struct {
		Command string   `json:"command" toml:"command" yaml:"command"`
		Args    []string `json:"args" toml:"args" yaml:"args"`
	} `json:"filter" toml:"filter" yaml:"filter"`
	Format  ArchiveFormat `json:"format" toml:"format" yaml:"format"`
	Include []string      `json:"include" toml:"include" yaml:"include"`
	Pull    struct {
		Args []string `json:"args" toml:"args" yaml:"args"`
	} `json:"pull" toml:"pull" yaml:"pull"`
	RefreshPeriod   Duration `json:"refreshPeriod" toml:"refreshPeriod" yaml:"refreshPeriod"`
	StripComponents int      `json:"stripComponents" toml:"stripComponents" yaml:"stripComponents"`
	URL             string   `json:"url" toml:"url" yaml:"url"`
	sourceAbsPath   AbsPath
}

// A externalCacheEntry is an external cache entry.
type externalCacheEntry struct {
	URL  string    `json:"url" time:"url"`
	Time time.Time `json:"time" yaml:"time"`
	Data []byte    `json:"data" yaml:"data"`
}

var externalCacheFormat = formatGzippedJSON{}

// A SourceState is a source state.
type SourceState struct {
	sync.Mutex
	root                    sourceStateEntryTreeNode
	removeDirs              map[RelPath]struct{}
	baseSystem              System
	system                  System
	sourceDirAbsPath        AbsPath
	destDirAbsPath          AbsPath
	cacheDirAbsPath         AbsPath
	umask                   fs.FileMode
	encryption              Encryption
	ignore                  *patternSet
	remove                  *patternSet
	interpreters            map[string]*Interpreter
	httpClient              *http.Client
	logger                  *zerolog.Logger
	version                 semver.Version
	mode                    Mode
	defaultTemplateDataFunc func() map[string]any
	templateDataOnly        bool
	readTemplateData        bool
	userTemplateData        map[string]any
	priorityTemplateData    map[string]any
	templateData            map[string]any
	templateFuncs           template.FuncMap
	templateOptions         []string
	templates               map[string]*template.Template
	externals               map[RelPath]*External
	ignoredRelPaths         map[RelPath]struct{}
}

// A SourceStateOption sets an option on a source state.
type SourceStateOption func(*SourceState)

// WithBaseSystem sets the base system.
func WithBaseSystem(baseSystem System) SourceStateOption {
	return func(s *SourceState) {
		s.baseSystem = baseSystem
	}
}

// WithCacheDir sets the cache directory.
func WithCacheDir(cacheDirAbsPath AbsPath) SourceStateOption {
	return func(s *SourceState) {
		s.cacheDirAbsPath = cacheDirAbsPath
	}
}

// WithDefaultTemplateDataFunc sets the default template data function.
func WithDefaultTemplateDataFunc(defaultTemplateDataFunc func() map[string]any) SourceStateOption {
	return func(s *SourceState) {
		s.defaultTemplateDataFunc = defaultTemplateDataFunc
	}
}

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

// WithHTTPClient sets the HTTP client.
func WithHTTPClient(httpClient *http.Client) SourceStateOption {
	return func(s *SourceState) {
		s.httpClient = httpClient
	}
}

// WithInterpreters sets the interpreters.
func WithInterpreters(interpreters map[string]*Interpreter) SourceStateOption {
	return func(s *SourceState) {
		s.interpreters = interpreters
	}
}

// WithLogger sets the logger.
func WithLogger(logger *zerolog.Logger) SourceStateOption {
	return func(s *SourceState) {
		s.logger = logger
	}
}

// WithMode sets the mode.
func WithMode(mode Mode) SourceStateOption {
	return func(s *SourceState) {
		s.mode = mode
	}
}

// WithPriorityTemplateData adds priority template data.
func WithPriorityTemplateData(priorityTemplateData map[string]any) SourceStateOption {
	return func(s *SourceState) {
		RecursiveMerge(s.priorityTemplateData, priorityTemplateData)
	}
}

// WithReadTemplateData sets whether to read .chezmoidata.<format> files.
func WithReadTemplateData(readTemplateData bool) SourceStateOption {
	return func(s *SourceState) {
		s.readTemplateData = readTemplateData
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

// WithTemplateDataOnly sets whether only template data should be read.
func WithTemplateDataOnly(templateDataOnly bool) SourceStateOption {
	return func(s *SourceState) {
		s.templateDataOnly = templateDataOnly
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

// WithVersion sets the version.
func WithVersion(version semver.Version) SourceStateOption {
	return func(s *SourceState) {
		s.version = version
	}
}

// A targetStateEntryFunc returns a TargetStateEntry based on reading an AbsPath
// on a System.
type targetStateEntryFunc func(System, AbsPath) (TargetStateEntry, error)

// NewSourceState creates a new source state with the given options.
func NewSourceState(options ...SourceStateOption) *SourceState {
	s := &SourceState{
		removeDirs:           make(map[RelPath]struct{}),
		umask:                Umask,
		encryption:           NoEncryption{},
		ignore:               newPatternSet(),
		remove:               newPatternSet(),
		httpClient:           http.DefaultClient,
		logger:               &log.Logger,
		readTemplateData:     true,
		priorityTemplateData: make(map[string]any),
		userTemplateData:     make(map[string]any),
		templateOptions:      DefaultTemplateOptions,
		templates:            make(map[string]*template.Template),
		externals:            make(map[RelPath]*External),
		ignoredRelPaths:      make(map[RelPath]struct{}),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// A PreAddFunc is called before a new source state entry is added.
type PreAddFunc func(targetRelPath RelPath) error

// A ReplaceFunc is called before a source state entry is replaced.
type ReplaceFunc func(targetRelPath RelPath, newSourceStateEntry, oldSourceStateEntry SourceStateEntry) error

// AddOptions are options to SourceState.Add.
type AddOptions struct {
	AutoTemplate     bool          // Automatically create templates, if possible.
	Create           bool          // Add create_ entries instead of normal entries.
	Encrypt          bool          // Encrypt files.
	EncryptedSuffix  string        // Suffix for encrypted files.
	Exact            bool          // Add the exact_ attribute to added directories.
	Include          *EntryTypeSet // Only add types in this set.
	PreAddFunc       PreAddFunc    // Function to be called before a source entry is added.
	RemoveDir        RelPath       // Directory to remove before adding.
	ReplaceFunc      ReplaceFunc   // Function to be called before a source entry is replaced.
	Template         bool          // Add the .tmpl attribute to added files.
	TemplateSymlinks bool          // Add symlinks with targets in the source or home directories as templates.
}

// Add adds destAbsPathInfos to s.
func (s *SourceState) Add(
	sourceSystem System, persistentState PersistentState, destSystem System, destAbsPathInfos map[AbsPath]fs.FileInfo,
	options *AddOptions,
) error {
	type sourceUpdate struct {
		destAbsPath    AbsPath
		entryState     *EntryState
		sourceRelPaths []SourceRelPath
	}

	destAbsPaths := AbsPaths(maps.Keys(destAbsPathInfos))
	sort.Sort(destAbsPaths)

	sourceUpdates := make([]sourceUpdate, 0, len(destAbsPathInfos))
	newSourceStateEntries := make(map[SourceRelPath]SourceStateEntry)
	newSourceStateEntriesByTargetRelPath := make(map[RelPath]SourceStateEntry)
	nonEmptyDirs := make(map[SourceRelPath]struct{})
	dirRenames := make(map[AbsPath]AbsPath)
DESTABSPATH:
	for _, destAbsPath := range destAbsPaths {
		destAbsPathInfo := destAbsPathInfos[destAbsPath]
		if !options.Include.IncludeFileInfo(destAbsPathInfo) {
			continue
		}
		targetRelPath := destAbsPath.MustTrimDirPrefix(s.destDirAbsPath)

		if s.Ignore(targetRelPath) {
			continue
		}

		// Find the target's parent directory in the source state.
		var parentSourceRelPath SourceRelPath
		if targetParentRelPath := targetRelPath.Dir(); targetParentRelPath == DotRelPath {
			parentSourceRelPath = SourceRelPath{}
		} else if parentEntry, ok := newSourceStateEntriesByTargetRelPath[targetParentRelPath]; ok {
			parentSourceRelPath = parentEntry.SourceRelPath()
		} else if parentEntry := s.root.Get(targetParentRelPath); parentEntry != nil {
			parentSourceRelPath = parentEntry.SourceRelPath()
		} else {
			return fmt.Errorf("%s: parent directory not in source state", destAbsPath)
		}
		nonEmptyDirs[parentSourceRelPath] = struct{}{}

		actualStateEntry, err := NewActualStateEntry(destSystem, destAbsPath, destAbsPathInfo, nil)
		if err != nil {
			return err
		}
		newSourceStateEntry, err := s.sourceStateEntry(
			actualStateEntry, destAbsPath, destAbsPathInfo, parentSourceRelPath, options,
		)
		if err != nil {
			return err
		}
		if newSourceStateEntry == nil {
			continue
		}

		if options.PreAddFunc != nil {
			switch err := options.PreAddFunc(targetRelPath); {
			case errors.Is(err, Skip):
				continue DESTABSPATH
			case err != nil:
				return err
			}
		}

		sourceEntryRelPath := newSourceStateEntry.SourceRelPath()

		entryState, err := actualStateEntry.EntryState()
		if err != nil {
			return err
		}
		update := sourceUpdate{
			destAbsPath:    destAbsPath,
			entryState:     entryState,
			sourceRelPaths: []SourceRelPath{sourceEntryRelPath},
		}

		if oldSourceStateEntry := s.root.Get(targetRelPath); oldSourceStateEntry != nil {
			oldSourceEntryRelPath := oldSourceStateEntry.SourceRelPath()
			if !oldSourceEntryRelPath.Empty() && oldSourceEntryRelPath != sourceEntryRelPath {
				if options.ReplaceFunc != nil {
					switch err := options.ReplaceFunc(targetRelPath, newSourceStateEntry, oldSourceStateEntry); {
					case errors.Is(err, Skip):
						continue DESTABSPATH
					case err != nil:
						return err
					}
				}

				// If both the new and old source state entries are directories
				// but the name has changed, rename to avoid losing the
				// directory's contents.
				_, newIsDir := newSourceStateEntry.(*SourceStateDir)
				_, oldIsDir := oldSourceStateEntry.(*SourceStateDir)
				if newIsDir && oldIsDir {
					oldSourceAbsPath := s.sourceDirAbsPath.Join(oldSourceEntryRelPath.RelPath())
					newSourceAbsPath := s.sourceDirAbsPath.Join(sourceEntryRelPath.RelPath())
					dirRenames[oldSourceAbsPath] = newSourceAbsPath
					continue DESTABSPATH
				}

				// Otherwise, remove the old entry.
				newSourceStateEntries[oldSourceEntryRelPath] = &SourceStateRemove{
					origin: SourceStateOriginRemove{},
				}
				update.sourceRelPaths = append(update.sourceRelPaths, oldSourceEntryRelPath)
			}
		}

		newSourceStateEntries[sourceEntryRelPath] = newSourceStateEntry
		newSourceStateEntriesByTargetRelPath[targetRelPath] = newSourceStateEntry

		sourceUpdates = append(sourceUpdates, update)
	}

	// Create .keep files in empty added directories.
	for sourceEntryRelPath, sourceStateEntry := range newSourceStateEntries {
		if _, ok := sourceStateEntry.(*SourceStateDir); !ok {
			continue
		}
		if _, ok := nonEmptyDirs[sourceEntryRelPath]; ok {
			continue
		}

		dotKeepFileRelPath := sourceEntryRelPath.Join(NewSourceRelPath(".keep"))

		dotKeepFileSourceUpdate := sourceUpdate{
			entryState: &EntryState{
				Type: EntryStateTypeFile,
				Mode: 0o666 &^ s.umask,
			},
			sourceRelPaths: []SourceRelPath{dotKeepFileRelPath},
		}
		sourceUpdates = append(sourceUpdates, dotKeepFileSourceUpdate)

		newSourceStateEntries[dotKeepFileRelPath] = &SourceStateFile{
			targetStateEntry: &TargetStateFile{
				empty: true,
				perm:  0o666 &^ s.umask,
			},
		}
	}

	var sourceRoot sourceStateEntryTreeNode
	for sourceRelPath, sourceStateEntry := range newSourceStateEntries {
		sourceRoot.Set(sourceRelPath.RelPath(), sourceStateEntry)
	}

	// Simulate removing a directory by creating SourceStateRemove entries for
	// all existing source state entries that are in options.RemoveDir and not
	// in the new source state.
	if options.RemoveDir != EmptyRelPath {
		_ = s.root.ForEach(EmptyRelPath, func(targetRelPath RelPath, sourceStateEntry SourceStateEntry) error {
			if !targetRelPath.HasDirPrefix(options.RemoveDir) {
				return nil
			}
			if _, ok := newSourceStateEntriesByTargetRelPath[targetRelPath]; ok {
				return nil
			}
			sourceRelPath := sourceStateEntry.SourceRelPath()
			sourceRoot.Set(sourceRelPath.RelPath(), &SourceStateRemove{
				sourceRelPath: sourceRelPath,
				targetRelPath: targetRelPath,
			})
			update := sourceUpdate{
				destAbsPath: s.destDirAbsPath.Join(targetRelPath),
				entryState: &EntryState{
					Type: EntryStateTypeRemove,
				},
				sourceRelPaths: []SourceRelPath{sourceRelPath},
			}
			sourceUpdates = append(sourceUpdates, update)
			return nil
		})
	}

	targetSourceState := &SourceState{
		root: sourceRoot,
	}

	for _, sourceUpdate := range sourceUpdates {
		for _, sourceRelPath := range sourceUpdate.sourceRelPaths {
			err := targetSourceState.Apply(
				sourceSystem, sourceSystem, NullPersistentState{}, s.sourceDirAbsPath, sourceRelPath.RelPath(),
				ApplyOptions{
					Include: options.Include,
					Umask:   s.umask,
				},
			)
			if err != nil {
				return err
			}
		}
		if !sourceUpdate.destAbsPath.Empty() {
			if err := persistentStateSet(
				persistentState, EntryStateBucket, sourceUpdate.destAbsPath.Bytes(), sourceUpdate.entryState,
			); err != nil {
				return err
			}
		}
	}

	// Rename directories last because updates assume that directory names have
	// not changed. Rename directories in reverse order so children are renamed
	// before their parents.
	oldDirAbsPaths := make([]AbsPath, 0, len(dirRenames))
	for oldDirAbsPath := range dirRenames {
		oldDirAbsPaths = append(oldDirAbsPaths, oldDirAbsPath)
	}
	sort.Slice(oldDirAbsPaths, func(i, j int) bool {
		return oldDirAbsPaths[j].Less(oldDirAbsPaths[i])
	})
	for _, oldDirAbsPath := range oldDirAbsPaths {
		newDirAbsPath := dirRenames[oldDirAbsPath]
		if err := sourceSystem.Rename(oldDirAbsPath, newDirAbsPath); err != nil {
			return err
		}
	}

	return nil
}

// AddDestAbsPathInfos adds an fs.FileInfo to destAbsPathInfos for destAbsPath
// and any of its parents which are not already known.
func (s *SourceState) AddDestAbsPathInfos(
	destAbsPathInfos map[AbsPath]fs.FileInfo, system System, destAbsPath AbsPath, fileInfo fs.FileInfo,
) error {
	for {
		if _, err := destAbsPath.TrimDirPrefix(s.destDirAbsPath); err != nil {
			return err
		}

		if _, ok := destAbsPathInfos[destAbsPath]; ok {
			return nil
		}

		if fileInfo == nil {
			var err error
			fileInfo, err = system.Lstat(destAbsPath)
			if err != nil {
				return err
			}
		}
		destAbsPathInfos[destAbsPath] = fileInfo

		parentAbsPath := destAbsPath.Dir()
		if parentAbsPath == s.destDirAbsPath {
			return nil
		}
		parentRelPath := parentAbsPath.MustTrimDirPrefix(s.destDirAbsPath)
		if s.root.Get(parentRelPath) != nil {
			return nil
		}

		destAbsPath = parentAbsPath
		fileInfo = nil
	}
}

// A PreApplyFunc is called before a target is applied.
type PreApplyFunc func(
	targetRelPath RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *EntryState,
) error

// ApplyOptions are options to SourceState.ApplyAll and SourceState.ApplyOne.
type ApplyOptions struct {
	Include      *EntryTypeSet
	PreApplyFunc PreApplyFunc
	Umask        fs.FileMode
}

// Apply updates targetRelPath in targetDirAbsPath in destSystem to match s.
func (s *SourceState) Apply(
	targetSystem, destSystem System, persistentState PersistentState, targetDirAbsPath AbsPath, targetRelPath RelPath,
	options ApplyOptions,
) error {
	sourceStateEntry := s.root.Get(targetRelPath)

	if !options.Include.IncludeEncrypted() {
		if sourceStateFile, ok := sourceStateEntry.(*SourceStateFile); ok && sourceStateFile.Attr.Encrypted {
			return nil
		}
	}

	if !options.Include.IncludeExternals() {
		if _, ok := sourceStateEntry.Origin().(*External); ok {
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

	targetAbsPath := targetDirAbsPath.Join(targetRelPath)

	targetEntryState, err := targetStateEntry.EntryState(options.Umask)
	if err != nil {
		return err
	}

	switch skip, err := targetStateEntry.SkipApply(persistentState, targetAbsPath); {
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
		ok, err := persistentStateGet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), &entryState)
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

		// If the target entry state matches the actual entry state, but not the
		// last written entry state then silently update the last written entry
		// state. This handles the case where the user makes identical edits to
		// the source and target states: instead of reporting a diff with
		// respect to the last written state, we record the effect of the last
		// apply as the last written state.
		if targetEntryState.Equivalent(actualEntryState) && !lastWrittenEntryState.Equivalent(actualEntryState) {
			err := persistentStateSet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), targetEntryState)
			if err != nil {
				return err
			}
			lastWrittenEntryState = targetEntryState
		}

		err = options.PreApplyFunc(targetRelPath, targetEntryState, lastWrittenEntryState, actualEntryState)
		if err != nil {
			return err
		}
	}

	if changed, err := targetStateEntry.Apply(targetSystem, persistentState, actualStateEntry); err != nil {
		return err
	} else if !changed {
		return nil
	}

	return persistentStateSet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), targetEntryState)
}

// Contains returns the source state entry for targetRelPath.
func (s *SourceState) Contains(targetRelPath RelPath) bool {
	return s.root.Get(targetRelPath) != nil
}

// Encryption returns s's encryption.
func (s *SourceState) Encryption() Encryption {
	return s.encryption
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

	// Temporarily set .chezmoi.sourceFile to the name of the template.
	templateData := s.TemplateData()
	if chezmoiTemplateData, ok := templateData["chezmoi"].(map[string]any); ok {
		chezmoiTemplateData["sourceFile"] = name
		defer delete(chezmoiTemplateData, "sourceFile")
	}

	builder := strings.Builder{}
	if err = tmpl.ExecuteTemplate(&builder, name, templateData); err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}

// ForEach calls f for each source state entry.
func (s *SourceState) ForEach(f func(RelPath, SourceStateEntry) error) error {
	return s.root.ForEach(EmptyRelPath, func(targetRelPath RelPath, entry SourceStateEntry) error {
		return f(targetRelPath, entry)
	})
}

// Ignore returns if targetRelPath should be ignored.
func (s *SourceState) Ignore(targetRelPath RelPath) bool {
	s.Lock()
	defer s.Unlock()
	ignore := s.ignore.match(targetRelPath.String()) == patternSetMatchInclude
	if ignore {
		s.ignoredRelPaths[targetRelPath] = struct{}{}
	}
	return ignore
}

// Ignored returns all ignored RelPaths.
func (s *SourceState) Ignored() RelPaths {
	relPaths := make(RelPaths, 0, len(s.ignoredRelPaths))
	for relPath := range s.ignoredRelPaths {
		relPaths = append(relPaths, relPath)
	}
	sort.Sort(relPaths)
	return relPaths
}

// MustEntry returns the source state entry associated with targetRelPath, and
// panics if it does not exist.
func (s *SourceState) MustEntry(targetRelPath RelPath) SourceStateEntry {
	sourceStateEntry := s.root.Get(targetRelPath)
	if sourceStateEntry == nil {
		panic(fmt.Sprintf("%s: not in source state", targetRelPath))
	}
	return sourceStateEntry
}

// PostApply performs all updates required after s.Apply.
func (s *SourceState) PostApply(targetSystem System, targetDirAbsPath AbsPath, targetRelPaths RelPaths) error {
	// Remove empty directories with the remove_ attribute. This assumes that
	// targetRelPaths is already sorted and iterates in reverse order so that
	// children are removed before their parents.
TARGET:
	for i := len(targetRelPaths) - 1; i >= 0; i-- {
		targetRelPath := targetRelPaths[i]
		if _, ok := s.removeDirs[targetRelPath]; !ok {
			continue
		}

		// Ensure that we are attempting to remove a directory, not any other entry type.
		targetAbsPath := targetDirAbsPath.Join(targetRelPath)
		switch fileInfo, err := targetSystem.Stat(targetAbsPath); {
		case errors.Is(err, fs.ErrNotExist):
			continue TARGET
		case err != nil:
			return err
		case !fileInfo.IsDir():
			return fmt.Errorf("%s: not a directory", targetAbsPath)
		}

		// Attempt to remove the directory, but ignore any "not exist" or "not
		// empty" errors.
		switch err := targetSystem.Remove(targetAbsPath); {
		case err == nil:
			// Do nothing.
		case errors.Is(err, fs.ErrExist):
			// Do nothing.
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		default:
			return err
		}
	}

	return nil
}

// ReadOptions are options to SourceState.Read.
type ReadOptions struct {
	RefreshExternals bool
	TimeNow          func() time.Time
}

// Read reads the source state from the source directory.
func (s *SourceState) Read(ctx context.Context, options *ReadOptions) error {
	switch fileInfo, err := s.system.Stat(s.sourceDirAbsPath); {
	case errors.Is(err, fs.ErrNotExist):
		return nil
	case err != nil:
		return err
	case !fileInfo.IsDir():
		return fmt.Errorf("%s: not a directory", s.sourceDirAbsPath)
	}

	// Read all source entries.
	var allSourceStateEntriesMu sync.Mutex
	allSourceStateEntries := make(map[RelPath][]SourceStateEntry)
	addSourceStateEntries := func(relPath RelPath, sourceStateEntries ...SourceStateEntry) {
		allSourceStateEntriesMu.Lock()
		allSourceStateEntries[relPath] = append(allSourceStateEntries[relPath], sourceStateEntries...)
		allSourceStateEntriesMu.Unlock()
	}
	walkFunc := func(sourceAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if sourceAbsPath == s.sourceDirAbsPath {
			return nil
		}

		// Follow symlinks in the source directory.
		if fileInfo.Mode().Type() == fs.ModeSymlink {
			// Some programs (notably emacs) use invalid symlinks as lockfiles.
			// To avoid following them and getting an ENOENT error, check first
			// if this is an entry that we will ignore anyway.
			if strings.HasPrefix(fileInfo.Name(), ignorePrefix) && !strings.HasPrefix(fileInfo.Name(), Prefix) {
				return nil
			}
			fileInfo, err = s.system.Stat(sourceAbsPath)
			if err != nil {
				return err
			}
		}

		sourceRelPath := SourceRelPath{
			relPath: sourceAbsPath.MustTrimDirPrefix(s.sourceDirAbsPath),
			isDir:   fileInfo.IsDir(),
		}
		parentSourceRelPath, sourceName := sourceRelPath.Split()

		switch {
		case isPrefixDotFormat(fileInfo.Name(), dataName):
			if !s.readTemplateData {
				return nil
			}
			return s.addTemplateData(sourceAbsPath)
		case fileInfo.Name() == TemplatesDirName:
			if err := s.addTemplatesDir(ctx, sourceAbsPath); err != nil {
				return err
			}
			return vfs.SkipDir
		case s.templateDataOnly:
			return nil
		case isPrefixDotFormat(fileInfo.Name(), externalName):
			return s.addExternal(sourceAbsPath)
		case fileInfo.Name() == ignoreName:
			return s.addPatterns(s.ignore, sourceAbsPath, parentSourceRelPath)
		case fileInfo.Name() == removeName:
			return s.addPatterns(s.remove, sourceAbsPath, parentSourceRelPath)
		case fileInfo.Name() == scriptsDirName:
			scriptsDirSourceStateEntries, err := s.readScriptsDir(ctx, sourceAbsPath)
			if err != nil {
				return err
			}
			for relPath, scriptSourceStateEntries := range scriptsDirSourceStateEntries {
				addSourceStateEntries(relPath, scriptSourceStateEntries...)
			}
			return vfs.SkipDir
		case fileInfo.Name() == VersionName:
			return s.readVersionFile(sourceAbsPath)
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			fallthrough
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return vfs.SkipDir
			}
			return nil
		case fileInfo.IsDir():
			da := parseDirAttr(sourceName.String())
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).JoinString(da.TargetName)
			if s.Ignore(targetRelPath) {
				return vfs.SkipDir
			}
			sourceStateDir := s.newSourceStateDir(sourceAbsPath, sourceRelPath, da)
			addSourceStateEntries(targetRelPath, sourceStateDir)
			if sourceStateDir.Attr.Remove {
				s.Lock()
				s.removeDirs[targetRelPath] = struct{}{}
				s.Unlock()
			}
			return nil
		case fileInfo.Mode().IsRegular():
			fa := parseFileAttr(sourceName.String(), s.encryption.EncryptedSuffix())
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).JoinString(fa.TargetName)
			if s.Ignore(targetRelPath) {
				return nil
			}
			var sourceStateEntry SourceStateEntry
			targetRelPath, sourceStateEntry = s.newSourceStateFile(sourceAbsPath, sourceRelPath, fa, targetRelPath)
			addSourceStateEntries(targetRelPath, sourceStateEntry)
			return nil
		default:
			return &unsupportedFileTypeError{
				absPath: sourceAbsPath,
				mode:    fileInfo.Mode(),
			}
		}
	}
	if err := WalkSourceDir(s.system, s.sourceDirAbsPath, walkFunc); err != nil {
		return err
	}

	if s.templateDataOnly {
		return nil
	}

	// Read externals.
	externalRelPaths := make(RelPaths, 0, len(s.externals))
	for externalRelPath := range s.externals {
		externalRelPaths = append(externalRelPaths, externalRelPath)
	}
	sort.Sort(externalRelPaths)
	for _, externalRelPath := range externalRelPaths {
		if s.Ignore(externalRelPath) {
			continue
		}
		external := s.externals[externalRelPath]
		parentRelPath, _ := externalRelPath.Split()
		var parentSourceRelPath SourceRelPath
		switch parentSourceStateEntry, err := s.root.MkdirAll(parentRelPath, external, s.umask); {
		case err != nil:
			return err
		case parentSourceStateEntry != nil:
			parentSourceRelPath = parentSourceStateEntry.SourceRelPath()
		}
		externalSourceStateEntries, err := s.readExternal(ctx, externalRelPath, parentSourceRelPath, external, options)
		if err != nil {
			return err
		}
		for targetRelPath, sourceStateEntries := range externalSourceStateEntries {
			if s.Ignore(targetRelPath) {
				continue
			}
			allSourceStateEntries[targetRelPath] = append(allSourceStateEntries[targetRelPath], sourceStateEntries...)
		}
	}

	// Remove all ignored targets.
	for targetRelPath := range allSourceStateEntries {
		if s.Ignore(targetRelPath) {
			delete(allSourceStateEntries, targetRelPath)
		}
	}

	// Generate SourceStateRemoves for existing targets.
	matches, err := s.remove.glob(s.system.UnderlyingFS(), ensureSuffix(s.destDirAbsPath.String(), "/"))
	if err != nil {
		return err
	}
	for _, match := range matches {
		targetRelPath := NewRelPath(match)
		if s.Ignore(targetRelPath) {
			continue
		}
		sourceStateEntry := &SourceStateRemove{
			origin:        SourceStateOriginRemove{},
			sourceRelPath: NewSourceRelPath(".chezmoiremove"),
			targetRelPath: targetRelPath,
		}
		allSourceStateEntries[targetRelPath] = append(allSourceStateEntries[targetRelPath], sourceStateEntry)
	}

	// Generate SourceStateRemoves for exact directories.
	for targetRelPath, sourceStateEntries := range allSourceStateEntries {
		if len(sourceStateEntries) != 1 {
			continue
		}

		sourceStateDir, ok := sourceStateEntries[0].(*SourceStateDir)
		switch {
		case !ok:
			continue
		case !sourceStateDir.Attr.Exact:
			continue
		}

		switch fileInfos, err := s.system.ReadDir(s.destDirAbsPath.Join(targetRelPath)); {
		case err == nil:
			for _, fileInfo := range fileInfos {
				name := fileInfo.Name()
				if name == "." || name == ".." {
					continue
				}
				destEntryRelPath := targetRelPath.JoinString(name)
				if _, ok := allSourceStateEntries[destEntryRelPath]; ok {
					continue
				}
				if s.Ignore(destEntryRelPath) {
					continue
				}
				sourceStateRemove := &SourceStateRemove{
					origin:        sourceStateDir.Origin(),
					sourceRelPath: sourceStateDir.sourceRelPath,
					targetRelPath: destEntryRelPath,
				}
				allSourceStateEntries[destEntryRelPath] = append(allSourceStateEntries[destEntryRelPath], sourceStateRemove)
			}
		case errors.Is(err, fs.ErrNotExist):
			// Do nothing.
		case errors.Is(err, syscall.ENOTDIR):
			// Do nothing.
		default:
			return err
		}
	}

	// Generate SourceStateCommands for git-repo externals.
	var gitRepoExternalRelPaths RelPaths
	for externalRelPath, external := range s.externals {
		if external.Type == ExternalTypeGitRepo {
			gitRepoExternalRelPaths = append(gitRepoExternalRelPaths, externalRelPath)
		}
	}
	sort.Sort(gitRepoExternalRelPaths)
	for _, externalRelPath := range gitRepoExternalRelPaths {
		external := s.externals[externalRelPath]
		destAbsPath := s.destDirAbsPath.Join(externalRelPath)
		switch _, err := s.system.Lstat(destAbsPath); {
		case errors.Is(err, fs.ErrNotExist):
			// FIXME add support for using builtin git
			args := []string{"clone"}
			args = append(args, external.Clone.Args...)
			args = append(args, external.URL, destAbsPath.String())
			cmd := exec.Command("git", args...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			sourceStateCommand := &SourceStateCommand{
				cmd:           cmd,
				origin:        external,
				forceRefresh:  options.RefreshExternals,
				refreshPeriod: external.RefreshPeriod,
			}
			allSourceStateEntries[externalRelPath] = append(allSourceStateEntries[externalRelPath], sourceStateCommand)
		case err != nil:
			return err
		default:
			// FIXME add support for using builtin git
			args := []string{"pull"}
			args = append(args, external.Pull.Args...)
			cmd := exec.Command("git", args...)
			cmd.Dir = destAbsPath.String()
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			sourceStateCommand := &SourceStateCommand{
				cmd:           cmd,
				origin:        external,
				forceRefresh:  options.RefreshExternals,
				refreshPeriod: external.RefreshPeriod,
			}
			allSourceStateEntries[externalRelPath] = append(allSourceStateEntries[externalRelPath], sourceStateCommand)
		}
	}

	// Check for inconsistent source entries. Iterate over the target names in
	// order so that any error is deterministic.
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

		// Allow duplicate equivalent source entries for directories.
		if allEquivalentDirs(sourceStateEntries) {
			continue
		}

		origins := make([]string, 0, len(sourceStateEntries))
		for _, sourceStateEntry := range sourceStateEntries {
			origins = append(origins, sourceStateEntry.Origin().OriginString())
		}
		sort.Strings(origins)
		err = multierr.Append(err, &inconsistentStateError{
			targetRelPath: targetRelPath,
			origins:       origins,
		})
	}
	if err != nil {
		return err
	}

	// Populate s.Entries with the unique source entry for each target.
	for targetRelPath, sourceEntries := range allSourceStateEntries {
		s.root.Set(targetRelPath, sourceEntries[0])
	}

	return nil
}

// TargetRelPaths returns all of s's target relative paths in order.
func (s *SourceState) TargetRelPaths() []RelPath {
	entries := s.root.Map()
	targetRelPaths := make([]RelPath, 0, len(entries))
	for targetRelPath := range entries {
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	sort.Slice(targetRelPaths, func(i, j int) bool {
		orderI := entries[targetRelPaths[i]].Order()
		orderJ := entries[targetRelPaths[j]].Order()
		switch {
		case orderI < orderJ:
			return true
		case orderI == orderJ:
			return targetRelPaths[i].Less(targetRelPaths[j])
		default:
			return false
		}
	})
	return targetRelPaths
}

// TemplateData returns s's template data.
func (s *SourceState) TemplateData() map[string]any {
	if s.templateData == nil {
		s.templateData = make(map[string]any)
		if s.defaultTemplateDataFunc != nil {
			RecursiveMerge(s.templateData, s.defaultTemplateDataFunc())
			s.defaultTemplateDataFunc = nil
		}
		RecursiveMerge(s.templateData, s.userTemplateData)
		RecursiveMerge(s.templateData, s.priorityTemplateData)
	}
	return s.templateData
}

// addExternal adds external source entries to s.
func (s *SourceState) addExternal(sourceAbsPath AbsPath) error {
	parentAbsPath, _ := sourceAbsPath.Split()

	parentRelPath, err := parentAbsPath.TrimDirPrefix(s.sourceDirAbsPath)
	if err != nil {
		return err
	}
	parentSourceRelPath := NewSourceRelDirPath(parentRelPath.String())
	parentTargetSourceRelPath := parentSourceRelPath.TargetRelPath(s.encryption.EncryptedSuffix())

	format, ok := Formats[strings.TrimPrefix(sourceAbsPath.Ext(), ".")]
	if !ok {
		return fmt.Errorf("%s: unknown format", sourceAbsPath)
	}
	data, err := s.executeTemplate(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	externals := make(map[string]*External)
	if err := format.Unmarshal(data, &externals); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	s.Lock()
	defer s.Unlock()
	for path, external := range externals {
		if filepath.IsAbs(path) {
			return fmt.Errorf("%s: %s: path is not relative", sourceAbsPath, path)
		}
		targetRelPath := parentTargetSourceRelPath.JoinString(path)
		if _, ok := s.externals[targetRelPath]; ok {
			return fmt.Errorf("%s: duplicate externals", targetRelPath)
		}
		external.sourceAbsPath = sourceAbsPath
		s.externals[targetRelPath] = external
	}
	return nil
}

// addPatterns executes the template at sourceAbsPath, interprets the result as
// a list of patterns, and adds all patterns found to patternSet.
func (s *SourceState) addPatterns(patternSet *patternSet, sourceAbsPath AbsPath, sourceRelPath SourceRelPath) error {
	data, err := s.executeTemplate(sourceAbsPath)
	if err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	dir := sourceRelPath.Dir().TargetRelPath("")
	scanner := bufio.NewScanner(bytes.NewReader(data))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		text := scanner.Text()
		text, _, _ = strings.Cut(text, "#")
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		include := patternSetInclude
		if strings.HasPrefix(text, "!") {
			include = patternSetExclude
			text = mustTrimPrefix(text, "!")
		}
		pattern := dir.JoinString(text).String()
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
	format, ok := Formats[strings.TrimPrefix(sourceAbsPath.Ext(), ".")]
	if !ok {
		return fmt.Errorf("%s: unknown format", sourceAbsPath)
	}
	data, err := s.system.ReadFile(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	var templateData map[string]any
	if err := format.Unmarshal(data, &templateData); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	s.Lock()
	RecursiveMerge(s.userTemplateData, templateData)
	s.Unlock()
	return nil
}

// addTemplatesDir adds all templates in templatesDirAbsPath to s.
func (s *SourceState) addTemplatesDir(ctx context.Context, templatesDirAbsPath AbsPath) error {
	walkFunc := func(ctx context.Context, templateAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
		if templateAbsPath == templatesDirAbsPath {
			return nil
		}
		if err == nil && fileInfo.Mode().Type() == fs.ModeSymlink {
			fileInfo, err = s.system.Stat(templateAbsPath)
		}
		switch {
		case err != nil:
			return err
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			return fmt.Errorf("%s: not allowed in %s directory", TemplatesDirName, templateAbsPath)
		case strings.HasPrefix(fileInfo.Name(), "."):
			if fileInfo.IsDir() {
				return vfs.SkipDir
			}
			return nil
		case fileInfo.Mode().IsRegular():
			contents, err := s.system.ReadFile(templateAbsPath)
			if err != nil {
				return err
			}
			templateRelPath := templateAbsPath.MustTrimDirPrefix(templatesDirAbsPath)
			name := templateRelPath.String()
			tmpl, err := template.New(name).Option(s.templateOptions...).Funcs(s.templateFuncs).Parse(string(contents))
			if err != nil {
				return err
			}
			s.Lock()
			s.templates[name] = tmpl
			s.Unlock()
			return nil
		case fileInfo.IsDir():
			return nil
		default:
			return &unsupportedFileTypeError{
				absPath: templateAbsPath,
				mode:    fileInfo.Mode(),
			}
		}
	}
	return concurrentWalkSourceDir(ctx, s.system, templatesDirAbsPath, walkFunc)
}

// executeTemplate executes the template at path and returns the result.
func (s *SourceState) executeTemplate(templateAbsPath AbsPath) ([]byte, error) {
	data, err := s.system.ReadFile(templateAbsPath)
	if err != nil {
		return nil, err
	}
	return s.ExecuteTemplateData(templateAbsPath.String(), data)
}

// getExternalDataRaw returns the raw data for external at externalRelPath,
// possibly from the external cache.
func (s *SourceState) getExternalDataRaw(
	ctx context.Context, externalRelPath RelPath, external *External, options *ReadOptions,
) ([]byte, error) {
	var now time.Time
	if options != nil && options.TimeNow != nil {
		now = options.TimeNow()
	} else {
		now = time.Now()
	}
	now = now.UTC()

	cacheKey := hex.EncodeToString(SHA256Sum([]byte(external.URL)))
	cachedDataAbsPath := s.cacheDirAbsPath.JoinString("external", cacheKey+"."+externalCacheFormat.Name())
	if options == nil || !options.RefreshExternals {
		if data, err := s.system.ReadFile(cachedDataAbsPath); err == nil {
			var externalCacheEntry externalCacheEntry
			if err := externalCacheFormat.Unmarshal(data, &externalCacheEntry); err == nil {
				if externalCacheEntry.URL == external.URL {
					if external.RefreshPeriod == 0 || externalCacheEntry.Time.Add(time.Duration(external.RefreshPeriod)).After(now) {
						return externalCacheEntry.Data, nil
					}
				}
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, external.URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := chezmoilog.LogHTTPRequest(s.logger, s.httpClient, req)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
		return nil, fmt.Errorf("%s: %s: %s", externalRelPath, external.URL, resp.Status)
	}

	cachedExternalData, err := externalCacheFormat.Marshal(&externalCacheEntry{
		URL:  external.URL,
		Time: now,
		Data: data,
	})
	if err != nil {
		return nil, err
	}
	if err := MkdirAll(s.baseSystem, cachedDataAbsPath.Dir(), 0o700); err != nil {
		return nil, err
	}
	if err := s.baseSystem.WriteFile(cachedDataAbsPath, cachedExternalData, 0o600); err != nil {
		return nil, err
	}

	return data, nil
}

// getExternalDataRaw reads the external data for externalRelPath from
// external.URL.
func (s *SourceState) getExternalData(
	ctx context.Context, externalRelPath RelPath, external *External, options *ReadOptions,
) ([]byte, error) {
	data, err := s.getExternalDataRaw(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, err
	}

	if external.Encrypted {
		data, err = s.encryption.Decrypt(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", externalRelPath, external.URL, err)
		}
	}

	if external.Filter.Command != "" {
		//nolint:gosec
		cmd := exec.Command(external.Filter.Command, external.Filter.Args...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Stderr = os.Stderr
		data, err = chezmoilog.LogCmdOutput(cmd)
		if err != nil {
			return nil, fmt.Errorf("%s: %s: %w", externalRelPath, external.URL, err)
		}
	}

	return data, nil
}

// newSourceStateDir returns a new SourceStateDir.
func (s *SourceState) newSourceStateDir(absPath AbsPath, sourceRelPath SourceRelPath, dirAttr DirAttr) *SourceStateDir {
	targetStateDir := &TargetStateDir{
		perm: dirAttr.perm() &^ s.umask,
	}
	return &SourceStateDir{
		origin:           SourceStateOriginAbsPath(absPath),
		sourceRelPath:    sourceRelPath,
		Attr:             dirAttr,
		targetStateEntry: targetStateDir,
	}
}

// newCreateTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with sourceLazyContents if the file does not already exist, or returns
// the actual file's contents unchanged if the file already exists.
func (s *SourceState) newCreateTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents,
) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		var lazyContents *lazyContents
		switch contents, err := destSystem.ReadFile(destAbsPath); {
		case err == nil:
			lazyContents = newLazyContents(contents)
		case errors.Is(err, fs.ErrNotExist):
			lazyContents = newLazyContentsFunc(func() ([]byte, error) {
				contents, err = sourceLazyContents.Contents()
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
			})
		default:
			return nil, err
		}
		return &TargetStateFile{
			lazyContents: lazyContents,
			empty:        true,
			perm:         fileAttr.perm() &^ s.umask,
		}, nil
	}
}

// newFileTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with sourceLazyContents.
func (s *SourceState) newFileTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents,
) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		if s.mode == ModeSymlink && !fileAttr.Encrypted && !fileAttr.Executable && !fileAttr.Private && !fileAttr.Template {
			switch contents, err := sourceLazyContents.Contents(); {
			case err != nil:
				return nil, err
			case isEmpty(contents) && !fileAttr.Empty:
				return &TargetStateRemove{}, nil
			default:
				linkname := normalizeLinkname(s.sourceDirAbsPath.Join(sourceRelPath.RelPath()).String())
				return &TargetStateSymlink{
					lazyLinkname: newLazyLinkname(linkname),
				}, nil
			}
		}
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

// newModifyTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with the contents modified by running the sourceLazyContents script.
func (s *SourceState) newModifyTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents, interpreter *Interpreter,
) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := func() (contents []byte, err error) {
			// Read the current contents of the target.
			var currentContents []byte
			currentContents, err = destSystem.ReadFile(destAbsPath)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
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
			if tempFile, err = os.CreateTemp("", "*."+fileAttr.TargetName); err != nil {
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
			multierr.AppendInvoke(&err, multierr.Close(tempFile))
			if err != nil {
				return
			}

			// Run the modifier on the current contents.
			cmd := interpreter.ExecCommand(tempFile.Name())
			cmd.Stdin = bytes.NewReader(currentContents)
			cmd.Stderr = os.Stderr
			contents, err = chezmoilog.LogCmdOutput(cmd)
			return
		}
		return &TargetStateFile{
			lazyContents: newLazyContentsFunc(contentsFunc),
			overwrite:    true,
			perm:         fileAttr.perm() &^ s.umask,
		}, nil
	}
}

// newRemoveTargetStateEntryFunc returns a targetStateEntryFunc that removes a
// target.
func (s *SourceState) newRemoveTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr,
) targetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		return &TargetStateRemove{}, nil
	}
}

// newScriptTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// script with sourceLazyContents.
func (s *SourceState) newScriptTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr, targetRelPath RelPath, sourceLazyContents *lazyContents,
	interpreter *Interpreter,
) targetStateEntryFunc {
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
			condition:    fileAttr.Condition,
			interpreter:  interpreter,
		}, nil
	}
}

// newSymlinkTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// symlink with the linkname sourceLazyContents.
func (s *SourceState) newSymlinkTargetStateEntryFunc(
	sourceRelPath SourceRelPath, fileAttr FileAttr, sourceLazyContents *lazyContents,
) targetStateEntryFunc {
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
			linkname := normalizeLinkname(string(bytes.TrimSpace(linknameBytes)))
			return linkname, nil
		}
		return &TargetStateSymlink{
			lazyLinkname: newLazyLinknameFunc(linknameFunc),
		}, nil
	}
}

// newSourceStateFile returns a possibly new target RalPath and a new
// SourceStateFile.
func (s *SourceState) newSourceStateFile(
	absPath AbsPath, sourceRelPath SourceRelPath, fileAttr FileAttr, targetRelPath RelPath,
) (RelPath, *SourceStateFile) {
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
		targetStateEntryFunc = s.newCreateTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	case SourceFileTypeFile:
		targetStateEntryFunc = s.newFileTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	case SourceFileTypeModify:
		// If the target has an extension, determine if it indicates an
		// interpreter to use.
		ext := strings.ToLower(strings.TrimPrefix(targetRelPath.Ext(), "."))
		interpreter := s.interpreters[ext]
		if interpreter != nil {
			// For modify scripts, the script extension is not considered part
			// of the target name, so remove it.
			targetRelPath = targetRelPath.Slice(0, targetRelPath.Len()-len(ext)-1)
		}
		targetStateEntryFunc = s.newModifyTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents, interpreter)
	case SourceFileTypeRemove:
		targetStateEntryFunc = s.newRemoveTargetStateEntryFunc(sourceRelPath, fileAttr)
	case SourceFileTypeScript:
		// If the script has an extension, determine if it indicates an
		// interpreter to use.
		ext := strings.ToLower(strings.TrimPrefix(targetRelPath.Ext(), "."))
		interpreter := s.interpreters[ext]
		targetStateEntryFunc = s.newScriptTargetStateEntryFunc(
			sourceRelPath, fileAttr, targetRelPath, sourceLazyContents, interpreter,
		)
	case SourceFileTypeSymlink:
		targetStateEntryFunc = s.newSymlinkTargetStateEntryFunc(sourceRelPath, fileAttr, sourceLazyContents)
	default:
		panic(fmt.Sprintf("%d: unsupported type", fileAttr.Type))
	}

	return targetRelPath, &SourceStateFile{
		lazyContents:         sourceLazyContents,
		origin:               SourceStateOriginAbsPath(absPath),
		sourceRelPath:        sourceRelPath,
		Attr:                 fileAttr,
		targetStateEntryFunc: targetStateEntryFunc,
	}
}

// newSourceStateDirEntry returns a SourceStateEntry constructed from a directory in s.
//
// We return a SourceStateEntry rather than a *SourceStateDir to simplify nil checks later.
func (s *SourceState) newSourceStateDirEntry(
	actualStateDir *ActualStateDir, fileInfo fs.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions,
) (SourceStateEntry, error) {
	dirAttr := DirAttr{
		TargetName: fileInfo.Name(),
		Exact:      options.Exact,
		Private:    isPrivate(fileInfo),
		ReadOnly:   isReadOnly(fileInfo),
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelDirPath(dirAttr.SourceName()))
	return &SourceStateDir{
		Attr:          dirAttr,
		origin:        actualStateDir,
		sourceRelPath: sourceRelPath,
		targetStateEntry: &TargetStateDir{
			perm: 0o777 &^ s.umask,
		},
	}, nil
}

// newSourceStateFileEntryFromFile returns a SourceStateEntry constructed from a
// file in s.
//
// We return a SourceStateEntry rather than a *SourceStateFile to simplify nil
// checks later.
func (s *SourceState) newSourceStateFileEntryFromFile(
	actualStateFile *ActualStateFile, fileInfo fs.FileInfo, parentSourceRelPath SourceRelPath, options *AddOptions,
) (SourceStateEntry, error) {
	fileAttr := FileAttr{
		TargetName: fileInfo.Name(),
		Encrypted:  options.Encrypt,
		Executable: isExecutable(fileInfo),
		Private:    isPrivate(fileInfo),
		ReadOnly:   isReadOnly(fileInfo),
		Template:   options.Template,
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
		var replacements bool
		contents, replacements = autoTemplate(contents, s.TemplateData())
		if replacements {
			fileAttr.Template = true
		}
	}
	if len(contents) == 0 {
		fileAttr.Empty = true
	}
	if options.Encrypt {
		contents, err = s.encryption.Encrypt(contents)
		if err != nil {
			return nil, err
		}
	}
	lazyContents := newLazyContents(contents)
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
	return &SourceStateFile{
		Attr:          fileAttr,
		origin:        actualStateFile,
		sourceRelPath: sourceRelPath,
		lazyContents:  lazyContents,
		targetStateEntry: &TargetStateFile{
			lazyContents: lazyContents,
			empty:        len(contents) == 0,
			perm:         0o666 &^ s.umask,
		},
	}, nil
}

// newSourceStateFileEntryFromSymlink returns a SourceStateEntry constructed
// from a symlink in s.
//
// We return a SourceStateEntry rather than a *SourceStateFile to simplify nil
// checks later.
func (s *SourceState) newSourceStateFileEntryFromSymlink(
	actualStateSymlink *ActualStateSymlink, fileInfo fs.FileInfo, parentSourceRelPath SourceRelPath,
	options *AddOptions,
) (SourceStateEntry, error) {
	linkname, err := actualStateSymlink.Linkname()
	if err != nil {
		return nil, err
	}
	contents := []byte(linkname)
	template := false
	switch {
	case options.AutoTemplate:
		contents, template = autoTemplate(contents, s.TemplateData())
	case options.Template:
		template = true
	case !options.Template && options.TemplateSymlinks:
		switch {
		case strings.HasPrefix(linkname, s.sourceDirAbsPath.String()+"/"):
			contents = []byte("{{ .chezmoi.sourceDir }}/" + linkname[s.sourceDirAbsPath.Len()+1:])
			template = true
		case strings.HasPrefix(linkname, s.destDirAbsPath.String()+"/"):
			contents = []byte("{{ .chezmoi.homeDir }}/" + linkname[s.destDirAbsPath.Len()+1:])
			template = true
		}
	}
	contents = append(contents, '\n')
	lazyContents := newLazyContents(contents)
	fileAttr := FileAttr{
		TargetName: fileInfo.Name(),
		Type:       SourceFileTypeSymlink,
		Template:   template,
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
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

// readExternal reads an external and returns its SourceStateEntries.
func (s *SourceState) readExternal(
	ctx context.Context, externalRelPath RelPath, parentSourceRelPath SourceRelPath, external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	switch external.Type {
	case ExternalTypeArchive:
		return s.readExternalArchive(ctx, externalRelPath, parentSourceRelPath, external, options)
	case ExternalTypeFile:
		return s.readExternalFile(ctx, externalRelPath, parentSourceRelPath, external, options)
	case ExternalTypeGitRepo:
		return nil, nil
	default:
		return nil, fmt.Errorf("%s: unknown external type: %s", externalRelPath, external.Type)
	}
}

// readExternalArchive reads an external archive and returns its
// SourceStateEntries.
func (s *SourceState) readExternalArchive(
	ctx context.Context, externalRelPath RelPath, parentSourceRelPath SourceRelPath, external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	data, err := s.getExternalData(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(external.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: %s: %w", externalRelPath, external.URL, err)
	}
	urlPath := url.Path
	if external.Encrypted {
		urlPath = strings.TrimSuffix(urlPath, s.encryption.EncryptedSuffix())
	}
	dirAttr := DirAttr{
		TargetName: externalRelPath.Base(),
		Exact:      external.Exact,
	}
	sourceStateDir := &SourceStateDir{
		Attr:          dirAttr,
		origin:        external,
		sourceRelPath: parentSourceRelPath.Join(NewSourceRelPath(dirAttr.SourceName())),
		targetStateEntry: &TargetStateDir{
			perm: 0o777 &^ s.umask,
		},
	}
	sourceStateEntries := map[RelPath][]SourceStateEntry{
		externalRelPath: {sourceStateDir},
	}

	format := external.Format
	if format == ArchiveFormatUnknown {
		format = GuessArchiveFormat(urlPath, data)
	}

	patternSet := newPatternSet()
	for _, includePattern := range external.Include {
		if err := patternSet.add(includePattern, patternSetInclude); err != nil {
			return nil, err
		}
	}
	for _, excludePattern := range external.Exclude {
		if err := patternSet.add(excludePattern, patternSetExclude); err != nil {
			return nil, err
		}
	}

	sourceRelPaths := make(map[RelPath]SourceRelPath)
	if err := WalkArchive(data, format, func(name string, fileInfo fs.FileInfo, r io.Reader, linkname string) error {
		// Perform matching against the name before stripping any components,
		// otherwise it is not possible to differentiate between
		// identically-named files at the same level.
		if patternSet.match(name) == patternSetMatchExclude {
			// In case that `name` is a directory which matched an explicit
			// exclude pattern, return fs.SkipDir to exclude not just the
			// directory itself but also everything it contains (recursively).
			if fileInfo.IsDir() && len(patternSet.excludePatterns) > 0 {
				return fs.SkipDir
			}
			return nil
		}

		if external.StripComponents > 0 {
			components := strings.Split(name, "/")
			if len(components) <= external.StripComponents {
				return nil
			}
			name = path.Join(components[external.StripComponents:]...)
		}
		if name == "" {
			return nil
		}
		targetRelPath := externalRelPath.JoinString(name)

		if s.Ignore(targetRelPath) {
			return nil
		}

		dirTargetRelPath, _ := targetRelPath.Split()
		dirSourceRelPath := sourceRelPaths[dirTargetRelPath]

		var sourceStateEntry SourceStateEntry
		switch {
		case fileInfo.IsDir():
			targetStateEntry := &TargetStateDir{
				perm: fileInfo.Mode().Perm() &^ s.umask,
			}
			dirAttr := DirAttr{
				TargetName: fileInfo.Name(),
				Exact:      external.Exact,
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			sourceStateEntry = &SourceStateDir{
				Attr:             dirAttr,
				origin:           external,
				sourceRelPath:    parentSourceRelPath.Join(dirSourceRelPath, NewSourceRelPath(dirAttr.SourceName())),
				targetStateEntry: targetStateEntry,
			}
		case fileInfo.Mode()&fs.ModeType == 0:
			contents, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			lazyContents := newLazyContents(contents)
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeFile,
				Empty:      fileInfo.Size() == 0,
				Executable: isExecutable(fileInfo),
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			sourceRelPath := NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))
			targetStateEntry := &TargetStateFile{
				lazyContents: lazyContents,
				empty:        fileAttr.Empty,
				perm:         fileAttr.perm() &^ s.umask,
			}
			sourceStateEntry = &SourceStateFile{
				lazyContents:     lazyContents,
				Attr:             fileAttr,
				origin:           external,
				sourceRelPath:    parentSourceRelPath.Join(dirSourceRelPath, sourceRelPath),
				targetStateEntry: targetStateEntry,
			}
		case fileInfo.Mode()&fs.ModeType == fs.ModeSymlink:
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeSymlink,
			}
			sourceRelPath := NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))
			targetStateEntry := &TargetStateSymlink{
				lazyLinkname: newLazyLinkname(linkname),
			}
			sourceStateEntry = &SourceStateFile{
				Attr:             fileAttr,
				origin:           external,
				sourceRelPath:    parentSourceRelPath.Join(dirSourceRelPath, sourceRelPath),
				targetStateEntry: targetStateEntry,
			}
		default:
			return fmt.Errorf("%s: unsupported mode %o", name, fileInfo.Mode()&fs.ModeType)
		}
		sourceStateEntries[targetRelPath] = append(sourceStateEntries[targetRelPath], sourceStateEntry)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("%s: %s: %w", externalRelPath, external.URL, err)
	}

	return sourceStateEntries, nil
}

// readExternalFile reads an external file and returns its SourceStateEntries.
func (s *SourceState) readExternalFile(
	ctx context.Context, externalRelPath RelPath, parentSourceRelPath SourceRelPath, external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	lazyContents := newLazyContentsFunc(func() ([]byte, error) {
		return s.getExternalData(ctx, externalRelPath, external, options)
	})
	fileAttr := FileAttr{
		Empty:      true,
		Executable: external.Executable,
	}
	targetStateEntry := &TargetStateFile{
		lazyContents: lazyContents,
		empty:        fileAttr.Empty,
		perm:         fileAttr.perm() &^ s.umask,
	}
	sourceStateEntry := &SourceStateFile{
		origin:           external,
		sourceRelPath:    parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))),
		targetStateEntry: targetStateEntry,
	}
	return map[RelPath][]SourceStateEntry{
		externalRelPath: {sourceStateEntry},
	}, nil
}

// readScriptsDir reads all scripts in scriptsDirAbsPath.
func (s *SourceState) readScriptsDir(
	ctx context.Context, scriptsDirAbsPath AbsPath,
) (map[RelPath][]SourceStateEntry, error) {
	var allSourceStateEntriesMu sync.Mutex
	allSourceStateEntries := make(map[RelPath][]SourceStateEntry)
	addSourceStateEntry := func(relPath RelPath, sourceStateEntry SourceStateEntry) {
		allSourceStateEntriesMu.Lock()
		allSourceStateEntries[relPath] = append(allSourceStateEntries[relPath], sourceStateEntry)
		allSourceStateEntriesMu.Unlock()
	}
	walkFunc := func(ctx context.Context, sourceAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if sourceAbsPath == scriptsDirAbsPath {
			return nil
		}

		// Follow symlinks in the source directory.
		if fileInfo.Mode().Type() == fs.ModeSymlink {
			// Some programs (notably emacs) use invalid symlinks as lockfiles.
			// To avoid following them and getting an ENOENT error, check first
			// if this is an entry that we will ignore anyway.
			if strings.HasPrefix(fileInfo.Name(), ignorePrefix) && !strings.HasPrefix(fileInfo.Name(), Prefix) {
				return nil
			}
			fileInfo, err = s.system.Stat(sourceAbsPath)
			if err != nil {
				return err
			}
		}

		sourceRelPath := SourceRelPath{
			relPath: sourceAbsPath.MustTrimDirPrefix(s.sourceDirAbsPath),
			isDir:   fileInfo.IsDir(),
		}
		parentSourceRelPath, sourceName := sourceRelPath.Split()

		switch {
		case err != nil:
			return err
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			return fmt.Errorf("%s: not allowed in %s directory", sourceAbsPath, scriptsDirName)
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return vfs.SkipDir
			}
			return nil
		case fileInfo.IsDir():
			return nil
		case fileInfo.Mode().IsRegular():
			fa := parseFileAttr(sourceName.String(), s.encryption.EncryptedSuffix())
			if fa.Type != SourceFileTypeScript {
				return fmt.Errorf("%s: not a script", sourceAbsPath)
			}
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).JoinString(fa.TargetName)
			if s.Ignore(targetRelPath) {
				return nil
			}
			var sourceStateEntry SourceStateEntry
			targetRelPath, sourceStateEntry = s.newSourceStateFile(sourceAbsPath, sourceRelPath, fa, targetRelPath)
			addSourceStateEntry(targetRelPath, sourceStateEntry)
			return nil
		default:
			return &unsupportedFileTypeError{
				absPath: sourceAbsPath,
				mode:    fileInfo.Mode(),
			}
		}
	}
	if err := concurrentWalkSourceDir(ctx, s.system, scriptsDirAbsPath, walkFunc); err != nil {
		return nil, err
	}
	return allSourceStateEntries, nil
}

// readVersionFile reads a .chezmoiversion file from sourceAbsPath and returns
// an error if the version is newer that s's version.
func (s *SourceState) readVersionFile(sourceAbsPath AbsPath) error {
	data, err := s.system.ReadFile(sourceAbsPath)
	if err != nil {
		return err
	}
	version, err := semver.NewVersion(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("%s: %q: %w", sourceAbsPath, data, err)
	}
	var zeroVersion semver.Version
	if s.version != zeroVersion && s.version.LessThan(*version) {
		return &TooOldError{
			Have: s.version,
			Need: *version,
		}
	}
	return nil
}

// sourceStateEntry returns a new SourceStateEntry based on actualStateEntry.
func (s *SourceState) sourceStateEntry(
	actualStateEntry ActualStateEntry, destAbsPath AbsPath, fileInfo fs.FileInfo, parentSourceRelPath SourceRelPath,
	options *AddOptions,
) (SourceStateEntry, error) {
	switch actualStateEntry := actualStateEntry.(type) {
	case *ActualStateAbsent:
		return nil, fmt.Errorf("%s: not found", destAbsPath)
	case *ActualStateDir:
		return s.newSourceStateDirEntry(actualStateEntry, fileInfo, parentSourceRelPath, options)
	case *ActualStateFile:
		return s.newSourceStateFileEntryFromFile(actualStateEntry, fileInfo, parentSourceRelPath, options)
	case *ActualStateSymlink:
		return s.newSourceStateFileEntryFromSymlink(actualStateEntry, fileInfo, parentSourceRelPath, options)
	default:
		panic(fmt.Sprintf("%T: unsupported type", actualStateEntry))
	}
}

func (e *External) Path() AbsPath {
	return e.sourceAbsPath
}

func (e *External) OriginString() string {
	return e.URL + " defined in " + e.sourceAbsPath.String()
}

// allEquivalentDirs returns if sourceStateEntries are all equivalent
// directories.
func allEquivalentDirs(sourceStateEntries []SourceStateEntry) bool {
	sourceStateDir0, ok := sourceStateEntries[0].(*SourceStateDir)
	if !ok {
		return false
	}
	for _, sourceStateEntry := range sourceStateEntries[1:] {
		sourceStateDir, ok := sourceStateEntry.(*SourceStateDir)
		if !ok {
			return false
		}
		if sourceStateDir0.Attr != sourceStateDir.Attr {
			return false
		}
	}
	return true
}
