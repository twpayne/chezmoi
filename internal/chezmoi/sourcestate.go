package chezmoi

// FIXME implement externals in chezmoi source state format

import (
	"bytes"
	"cmp"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/coreos/go-semver/semver"
	"github.com/mitchellh/copystructure"

	"chezmoi.io/chezmoi/internal/chezmoierrors"
	"chezmoi.io/chezmoi/internal/chezmoilog"
	"chezmoi.io/chezmoi/internal/chezmoiset"
)

// An ExternalType is a type of external source.
type ExternalType string

// ExternalTypes.
const (
	ExternalTypeArchive     ExternalType = "archive"
	ExternalTypeArchiveFile ExternalType = "archive-file"
	ExternalTypeFile        ExternalType = "file"
	ExternalTypeGitRepo     ExternalType = "git-repo"
)

var (
	commentRx                       = regexp.MustCompile(`(?:\A|\s+)#.*(?:\r?\n)?$`)
	lineEndingRx                    = regexp.MustCompile(`(?m)(?:\r\n|\r|\n)`)
	modifyTemplateRx                = regexp.MustCompile(`(?m)^.*chezmoi:modify-template.*$(?:\r?\n)?`)
	templateDirectiveRx             = regexp.MustCompile(`(?m)^.*?chezmoi:template:(.*)$(?:\r?\n)?`)
	templateDirectiveKeyValuePairRx = regexp.MustCompile(`\s*(\S+)=("(?:[^"]|\\")*"|\S+)`)

	// AppleDouble constants.
	appleDoubleNamePrefix     = "._"
	appleDoubleContentsPrefix = []byte{0x00, 0x05, 0x16, 0x07, 0x00, 0x02, 0x00, 0x00}
)

type ExternalArchive struct {
	ExtractAppleDoubleFiles bool `json:"extractAppleDoubleFiles" toml:"extractAppleDoubleFiles" yaml:"extractAppleDoubleFiles"`
}

type ExternalChecksum struct {
	MD5       HexBytes `json:"md5"       toml:"md5"       yaml:"md5"`
	RIPEMD160 HexBytes `json:"ripemd160" toml:"ripemd160" yaml:"ripemd160"`
	SHA1      HexBytes `json:"sha1"      toml:"sha1"      yaml:"sha1"`
	SHA256    HexBytes `json:"sha256"    toml:"sha256"    yaml:"sha256"`
	SHA384    HexBytes `json:"sha384"    toml:"sha384"    yaml:"sha384"`
	SHA512    HexBytes `json:"sha512"    toml:"sha512"    yaml:"sha512"`
	Size      int      `json:"size"      toml:"size"      yaml:"size"`
}

type ExternalClone struct {
	Args []string `json:"args" toml:"args" yaml:"args"`
}

type ExternalFilter struct {
	Command string   `json:"command" toml:"command" yaml:"command"`
	Args    []string `json:"args"    toml:"args"    yaml:"args"`
}

type ExternalPull struct {
	Args []string `json:"args" toml:"args" yaml:"args"`
}

// A WarnFunc is a function that warns the user.
type WarnFunc func(string, ...any)

// An External is an external source.
type External struct {
	Type            ExternalType      `json:"type"            toml:"type"            yaml:"type"`
	Encrypted       bool              `json:"encrypted"       toml:"encrypted"       yaml:"encrypted"`
	Exact           bool              `json:"exact"           toml:"exact"           yaml:"exact"`
	Executable      bool              `json:"executable"      toml:"executable"      yaml:"executable"`
	Private         bool              `json:"private"         toml:"private"         yaml:"private"`
	ReadOnly        bool              `json:"readonly"        toml:"readonly"        yaml:"readonly"`
	Checksum        ExternalChecksum  `json:"checksum"        toml:"checksum"        yaml:"checksum"`
	Clone           ExternalClone     `json:"clone"           toml:"clone"           yaml:"clone"`
	Decompress      CompressionFormat `json:"decompress"      toml:"decompress"      yaml:"decompress"`
	Exclude         []string          `json:"exclude"         toml:"exclude"         yaml:"exclude"`
	Filter          ExternalFilter    `json:"filter"          toml:"filter"          yaml:"filter"`
	Format          ArchiveFormat     `json:"format"          toml:"format"          yaml:"format"`
	Archive         ExternalArchive   `json:"archive"         toml:"archive"         yaml:"archive"`
	Include         []string          `json:"include"         toml:"include"         yaml:"include"`
	ArchivePath     string            `json:"path"            toml:"path"            yaml:"path"`
	Pull            ExternalPull      `json:"pull"            toml:"pull"            yaml:"pull"`
	RefreshPeriod   Duration          `json:"refreshPeriod"   toml:"refreshPeriod"   yaml:"refreshPeriod"`
	StripComponents int               `json:"stripComponents" toml:"stripComponents" yaml:"stripComponents"`
	URL             string            `json:"url"             toml:"url"             yaml:"url"`
	URLs            []string          `json:"urls"            toml:"urls"            yaml:"urls"`
	sourceAbsPath   AbsPath
}

// A SourceState is a source state.
type SourceState struct {
	mutex                   sync.Mutex
	root                    SourceStateEntryTreeNode
	removeDirs              chezmoiset.Set[RelPath]
	baseSystem              System
	system                  System
	sourceDirAbsPath        AbsPath
	destDirAbsPath          AbsPath
	cacheDirAbsPath         AbsPath
	createScriptTempDirOnce sync.Once
	scriptTempDirAbsPath    AbsPath
	umask                   fs.FileMode
	encryption              Encryption
	ignore                  *PatternSet
	remove                  *PatternSet
	interpreters            map[string]Interpreter
	httpClient              *http.Client
	logger                  *slog.Logger
	version                 semver.Version
	mode                    Mode
	defaultTemplateDataFunc func() map[string]any
	templateDataOnly        bool
	readTemplateData        bool
	readTemplates           bool
	defaultTemplateData     map[string]any
	userTemplateData        map[string]any
	priorityTemplateData    map[string]any
	templateData            map[string]any
	templateFuncs           template.FuncMap
	templateOptions         []string
	templates               map[string]*Template
	externals               map[RelPath][]*External
	ignoredRelPaths         chezmoiset.Set[RelPath]
	warnFunc                WarnFunc
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
func WithInterpreters(interpreters map[string]Interpreter) SourceStateOption {
	return func(s *SourceState) {
		s.interpreters = interpreters
	}
}

// WithLogger sets the logger.
func WithLogger(logger *slog.Logger) SourceStateOption {
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

// WithReadTemplates sets whether to read .chezmoitemplates directories.
func WithReadTemplates(readTemplates bool) SourceStateOption {
	return func(s *SourceState) {
		s.readTemplates = readTemplates
	}
}

// WithScriptTempDir sets the source directory.
func WithScriptTempDir(scriptDirAbsPath AbsPath) SourceStateOption {
	return func(s *SourceState) {
		s.scriptTempDirAbsPath = scriptDirAbsPath
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

// WithUmask sets the umask.
func WithUmask(umask fs.FileMode) SourceStateOption {
	return func(s *SourceState) {
		s.umask = umask
	}
}

// WithVersion sets the version.
func WithVersion(version semver.Version) SourceStateOption {
	return func(s *SourceState) {
		s.version = version
	}
}

// WithWarnFunc sets the warning function.
func WithWarnFunc(warnFunc WarnFunc) SourceStateOption {
	return func(s *SourceState) {
		s.warnFunc = warnFunc
	}
}

// A TargetStateEntryFunc returns a TargetStateEntry based on reading an AbsPath
// on a System.
type TargetStateEntryFunc func(System, AbsPath) (TargetStateEntry, error)

// NewSourceState creates a new source state with the given options.
func NewSourceState(options ...SourceStateOption) *SourceState {
	s := &SourceState{
		removeDirs:           chezmoiset.New[RelPath](),
		umask:                Umask,
		encryption:           NoEncryption{},
		ignore:               NewPatternSet(),
		remove:               NewPatternSet(),
		httpClient:           http.DefaultClient,
		logger:               slog.Default(),
		readTemplateData:     true,
		readTemplates:        true,
		priorityTemplateData: make(map[string]any),
		userTemplateData:     make(map[string]any),
		templateOptions:      DefaultTemplateOptions,
		templates:            make(map[string]*Template),
		externals:            make(map[RelPath][]*External),
		ignoredRelPaths:      chezmoiset.New[RelPath](),
	}
	for _, option := range options {
		option(s)
	}
	return s
}

// A PreAddFunc is called before a new source state entry is added.
type PreAddFunc func(targetRelPath RelPath, fileInfo fs.FileInfo) error

// A ReplaceFunc is called before a source state entry is replaced.
type ReplaceFunc func(targetRelPath RelPath, newSourceStateEntry, oldSourceStateEntry SourceStateEntry) error

// AddOptions are options to SourceState.Add.
type AddOptions struct {
	AutoTemplate      bool                 // Automatically create templates, if possible.
	Create            bool                 // Add create_ entries instead of normal entries.
	Encrypt           bool                 // Encrypt files.
	EncryptedSuffix   string               // Suffix for encrypted files.
	Errorf            func(string, ...any) // Function to print errors.
	Exact             bool                 // Add the exact_ attribute to added directories.
	Filter            *EntryTypeFilter     // Entry type filter.
	OnIgnoreFunc      func(RelPath)        // Function to call when a target is ignored.
	PreAddFunc        PreAddFunc           // Function to be called before a source entry is added.
	ConfigFileAbsPath AbsPath              // Path to config file.
	ProtectedAbsPaths []AbsPath            // Paths that must not be added.
	RemoveDir         RelPath              // Directory to remove before adding.
	ReplaceFunc       ReplaceFunc          // Function to be called before a source entry is replaced.
	Template          bool                 // Add the .tmpl attribute to added files.
	TemplateSymlinks  bool                 // Add symlinks with targets in the source or home directories as templates.
}

// Add adds destAbsPathInfos to s.
func (s *SourceState) Add(
	sourceSystem System,
	persistentState PersistentState,
	destSystem System,
	destAbsPathInfos map[AbsPath]fs.FileInfo,
	options *AddOptions,
) error {
	// Filter out excluded and ignored paths.
	destAbsPaths := slices.Sorted(maps.Keys(destAbsPathInfos))
	n := 0
	for _, destAbsPath := range destAbsPaths {
		destAbsPathInfo := destAbsPathInfos[destAbsPath]
		if !options.Filter.IncludeFileInfo(destAbsPathInfo) {
			continue
		}

		targetRelPath := destAbsPath.MustTrimDirPrefix(s.destDirAbsPath)
		if s.Ignore(targetRelPath) {
			if options.OnIgnoreFunc != nil {
				options.OnIgnoreFunc(targetRelPath)
			}
			continue
		}

		destAbsPaths[n] = destAbsPath
		n++
	}
	destAbsPaths = destAbsPaths[:n]

	// Check for protected paths.
	for _, destAbsPath := range destAbsPaths {
		if destAbsPath == options.ConfigFileAbsPath {
			format := "%s: cannot add chezmoi's config file to chezmoi, use a config file template instead"
			return fmt.Errorf(format, destAbsPath)
		}
	}
	for _, destAbsPath := range destAbsPaths {
		for _, protectedAbsPath := range options.ProtectedAbsPaths {
			switch {
			case protectedAbsPath.IsEmpty():
				// Do nothing.
			case strings.HasPrefix(destAbsPath.String(), protectedAbsPath.String()):
				format := "%s: cannot add chezmoi file to chezmoi (%s is protected)"
				return fmt.Errorf(format, destAbsPath, protectedAbsPath)
			}
		}
	}

	type sourceUpdate struct {
		destAbsPath    AbsPath
		entryState     *EntryState
		sourceRelPaths []SourceRelPath
	}

	sourceUpdates := make([]sourceUpdate, 0, len(destAbsPaths))
	newSourceStateEntries := make(map[SourceRelPath]SourceStateEntry)
	newSourceStateEntriesByTargetRelPath := make(map[RelPath]SourceStateEntry)
	nonEmptyDirs := chezmoiset.New[SourceRelPath]()
	externalDirRelPaths := chezmoiset.New[RelPath]()
	dirRenames := make(map[AbsPath]AbsPath)
DEST_ABS_PATH:
	for _, destAbsPath := range destAbsPaths {
		targetRelPath := destAbsPath.MustTrimDirPrefix(s.destDirAbsPath)

		// Skip any entries in known external dirs.
		for externalDir := range externalDirRelPaths {
			if targetRelPath.HasDirPrefix(externalDir) {
				continue DEST_ABS_PATH
			}
		}

		// Find the target's parent directory in the source state.
		var parentSourceRelPath SourceRelPath
		if targetParentRelPath := targetRelPath.Dir(); targetParentRelPath == DotRelPath {
			parentSourceRelPath = SourceRelPath{}
		} else if parentEntry, ok := newSourceStateEntriesByTargetRelPath[targetParentRelPath]; ok {
			parentSourceRelPath = parentEntry.SourceRelPath()
		} else if nodes := s.root.GetNodes(targetParentRelPath); nodes != nil {
			for i, node := range nodes {
				if i == 0 {
					// nodes[0].sourceStateEntry should always be nil because it
					// refers to the destination directory, which is not managed.
					// chezmoi manages the destination directory's contents, not
					// the destination directory itself. For example, chezmoi
					// does not set the name or permissions of the user's home
					// directory.
					if node.SourceStateEntry != nil {
						panic(fmt.Errorf("nodes[0]: expected nil, got %+v", node.SourceStateEntry))
					}
					continue
				}
				switch sourceStateDir, ok := node.SourceStateEntry.(*SourceStateDir); {
				case i != len(nodes)-1 && !ok:
					panic(fmt.Errorf("nodes[%d]: unexpected non-terminal source state entry, got %T", i, node.SourceStateEntry))
				case ok && sourceStateDir.attr.External:
					targetRelPathComponents := targetRelPath.SplitAll()
					externalDirRelPath := EmptyRelPath.Join(targetRelPathComponents[:i]...)
					externalDirRelPaths.Add(externalDirRelPath)
					if options.Errorf != nil {
						options.Errorf("%s: skipping entries in external_ directory\n", externalDirRelPath)
					}
					continue DEST_ABS_PATH
				}
			}
			parentSourceRelPath = nodes[len(nodes)-1].SourceStateEntry.SourceRelPath()
		} else {
			return fmt.Errorf("%s: parent directory not in source state", destAbsPath)
		}
		nonEmptyDirs.Add(parentSourceRelPath)

		destAbsPathInfo := destAbsPathInfos[destAbsPath]
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

		if options.PreAddFunc != nil {
			switch err := options.PreAddFunc(targetRelPath, destAbsPathInfo); {
			case errors.Is(err, fs.SkipDir):
				continue DEST_ABS_PATH
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
			if !oldSourceEntryRelPath.IsEmpty() && oldSourceEntryRelPath != sourceEntryRelPath {
				if options.ReplaceFunc != nil {
					switch err := options.ReplaceFunc(targetRelPath, newSourceStateEntry, oldSourceStateEntry); {
					case errors.Is(err, fs.SkipDir):
						continue DEST_ABS_PATH
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
					continue DEST_ABS_PATH
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
		if nonEmptyDirs.Contains(sourceEntryRelPath) {
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
			origin: SourceStateOriginAbsPath(s.sourceDirAbsPath.Join(sourceEntryRelPath.RelPath())),
			targetStateEntry: &TargetStateFile{
				contentsFunc:       eagerNoErr[[]byte](nil),
				contentsSHA256Func: eagerNoErr(sha256.Sum256(nil)),
				empty:              true,
				perm:               0o666 &^ s.umask,
			},
		}
	}

	var sourceRoot SourceStateEntryTreeNode
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
				origin:        SourceStateOriginRemove{},
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
				sourceSystem,
				sourceSystem,
				NullPersistentState{},
				s.sourceDirAbsPath,
				sourceRelPath.RelPath(),
				ApplyOptions{
					Filter: options.Filter,
					Umask:  s.umask,
				},
			)
			if err != nil {
				return err
			}
		}
		if !sourceUpdate.destAbsPath.IsEmpty() {
			if err := PersistentStateSet(persistentState, EntryStateBucket, sourceUpdate.destAbsPath.Bytes(), sourceUpdate.entryState); err != nil {
				return err
			}
		}
	}

	// Rename directories last because updates assume that directory names have
	// not changed. Rename directories in reverse order so children are renamed
	// before their parents.
	oldDirAbsPaths := slices.Sorted(maps.Keys(dirRenames))
	for _, oldDirAbsPath := range slices.Backward(oldDirAbsPaths) {
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
	destAbsPathInfos map[AbsPath]fs.FileInfo,
	system System,
	destAbsPath AbsPath,
	fileInfo fs.FileInfo,
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
		if _, ok := s.root.Get(parentRelPath).(*SourceStateDir); ok {
			return nil
		}

		destAbsPath = parentAbsPath
		fileInfo = nil
	}
}

// A PreApplyFunc is called before a target is applied.
type PreApplyFunc func(targetRelPath RelPath, targetEntryState, lastWrittenEntryState, actualEntryState *EntryState) error

// ApplyOptions are options to SourceState.ApplyAll and SourceState.ApplyOne.
type ApplyOptions struct {
	Filter       *EntryTypeFilter
	PreApplyFunc PreApplyFunc
	Umask        fs.FileMode
}

// Apply updates targetRelPath in targetDirAbsPath in destSystem to match s.
func (s *SourceState) Apply(
	targetSystem, destSystem System,
	persistentState PersistentState,
	targetDirAbsPath AbsPath,
	targetRelPath RelPath,
	options ApplyOptions,
) error {
	sourceStateEntry := s.root.Get(targetRelPath)
	if sourceStateEntry == nil {
		return nil
	}

	if !options.Filter.IncludeSourceStateEntry(sourceStateEntry) {
		return nil
	}

	destAbsPath := s.destDirAbsPath.Join(targetRelPath)
	targetStateEntry, err := sourceStateEntry.TargetStateEntry(destSystem, destAbsPath)
	if err != nil {
		return err
	}

	if !options.Filter.IncludeTargetStateEntry(targetStateEntry) {
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
		ok, err := PersistentStateGet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), &entryState)
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
			err := PersistentStateSet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), targetEntryState)
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

	return PersistentStateSet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), targetEntryState)
}

// Encryption returns s's encryption.
func (s *SourceState) Encryption() Encryption {
	return s.encryption
}

// ExecuteTemplateDataOptions are options to SourceState.ExecuteTemplateData.
type ExecuteTemplateDataOptions struct {
	Name            string
	Destination     string
	Data            []byte
	TemplateOptions TemplateOptions
	ExtraData       map[string]any
}

// ExecuteTemplateData returns the result of executing template data.
func (s *SourceState) ExecuteTemplateData(options ExecuteTemplateDataOptions) ([]byte, error) {
	templateOptions := options.TemplateOptions
	templateOptions.Funcs = s.templateFuncs
	templateOptions.Options = slices.Clone(s.templateOptions)

	tmpl, err := ParseTemplate(options.Name, options.Data, templateOptions)
	if err != nil {
		return nil, err
	}

	for _, t := range s.templates {
		tmpl, err = tmpl.AddParseTree(t)
		if err != nil {
			return nil, err
		}
	}

	// Set .chezmoi.sourceFile to the name of the template.
	templateData := s.TemplateData()
	if chezmoiTemplateData, ok := templateData["chezmoi"].(map[string]any); ok {
		chezmoiTemplateData["sourceFile"] = options.Name
		chezmoiTemplateData["targetFile"] = options.Destination
	}
	RecursiveMerge(templateData, options.ExtraData)

	result, err := tmpl.Execute(templateData)
	if errors.Is(err, errReturnEmpty) {
		return nil, nil
	}
	return result, err
}

// ForEach calls f for each source state entry.
func (s *SourceState) ForEach(f func(RelPath, SourceStateEntry) error) error {
	return s.root.ForEach(EmptyRelPath, func(targetRelPath RelPath, entry SourceStateEntry) error {
		return f(targetRelPath, entry)
	})
}

// Get returns the source state entry for targetRelPath.
func (s *SourceState) Get(targetRelPath RelPath) SourceStateEntry {
	return s.root.Get(targetRelPath)
}

// Ignore returns if targetRelPath should be ignored.
func (s *SourceState) Ignore(targetRelPath RelPath) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	ignore := s.ignore.Match(targetRelPath.String()) == PatternSetMatchInclude
	if ignore {
		s.ignoredRelPaths.Add(targetRelPath)
	}
	return ignore
}

// Ignored returns all ignored RelPaths.
func (s *SourceState) Ignored() []RelPath {
	relPaths := s.ignoredRelPaths.Elements()
	slices.SortFunc(relPaths, CompareRelPaths)
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
func (s *SourceState) PostApply(
	targetSystem System,
	persistentState PersistentState,
	targetDirAbsPath AbsPath,
	targetRelPaths []RelPath,
) error {
	// Remove empty directories with the remove_ attribute. This assumes that
	// targetRelPaths is already sorted and iterates in reverse order so that
	// children are removed before their parents.
TARGET:
	for i := len(targetRelPaths) - 1; i >= 0; i-- {
		targetRelPath := targetRelPaths[i]
		if !s.removeDirs.Contains(targetRelPath) {
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
			continue TARGET
		case errors.Is(err, fs.ErrNotExist):
			continue TARGET
		default:
			return err
		}
		entryState := &EntryState{
			Type: EntryStateTypeRemove,
		}
		if err := PersistentStateSet(persistentState, EntryStateBucket, targetAbsPath.Bytes(), entryState); err != nil {
			return err
		}
	}

	return nil
}

// ReadOptions are options to SourceState.Read.
type ReadOptions struct {
	ReadHTTPResponse func(string, *http.Response) ([]byte, error)
	RefreshExternals RefreshExternals
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
		defer allSourceStateEntriesMu.Unlock()
		allSourceStateEntries[relPath] = append(allSourceStateEntries[relPath], sourceStateEntries...)
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
		case fileInfo.Name() == dataName:
			if !s.readTemplateData {
				return nil
			}
			if err := s.addTemplateDataDir(sourceAbsPath, fileInfo); err != nil {
				return err
			}
			return fs.SkipDir
		case isPrefixDotFormat(fileInfo.Name(), dataName):
			if !s.readTemplateData {
				return nil
			}
			return s.addTemplateData(sourceAbsPath)
		case fileInfo.Name() == TemplatesDirName:
			if s.readTemplates {
				if err := s.addTemplatesDir(ctx, sourceAbsPath); err != nil {
					return err
				}
			}
			return fs.SkipDir
		case s.templateDataOnly:
			return nil
		case isPrefixDotFormat(fileInfo.Name(), externalName) || isPrefixDotFormatDotTmpl(fileInfo.Name(), externalName):
			parentAbsPath, _ := sourceAbsPath.Split()
			return s.addExternal(sourceAbsPath, parentAbsPath)
		case fileInfo.Name() == externalsDirName:
			if err := s.addExternalDir(ctx, sourceAbsPath); err != nil {
				return err
			}
			return fs.SkipDir
		case fileInfo.Name() == ignoreName || fileInfo.Name() == ignoreName+TemplateSuffix:
			return s.addPatterns(s.ignore, sourceAbsPath, parentSourceRelPath)
		case fileInfo.Name() == removeName || fileInfo.Name() == removeName+TemplateSuffix:
			return s.addPatterns(s.remove, sourceAbsPath, parentSourceRelPath)
		case fileInfo.Name() == scriptsDirName:
			scriptsDirSourceStateEntries, err := s.readScriptsDir(ctx, sourceAbsPath)
			if err != nil {
				return err
			}
			for relPath, scriptSourceStateEntries := range scriptsDirSourceStateEntries {
				addSourceStateEntries(relPath, scriptSourceStateEntries...)
			}
			return fs.SkipDir
		case fileInfo.Name() == VersionName:
			return s.readVersionFile(sourceAbsPath)
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			fallthrough
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		case fileInfo.IsDir():
			da := parseDirAttr(sourceName.String())
			targetRelPath := parentSourceRelPath.Dir().TargetRelPath(s.encryption.EncryptedSuffix()).JoinString(da.TargetName)
			if s.Ignore(targetRelPath) {
				return fs.SkipDir
			}
			sourceStateDir := s.newSourceStateDir(sourceAbsPath, sourceRelPath, da)
			addSourceStateEntries(targetRelPath, sourceStateDir)
			if da.External {
				sourceStateEntries, err := s.readExternalDir(sourceAbsPath, sourceRelPath, targetRelPath)
				if err != nil {
					return err
				}
				allSourceStateEntriesMu.Lock()
				for relPath, entries := range sourceStateEntries {
					allSourceStateEntries[relPath] = append(allSourceStateEntries[relPath], entries...)
				}
				allSourceStateEntriesMu.Unlock()
				return fs.SkipDir
			}
			if sourceStateDir.attr.Remove {
				s.mutex.Lock()
				s.removeDirs.Add(targetRelPath)
				s.mutex.Unlock()
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
			return &UnsupportedFileTypeError{
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
	externalRelPaths := make([]RelPath, 0, len(s.externals))
	for externalRelPath := range s.externals {
		externalRelPaths = append(externalRelPaths, externalRelPath)
	}
	slices.SortFunc(externalRelPaths, CompareRelPaths)
	for _, externalRelPath := range externalRelPaths {
		if s.Ignore(externalRelPath) {
			continue
		}
		for _, external := range s.externals[externalRelPath] {
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
	}

	// Remove all ignored targets.
	for targetRelPath := range allSourceStateEntries {
		if s.Ignore(targetRelPath) {
			delete(allSourceStateEntries, targetRelPath)
		}
	}

	// Generate SourceStateRemoves for existing targets.
	matches, err := s.remove.Glob(s.system.UnderlyingFS(), ensureSuffix(s.destDirAbsPath.String(), "/"))
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

	// Where there are multiple SourceStateEntries for a single target, replace
	// them with a single canonical SourceStateEntry if possible.
	for targetRelPath, sourceStateEntries := range allSourceStateEntries {
		if sourceStateEntry, ok := canonicalSourceStateEntry(sourceStateEntries); ok {
			allSourceStateEntries[targetRelPath] = []SourceStateEntry{sourceStateEntry}
		}
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
		case !sourceStateDir.attr.Exact:
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
	var gitRepoExternalRelPaths []RelPath
	for externalRelPath, externals := range s.externals {
		if s.Ignore(externalRelPath) {
			continue
		}
		for _, external := range externals {
			if external.Type == ExternalTypeGitRepo {
				gitRepoExternalRelPaths = append(gitRepoExternalRelPaths, externalRelPath)
			}
		}
	}
	slices.SortFunc(gitRepoExternalRelPaths, CompareRelPaths)
	for _, externalRelPath := range gitRepoExternalRelPaths {
		for _, external := range s.externals[externalRelPath] {
			destAbsPath := s.destDirAbsPath.Join(externalRelPath)
			switch _, err := s.system.Lstat(destAbsPath); {
			case errors.Is(err, fs.ErrNotExist):
				// FIXME add support for using builtin git
				sourceStateCommand := &SourceStateCommand{
					// Use a sync.OnceValue to defer the call to
					// os/exec.Command because os/exec.Command calls
					// os/exec.LookupPath and therefore depends on the state of
					// $PATH when os/exec.Command is called, not the state of
					// $PATH when os/exec.Cmd.{Run,Start} is called.
					cmdFunc: sync.OnceValue(func() *exec.Cmd {
						args := []string{"clone"}
						args = append(args, external.Clone.Args...)
						args = append(args, external.URL, destAbsPath.String())
						cmd := exec.Command("git", args...)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						return cmd
					}),
					origin:        external,
					forceRefresh:  options.RefreshExternals == RefreshExternalsAlways,
					refreshPeriod: external.RefreshPeriod,
					sourceAttr: SourceAttr{
						External: true,
					},
				}
				allSourceStateEntries[externalRelPath] = append(allSourceStateEntries[externalRelPath], sourceStateCommand)
			case err != nil:
				return err
			default:
				// FIXME add support for using builtin git
				sourceStateCommand := &SourceStateCommand{
					// Use a sync.OnceValue to defer the call to
					// os/exec.Command because os/exec.Command calls
					// os/exec.LookupPath and therefore depends on the state of
					// $PATH when os/exec.Command is called, not the state of
					// $PATH when os/exec.Cmd.{Run,Start} is called.
					cmdFunc: sync.OnceValue(func() *exec.Cmd {
						args := []string{"pull"}
						args = append(args, external.Pull.Args...)
						cmd := exec.Command("git", args...)
						cmd.Dir = destAbsPath.String()
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						return cmd
					}),
					origin:        external,
					forceRefresh:  options.RefreshExternals == RefreshExternalsAlways,
					refreshPeriod: external.RefreshPeriod,
					sourceAttr: SourceAttr{
						External: true,
					},
				}
				allSourceStateEntries[externalRelPath] = append(allSourceStateEntries[externalRelPath], sourceStateCommand)
			}
		}
	}

	// Check for inconsistent source entries. Iterate over the target names in
	// order so that any error is deterministic.
	targetRelPaths := make([]RelPath, 0, len(allSourceStateEntries))
	for targetRelPath := range allSourceStateEntries {
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	slices.SortFunc(targetRelPaths, CompareRelPaths)
	errs := make([]error, 0, len(targetRelPaths))
	for _, targetRelPath := range targetRelPaths {
		sourceStateEntries := allSourceStateEntries[targetRelPath]
		if len(sourceStateEntries) == 1 {
			continue
		}

		origins := make([]string, len(sourceStateEntries))
		for i, sourceStateEntry := range sourceStateEntries {
			origins[i] = sourceStateEntry.Origin().OriginString()
		}
		slices.Sort(origins)
		errs = append(errs, &InconsistentStateError{
			targetRelPath: targetRelPath,
			origins:       origins,
		})
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	// Populate s.Entries with the unique source entry for each target.
	for targetRelPath, sourceEntries := range allSourceStateEntries {
		s.root.Set(targetRelPath, sourceEntries[0])
	}

	return nil
}

// TargetRelPaths returns all of s's target relative paths in order.
func (s *SourceState) TargetRelPaths() []RelPath {
	entries := s.root.GetMap()
	targetRelPaths := make([]RelPath, 0, len(entries))
	for targetRelPath := range entries {
		targetRelPaths = append(targetRelPaths, targetRelPath)
	}
	slices.SortFunc(targetRelPaths, func(a, b RelPath) int {
		if compare := cmp.Compare(entries[a].Order(), entries[b].Order()); compare != 0 {
			return compare
		}
		return CompareRelPaths(a, b)
	})
	return targetRelPaths
}

// TemplateData returns a copy of s's template data.
func (s *SourceState) TemplateData() map[string]any {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.templateData == nil {
		s.templateData = make(map[string]any)
		if s.defaultTemplateDataFunc != nil {
			s.defaultTemplateData = s.defaultTemplateDataFunc()
			s.defaultTemplateDataFunc = nil
		}
		RecursiveMerge(s.templateData, s.defaultTemplateData)
		RecursiveMerge(s.templateData, s.userTemplateData)
		RecursiveMerge(s.templateData, s.priorityTemplateData)
	}
	templateData, err := copystructure.Copy(s.templateData)
	if err != nil {
		panic(err)
	}
	return templateData.(map[string]any) //nolint:forcetypeassert,revive
}

// addExternal adds external source entries to s.
func (s *SourceState) addExternal(sourceAbsPath, parentAbsPath AbsPath) error {
	parentRelPath, err := parentAbsPath.TrimDirPrefix(s.sourceDirAbsPath)
	if err != nil {
		return err
	}
	parentSourceRelPath := NewSourceRelDirPath(parentRelPath.String())
	parentTargetSourceRelPath := parentSourceRelPath.TargetRelPath(s.encryption.EncryptedSuffix())

	format, err := FormatFromAbsPath(sourceAbsPath.TrimSuffix(TemplateSuffix))
	if err != nil {
		return err
	}
	data, err := s.executeTemplate(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	externals := make(map[string]External)
	if err := format.Unmarshal(data, &externals); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for path, external := range externals {
		switch {
		case path == "":
			return fmt.Errorf("%s: empty path", sourceAbsPath)
		case strings.HasPrefix(path, "/") || filepath.IsAbs(path):
			return fmt.Errorf("%s: %s: path is not relative", sourceAbsPath, path)
		}
		switch relPath, err := filepath.Rel(".", path); {
		case err != nil:
			return fmt.Errorf("%s: %s: %w", sourceAbsPath, path, err)
		case relPath == ".":
			return fmt.Errorf("%s: %s: empty relative path", sourceAbsPath, path)
		case relPath == "..", strings.HasPrefix(relPath, "../"):
			return fmt.Errorf("%s: %s: relative path in parent", sourceAbsPath, path)
		}
		targetRelPath := parentTargetSourceRelPath.JoinString(path)
		external.sourceAbsPath = sourceAbsPath
		s.externals[targetRelPath] = append(s.externals[targetRelPath], &external)
	}
	return nil
}

// addExternalDir adds all externals in externalsDirAbsPath to s.
func (s *SourceState) addExternalDir(ctx context.Context, externalsDirAbsPath AbsPath) error {
	walkFunc := func(ctx context.Context, externalAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
		if externalAbsPath == externalsDirAbsPath {
			return nil
		}
		if err == nil && fileInfo.Mode().Type() == fs.ModeSymlink {
			fileInfo, err = s.system.Stat(externalAbsPath)
		}
		switch {
		case err != nil:
			return err
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			return fmt.Errorf("%s: not allowed in %s directory", externalAbsPath, externalsDirName)
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		case fileInfo.Mode().IsRegular():
			parentAbsPath, _ := externalAbsPath.Split()
			return s.addExternal(externalAbsPath, parentAbsPath.Dir())
		case fileInfo.IsDir():
			return nil
		default:
			return &UnsupportedFileTypeError{
				absPath: externalAbsPath,
				mode:    fileInfo.Mode(),
			}
		}
	}
	return concurrentWalkSourceDir(ctx, s.system, externalsDirAbsPath, walkFunc)
}

// addPatterns executes the template at sourceAbsPath, interprets the result as
// a list of patterns, and adds all patterns found to patternSet.
func (s *SourceState) addPatterns(patternSet *PatternSet, sourceAbsPath AbsPath, sourceRelPath SourceRelPath) error {
	data, err := s.executeTemplate(sourceAbsPath)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	dir := sourceRelPath.Dir().TargetRelPath("")
	lineNumber := 0
	for line := range bytes.Lines(data) {
		lineNumber++
		line = commentRx.ReplaceAll(line, nil)
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		include := PatternSetInclude
		line, ok := bytes.CutPrefix(line, []byte{'!'})
		if ok {
			include = PatternSetExclude
		}
		pattern := dir.JoinString(string(line)).String()
		if err := patternSet.Add(pattern, include); err != nil {
			return fmt.Errorf("%s:%d: %w", sourceAbsPath, lineNumber, err)
		}
	}
	return nil
}

// addTemplateData adds all template data in sourceAbsPath to s.
func (s *SourceState) addTemplateData(sourceAbsPath AbsPath) error {
	format, err := FormatFromAbsPath(sourceAbsPath)
	if err != nil {
		return err
	}
	data, err := s.system.ReadFile(sourceAbsPath)
	if err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	var templateData map[string]any
	if err := format.Unmarshal(data, &templateData); err != nil {
		return fmt.Errorf("%s: %w", sourceAbsPath, err)
	}
	s.mutex.Lock()
	RecursiveMerge(s.userTemplateData, templateData)
	// Clear the cached template data, as the change to the user template data
	// means that the cached value is now invalid.
	s.templateData = nil
	s.mutex.Unlock()
	return nil
}

// addTemplateDataDir adds all template data in the directory sourceAbsPath to
// s.
func (s *SourceState) addTemplateDataDir(sourceAbsPath AbsPath, fileInfo fs.FileInfo) error {
	walkFunc := func(dataAbsPath AbsPath, fileInfo fs.FileInfo, err error) error {
		if dataAbsPath == sourceAbsPath {
			return nil
		}
		if err == nil && fileInfo.Mode().Type() == fs.ModeSymlink {
			fileInfo, err = s.system.Stat(dataAbsPath)
		}
		switch {
		case err != nil:
			return err
		case strings.HasPrefix(fileInfo.Name(), Prefix):
			return fmt.Errorf("%s: not allowed in %s directory", dataAbsPath, dataName)
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		case fileInfo.Mode().IsRegular():
			return s.addTemplateData(dataAbsPath)
		case fileInfo.IsDir():
			return nil
		default:
			return &UnsupportedFileTypeError{
				absPath: dataAbsPath,
				mode:    fileInfo.Mode(),
			}
		}
	}
	return walkSourceDirHelper(s.system, sourceAbsPath, fileInfo, walkFunc)
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
			return fmt.Errorf("%s: not allowed in %s directory", templateAbsPath, TemplatesDirName)
		case strings.HasPrefix(fileInfo.Name(), ignorePrefix):
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		case fileInfo.Mode().IsRegular():
			contents, err := s.system.ReadFile(templateAbsPath)
			if err != nil {
				return err
			}
			templateRelPath := templateAbsPath.MustTrimDirPrefix(templatesDirAbsPath)
			name := templateRelPath.String()

			tmpl, err := ParseTemplate(name, contents, TemplateOptions{
				Funcs:   s.templateFuncs,
				Options: slices.Clone(s.templateOptions),
			})
			if err != nil {
				return err
			}
			s.mutex.Lock()
			s.templates[name] = tmpl
			s.mutex.Unlock()
			return nil
		case fileInfo.IsDir():
			return nil
		default:
			return &UnsupportedFileTypeError{
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
	return s.ExecuteTemplateData(ExecuteTemplateDataOptions{
		Name: templateAbsPath.String(),
		Data: data,
	})
}

// getExternalDataRaw returns the raw data for external at externalRelPath,
// possibly from the external cache.
func (s *SourceState) getExternalDataRaw(
	ctx context.Context,
	externalRelPath RelPath,
	urlStr string,
	refreshPeriod Duration,
	options *ReadOptions,
) ([]byte, error) {
	// Handle file:// URLs by always reading from disk.
	switch urlStruct, err := url.Parse(urlStr); {
	case err != nil:
		return nil, err
	case urlStruct.Scheme == "file":
		data, err := s.system.ReadFile(NewAbsPath(urlStruct.Path))
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	var now time.Time
	if options != nil && options.TimeNow != nil {
		now = options.TimeNow()
	} else {
		now = time.Now()
	}
	now = now.UTC()

	refreshExternals := RefreshExternalsAuto
	if options != nil {
		refreshExternals = options.RefreshExternals
	}
	urlSHA256 := sha256.Sum256([]byte(urlStr))
	cacheKey := hex.EncodeToString(urlSHA256[:])
	cachedDataAbsPath := s.cacheDirAbsPath.JoinString("external", cacheKey)
	switch refreshExternals {
	case RefreshExternalsAlways:
		// Never use the cache.
	case RefreshExternalsAuto:
		// Use the cache, if available and within the refresh period.
		if fileInfo, err := s.baseSystem.Stat(cachedDataAbsPath); err == nil {
			if refreshPeriod == 0 || fileInfo.ModTime().Add(time.Duration(refreshPeriod)).After(now) {
				if data, err := s.baseSystem.ReadFile(cachedDataAbsPath); err == nil {
					return data, nil
				}
			}
		}
	case RefreshExternalsNever:
		// Always use the cache, if available, irrespective of the refresh
		// period.
		if data, err := s.baseSystem.ReadFile(cachedDataAbsPath); err == nil {
			return data, nil
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := chezmoilog.LogHTTPRequest(ctx, s.logger, s.httpClient, req)
	if err != nil {
		return nil, err
	}
	var data []byte
	if options == nil || options.ReadHTTPResponse == nil {
		data, err = io.ReadAll(resp.Body)
	} else {
		data, err = options.ReadHTTPResponse(urlStr, resp)
	}
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
		return nil, fmt.Errorf("%s: %s: %s", externalRelPath, urlStr, resp.Status)
	}

	if err := MkdirAll(s.baseSystem, cachedDataAbsPath.Dir(), 0o700); err != nil {
		return nil, err
	}
	if err := s.baseSystem.WriteFile(cachedDataAbsPath, data, 0o600); err != nil {
		return nil, err
	}
	if err := s.baseSystem.Chtimes(cachedDataAbsPath, now, now); err != nil {
		return nil, err
	}

	return data, nil
}

// getExternalDataAndURL iterates over external.URL and external.URLs, returning
// the first data that is downloaded successfully and the URL it was downloaded
// from.
func (s *SourceState) getExternalDataAndURL(
	ctx context.Context,
	externalRelPath RelPath,
	external *External,
	options *ReadOptions,
) ([]byte, string, error) {
	var firstURLStr string
	var firstErr error
	for _, urlStr := range append([]string{external.URL}, external.URLs...) {
		if urlStr == "" {
			continue
		}
		data, err := s.getExternalDataRaw(ctx, externalRelPath, urlStr, external.RefreshPeriod, options)
		if err == nil {
			return data, urlStr, nil
		}
		if firstURLStr == "" {
			firstURLStr = urlStr
			firstErr = err
		}
	}
	if firstURLStr == "" {
		return nil, "", fmt.Errorf("%s: no URL", externalRelPath)
	}
	return nil, firstURLStr, firstErr
}

// getExternalData reads the external data for externalRelPath from
// external.URL or external.URLs, returning the data and URL.
func (s *SourceState) getExternalData(
	ctx context.Context,
	externalRelPath RelPath,
	external *External,
	options *ReadOptions,
) ([]byte, string, error) {
	data, urlStr, err := s.getExternalDataAndURL(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, "", err
	}

	var errs []error

	if external.Checksum.Size != 0 && external.Checksum.SHA256 == nil && external.Checksum.SHA384 == nil &&
		external.Checksum.SHA512 == nil {
		s.warnFunc("%s: warning: insecure size check without secure hash will be removed\n", externalRelPath)
		if len(data) != external.Checksum.Size {
			err := fmt.Errorf("size mismatch: expected %d, got %d", external.Checksum.Size, len(data))
			errs = append(errs, err)
		}
	}

	if external.Checksum.MD5 != nil {
		s.warnFunc(
			"%s: warning: insecure MD5 checksum will be removed, use a secure hash like SHA256 instead\n",
			externalRelPath,
		)
		if gotMD5Sum := md5Sum(data); !bytes.Equal(gotMD5Sum, external.Checksum.MD5) {
			err := fmt.Errorf("MD5 mismatch: expected %s, got %s", external.Checksum.MD5, hex.EncodeToString(gotMD5Sum))
			errs = append(errs, err)
		}
	}

	if external.Checksum.RIPEMD160 != nil {
		s.warnFunc(
			"%s: warning: insecure RIPEMD-160 checksum will be removed, use a secure hash like SHA256 instead\n",
			externalRelPath,
		)
		if gotRIPEMD160Sum := ripemd160Sum(data); !bytes.Equal(gotRIPEMD160Sum, external.Checksum.RIPEMD160) {
			format := "RIPEMD-160 mismatch: expected %s, got %s"
			err := fmt.Errorf(format, external.Checksum.RIPEMD160, hex.EncodeToString(gotRIPEMD160Sum))
			errs = append(errs, err)
		}
	}

	if external.Checksum.SHA1 != nil {
		s.warnFunc(
			"%s: warning: insecure SHA1 checksum will be removed, use a secure hash like SHA256 instead\n",
			externalRelPath,
		)
		if gotSHA1Sum := sha1Sum(data); !bytes.Equal(gotSHA1Sum, external.Checksum.SHA1) {
			err := fmt.Errorf("SHA1 mismatch: expected %s, got %s", external.Checksum.SHA1, hex.EncodeToString(gotSHA1Sum))
			errs = append(errs, err)
		}
	}

	if external.Checksum.SHA256 != nil {
		if gotSHA256Sum := sha256.Sum256(data); !bytes.Equal(gotSHA256Sum[:], external.Checksum.SHA256) {
			format := "SHA256 mismatch: expected %s, got %s"
			err := fmt.Errorf(format, external.Checksum.SHA256, hex.EncodeToString(gotSHA256Sum[:]))
			errs = append(errs, err)
		}
	}

	if external.Checksum.SHA384 != nil {
		if gotSHA384Sum := sha384Sum(data); !bytes.Equal(gotSHA384Sum, external.Checksum.SHA384) {
			errs = append(errs, fmt.Errorf("SHA384 mismatch: expected %s, got %s",
				external.Checksum.SHA384, hex.EncodeToString(gotSHA384Sum)))
		}
	}

	if external.Checksum.SHA512 != nil {
		if gotSHA512Sum := sha512Sum(data); !bytes.Equal(gotSHA512Sum, external.Checksum.SHA512) {
			errs = append(errs, fmt.Errorf("SHA512 mismatch: expected %s, got %s",
				external.Checksum.SHA512, hex.EncodeToString(gotSHA512Sum)))
		}
	}

	if len(errs) != 0 {
		return nil, urlStr, fmt.Errorf("%s: %w", externalRelPath, errors.Join(errs...))
	}

	if external.Encrypted {
		data, err = s.encryption.Decrypt(data)
		if err != nil {
			return nil, urlStr, fmt.Errorf("%s: %s: %w", externalRelPath, urlStr, err)
		}
	}

	data, err = decompress(external.Decompress, data)
	if err != nil {
		return nil, urlStr, fmt.Errorf("%s: %w", externalRelPath, err)
	}

	if external.Filter.Command != "" {
		cmd := exec.Command(external.Filter.Command, external.Filter.Args...)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Stderr = os.Stderr
		data, err = chezmoilog.LogCmdOutput(s.logger, cmd)
		if err != nil {
			return nil, urlStr, fmt.Errorf("%s: %s: %w", externalRelPath, urlStr, err)
		}
	}

	return data, urlStr, nil
}

// newSourceStateDir returns a new SourceStateDir.
func (s *SourceState) newSourceStateDir(absPath AbsPath, sourceRelPath SourceRelPath, dirAttr DirAttr) *SourceStateDir {
	targetStateDir := &TargetStateDir{
		perm: dirAttr.perm() &^ s.umask,
	}
	return &SourceStateDir{
		origin:           SourceStateOriginAbsPath(absPath),
		sourceRelPath:    sourceRelPath,
		attr:             dirAttr,
		targetStateEntry: targetStateDir,
	}
}

// newCreateTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with the value of sourceContentsFunc if the file does not already exist,
// or returns the actual file's contents unchanged if the file already exists.
func (s *SourceState) newCreateTargetStateEntryFunc(
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	sourceContentsFunc func() ([]byte, error),
) TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		var contentsFunc func() ([]byte, error)
		switch contents, err := destSystem.ReadFile(destAbsPath); {
		case err == nil:
			contentsFunc = eagerNoErr(contents)
		case errors.Is(err, fs.ErrNotExist):
			contentsFunc = sync.OnceValues(func() ([]byte, error) {
				contents, err = sourceContentsFunc()
				if err != nil {
					return nil, err
				}
				if fileAttr.Template {
					contents, err = s.ExecuteTemplateData(ExecuteTemplateDataOptions{
						Name:        sourceRelPath.String(),
						Data:        contents,
						Destination: destAbsPath.String(),
					})
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
			contentsFunc:       contentsFunc,
			contentsSHA256Func: lazySHA256(contentsFunc),
			empty:              fileAttr.Empty,
			perm:               fileAttr.perm() &^ s.umask,
			sourceAttr: SourceAttr{
				Encrypted: fileAttr.Encrypted,
				Template:  fileAttr.Template,
			},
		}, nil
	}
}

// newFileTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with the contents of the value of sourceContentsFunc.
func (s *SourceState) newFileTargetStateEntryFunc(
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	sourceContentsFunc func() ([]byte, error),
) TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		if s.mode == ModeSymlink && !fileAttr.Encrypted && !fileAttr.Executable && !fileAttr.Private && !fileAttr.Template {
			switch contents, err := sourceContentsFunc(); {
			case err != nil:
				return nil, err
			case isEmpty(contents) && !fileAttr.Empty:
				return &TargetStateRemove{}, nil
			default:
				linkname := normalizeLinkname(s.sourceDirAbsPath.Join(sourceRelPath.RelPath()).String())
				return &TargetStateSymlink{
					linknameFunc: eagerNoErr(linkname),
					sourceAttr: SourceAttr{
						Template: fileAttr.Template,
					},
				}, nil
			}
		}
		executedContentsFunc := sync.OnceValues(func() ([]byte, error) {
			contents, err := sourceContentsFunc()
			if err != nil {
				return nil, err
			}
			if fileAttr.Template {
				contents, err = s.ExecuteTemplateData(ExecuteTemplateDataOptions{
					Name:        sourceRelPath.String(),
					Data:        contents,
					Destination: destAbsPath.String(),
				})
				if err != nil {
					return nil, err
				}
			}
			return contents, nil
		})
		return &TargetStateFile{
			contentsFunc:       executedContentsFunc,
			contentsSHA256Func: lazySHA256(executedContentsFunc),
			empty:              fileAttr.Empty,
			perm:               fileAttr.perm() &^ s.umask,
			sourceAttr: SourceAttr{
				Encrypted: fileAttr.Encrypted,
				Template:  fileAttr.Template,
			},
		}, nil
	}
}

// newModifyTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// file with the contents modified by running the sourceLazyContents script.
func (s *SourceState) newModifyTargetStateEntryFunc(
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	contentsFunc func() ([]byte, error),
	interpreter *Interpreter,
) TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := sync.OnceValues(func() (contents []byte, err error) {
			// FIXME this should share code with RealSystem.RunScript

			// Read the current contents of the target.
			var currentContents []byte
			currentContents, err = destSystem.ReadFile(destAbsPath)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			}

			// Compute the contents of the modifier.
			var modifierContents []byte
			modifierContents, err = contentsFunc()
			if err != nil {
				return nil, err
			}
			if fileAttr.Template {
				modifierContents, err = s.ExecuteTemplateData(ExecuteTemplateDataOptions{
					Name:        sourceRelPath.String(),
					Data:        modifierContents,
					Destination: destAbsPath.String(),
				})
				if err != nil {
					return nil, err
				}
			}

			// If the modifier is empty then return the current contents unchanged.
			if isEmpty(modifierContents) {
				return currentContents, nil
			}

			// If the modifier contains chezmoi:modify-template then execute it
			// as a template.
			if matches := modifyTemplateRx.FindAllSubmatchIndex(modifierContents, -1); matches != nil {
				sourceFile := sourceRelPath.String()
				templateContents := removeMatches(modifierContents, matches)
				var tmpl *Template
				tmpl, err = ParseTemplate(sourceFile, templateContents, TemplateOptions{
					Funcs:   s.templateFuncs,
					Options: slices.Clone(s.templateOptions),
				})
				if err != nil {
					return nil, err
				}

				// Temporarily set .chezmoi.stdin to the current contents and
				// .chezmoi.sourceFile to the name of the template.
				templateData := s.TemplateData()
				if chezmoiTemplateData, ok := templateData["chezmoi"].(map[string]any); ok {
					chezmoiTemplateData["stdin"] = string(currentContents)
					chezmoiTemplateData["sourceFile"] = sourceFile
				}

				return tmpl.Execute(templateData)
			}

			// Create the script temporary directory, if needed.
			s.createScriptTempDirOnce.Do(func() {
				if !s.scriptTempDirAbsPath.IsEmpty() {
					err = os.MkdirAll(s.scriptTempDirAbsPath.String(), 0o700)
				}
			})
			if err != nil {
				return nil, err
			}

			// Write the modifier to a temporary file.
			var tempFile *os.File
			if tempFile, err = os.CreateTemp(s.scriptTempDirAbsPath.String(), "*."+fileAttr.TargetName); err != nil {
				return nil, err
			}
			defer chezmoierrors.CombineFunc(&err, func() error {
				return os.RemoveAll(tempFile.Name())
			})
			if runtime.GOOS != "windows" {
				if err := tempFile.Chmod(0o700); err != nil {
					return nil, err
				}
			}
			_, err = tempFile.Write(modifierContents)
			err = chezmoierrors.Combine(err, tempFile.Close())
			if err != nil {
				return nil, err
			}

			// Run the modifier on the current contents.
			cmd := interpreter.ExecCommand(tempFile.Name())
			cmd.Env = append(os.Environ(),
				"CHEZMOI_SOURCE_FILE="+sourceRelPath.String(),
			)
			cmd.Stdin = bytes.NewReader(currentContents)
			cmd.Stderr = os.Stderr
			return chezmoilog.LogCmdOutput(s.logger, cmd)
		})
		return &TargetStateFile{
			contentsFunc:       contentsFunc,
			contentsSHA256Func: lazySHA256(contentsFunc),
			overwrite:          true,
			perm:               fileAttr.perm() &^ s.umask,
		}, nil
	}
}

// newRemoveTargetStateEntryFunc returns a targetStateEntryFunc that removes a
// target.
func (s *SourceState) newRemoveTargetStateEntryFunc() TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		return &TargetStateRemove{}, nil
	}
}

// newScriptTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// script with sourceLazyContents.
func (s *SourceState) newScriptTargetStateEntryFunc(
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	targetRelPath RelPath,
	sourceContentsFunc func() ([]byte, error),
	interpreter *Interpreter,
) TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		contentsFunc := sync.OnceValues(func() ([]byte, error) {
			contents, err := sourceContentsFunc()
			if err != nil {
				return nil, err
			}
			if fileAttr.Template {
				contents, err = s.ExecuteTemplateData(ExecuteTemplateDataOptions{
					Name:        sourceRelPath.String(),
					Data:        contents,
					Destination: destAbsPath.String(),
				})
				if err != nil {
					return nil, err
				}
			}
			return contents, nil
		})
		return &TargetStateScript{
			name:               targetRelPath,
			contentsFunc:       contentsFunc,
			contentsSHA256Func: lazySHA256(contentsFunc),
			condition:          fileAttr.Condition,
			interpreter:        interpreter,
			sourceAttr: SourceAttr{
				Condition: fileAttr.Condition,
			},
			sourceRelPath: sourceRelPath,
		}, nil
	}
}

// newSymlinkTargetStateEntryFunc returns a targetStateEntryFunc that returns a
// symlink with the linkname sourceLazyContents.
func (s *SourceState) newSymlinkTargetStateEntryFunc(
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	contentsFunc func() ([]byte, error),
) TargetStateEntryFunc {
	return func(destSystem System, destAbsPath AbsPath) (TargetStateEntry, error) {
		linknameFunc := func() (string, error) {
			linknameBytes, err := contentsFunc()
			if err != nil {
				return "", err
			}
			if fileAttr.Template {
				linknameBytes, err = s.ExecuteTemplateData(ExecuteTemplateDataOptions{
					Name:        sourceRelPath.String(),
					Data:        linknameBytes,
					Destination: destAbsPath.String(),
				})
				if err != nil {
					return "", err
				}
			}
			linkname := normalizeLinkname(string(bytes.TrimSpace(linknameBytes)))
			return linkname, nil
		}
		return &TargetStateSymlink{
			linknameFunc: linknameFunc,
		}, nil
	}
}

// newSourceStateFile returns a possibly new target RalPath and a new
// SourceStateFile.
func (s *SourceState) newSourceStateFile(
	absPath AbsPath,
	sourceRelPath SourceRelPath,
	fileAttr FileAttr,
	targetRelPath RelPath,
) (RelPath, *SourceStateFile) {
	contentsFunc := sync.OnceValues(func() ([]byte, error) {
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

	var targetStateEntryFunc TargetStateEntryFunc
	switch fileAttr.Type {
	case SourceFileTypeCreate:
		targetStateEntryFunc = s.newCreateTargetStateEntryFunc(sourceRelPath, fileAttr, contentsFunc)
	case SourceFileTypeFile:
		targetStateEntryFunc = s.newFileTargetStateEntryFunc(sourceRelPath, fileAttr, contentsFunc)
	case SourceFileTypeModify:
		// If the target has an extension, determine if it indicates an
		// interpreter to use.
		extension := strings.ToLower(strings.TrimPrefix(targetRelPath.Ext(), "."))
		if interpreter, ok := s.interpreters[extension]; ok {
			// For modify scripts, the script extension is not considered part
			// of the target name, so remove it.
			targetRelPath = targetRelPath.Slice(0, targetRelPath.Len()-len(extension)-1)
			targetStateEntryFunc = s.newModifyTargetStateEntryFunc(sourceRelPath, fileAttr, contentsFunc, &interpreter)
		} else {
			targetStateEntryFunc = s.newModifyTargetStateEntryFunc(sourceRelPath, fileAttr, contentsFunc, nil)
		}
	case SourceFileTypeRemove:
		targetStateEntryFunc = s.newRemoveTargetStateEntryFunc()
	case SourceFileTypeScript:
		// If the script has an extension, determine if it indicates an
		// interpreter to use.
		extension := strings.ToLower(strings.TrimPrefix(targetRelPath.Ext(), "."))
		if interpreter, ok := s.interpreters[extension]; ok {
			targetStateEntryFunc = s.newScriptTargetStateEntryFunc(
				sourceRelPath,
				fileAttr,
				targetRelPath,
				contentsFunc,
				&interpreter,
			)
		} else {
			targetStateEntryFunc = s.newScriptTargetStateEntryFunc(sourceRelPath, fileAttr, targetRelPath, contentsFunc, nil)
		}
	case SourceFileTypeSymlink:
		targetStateEntryFunc = s.newSymlinkTargetStateEntryFunc(sourceRelPath, fileAttr, contentsFunc)
	default:
		panic(fmt.Sprintf("%d: unsupported type", fileAttr.Type))
	}

	return targetRelPath, &SourceStateFile{
		origin:               SourceStateOriginAbsPath(absPath),
		sourceRelPath:        sourceRelPath,
		attr:                 fileAttr,
		contentsFunc:         contentsFunc,
		contentsSHA256Func:   lazySHA256(contentsFunc),
		targetStateEntryFunc: targetStateEntryFunc,
	}
}

// newSourceStateDirEntry returns a SourceStateEntry constructed from a
// directory in s.
func (s *SourceState) newSourceStateDirEntry(
	actualStateDir *ActualStateDir,
	fileInfo fs.FileInfo,
	parentSourceRelPath SourceRelPath,
	options *AddOptions,
) *SourceStateDir {
	dirAttr := DirAttr{
		TargetName: fileInfo.Name(),
		Exact:      options.Exact,
		Private:    isPrivate(fileInfo),
		ReadOnly:   isReadOnly(fileInfo),
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelDirPath(dirAttr.SourceName()))
	return &SourceStateDir{
		attr:          dirAttr,
		origin:        actualStateDir,
		sourceRelPath: sourceRelPath,
		targetStateEntry: &TargetStateDir{
			perm: fs.ModePerm &^ s.umask,
		},
	}
}

// newSourceStateFileEntryFromFile returns a SourceStateEntry constructed from a
// file in s.
func (s *SourceState) newSourceStateFileEntryFromFile(
	actualStateFile *ActualStateFile,
	fileInfo fs.FileInfo,
	parentSourceRelPath SourceRelPath,
	options *AddOptions,
) (*SourceStateFile, error) {
	fileAttr := FileAttr{
		TargetName: fileInfo.Name(),
		Encrypted:  options.Encrypt,
		Executable: IsExecutable(fileInfo),
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
	if options.Template {
		if !utf8.Valid(contents) {
			s.warnFunc("%s: invalid UTF-8\n", fileInfo.Name())
		}
		for _, byteOrderMark := range byteOrderMarks {
			if bytes.HasPrefix(contents, byteOrderMark.prefix) && byteOrderMark.name != "UTF-8" {
				s.warnFunc(
					"%s: detected %s byte order mark, ensure that template is in UTF-8\n",
					fileInfo.Name(),
					byteOrderMark.name,
				)
			}
		}
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
	contentsFunc := eagerNoErr(contents)
	contentsSHA256Func := lazySHA256(contentsFunc)
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
	return &SourceStateFile{
		attr:               fileAttr,
		origin:             actualStateFile,
		sourceRelPath:      sourceRelPath,
		contentsFunc:       contentsFunc,
		contentsSHA256Func: contentsSHA256Func,
		targetStateEntry: &TargetStateFile{
			contentsFunc:       contentsFunc,
			contentsSHA256Func: contentsSHA256Func,
			empty:              len(contents) == 0,
			perm:               0o666 &^ s.umask,
		},
	}, nil
}

// newSourceStateFileEntryFromSymlink returns a SourceStateEntry constructed
// from a symlink in s.
func (s *SourceState) newSourceStateFileEntryFromSymlink(
	actualStateSymlink *ActualStateSymlink,
	fileInfo fs.FileInfo,
	parentSourceRelPath SourceRelPath,
	options *AddOptions,
) (*SourceStateFile, error) {
	linkname, err := actualStateSymlink.Linkname()
	if err != nil {
		return nil, err
	}
	contents := []byte(linkname)
	isTemplate := false
	switch {
	case options.AutoTemplate:
		contents, isTemplate = autoTemplate(contents, s.TemplateData())
	case options.Template:
		isTemplate = true
	case !options.Template && options.TemplateSymlinks:
		switch {
		case strings.HasPrefix(linkname, s.sourceDirAbsPath.String()+"/"):
			contents = []byte("{{ .chezmoi.sourceDir }}/" + linkname[s.sourceDirAbsPath.Len()+1:])
			isTemplate = true
		case strings.HasPrefix(linkname, s.destDirAbsPath.String()+"/"):
			contents = []byte("{{ .chezmoi.homeDir }}/" + linkname[s.destDirAbsPath.Len()+1:])
			isTemplate = true
		}
	}
	contents = append(contents, '\n')
	contentsFunc := eagerNoErr(contents)
	contentsSHA256Func := lazySHA256(contentsFunc)
	fileAttr := FileAttr{
		TargetName: fileInfo.Name(),
		Type:       SourceFileTypeSymlink,
		Template:   isTemplate,
	}
	sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
	return &SourceStateFile{
		attr:               fileAttr,
		sourceRelPath:      sourceRelPath,
		contentsFunc:       contentsFunc,
		contentsSHA256Func: contentsSHA256Func,
		origin:             SourceStateOriginAbsPath(s.sourceDirAbsPath.Join(sourceRelPath.RelPath())),
		targetStateEntry: &TargetStateFile{
			contentsFunc:       contentsFunc,
			contentsSHA256Func: contentsSHA256Func,
			perm:               0o666 &^ s.umask,
		},
	}, nil
}

// populateImplicitParentDirs creates implicit parent directories for
// externalRelPath.
func (s *SourceState) populateImplicitParentDirs(
	externalRelPath RelPath,
	external *External,
	sourceStateEntries map[RelPath][]SourceStateEntry,
) map[RelPath][]SourceStateEntry {
	for relPath := externalRelPath.Dir(); relPath != DotRelPath; relPath = relPath.Dir() {
		sourceStateEntries[relPath] = append(sourceStateEntries[relPath],
			&SourceStateImplicitDir{
				origin: external,
				targetStateEntry: &TargetStateDir{
					perm: fs.ModePerm &^ s.umask,
				},
			},
		)
	}
	return sourceStateEntries
}

// readExternal reads an external and returns its SourceStateEntries.
func (s *SourceState) readExternal(
	ctx context.Context,
	externalRelPath RelPath,
	parentSourceRelPath SourceRelPath,
	external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	switch external.Type {
	case ExternalTypeArchive:
		return s.readExternalArchive(ctx, externalRelPath, parentSourceRelPath, external, options)
	case ExternalTypeArchiveFile:
		return s.readExternalArchiveFile(ctx, externalRelPath, parentSourceRelPath, external, options)
	case ExternalTypeFile:
		return s.readExternalFile(ctx, externalRelPath, parentSourceRelPath, external, options)
	case ExternalTypeGitRepo:
		return nil, nil
	case "":
		return nil, fmt.Errorf("%s: missing external type", externalRelPath)
	default:
		return nil, fmt.Errorf("%s: unknown external type: %s", externalRelPath, external.Type)
	}
}

// readExternalArchive reads an external archive and returns its
// SourceStateEntries.
func (s *SourceState) readExternalArchive(
	ctx context.Context,
	externalRelPath RelPath,
	parentSourceRelPath SourceRelPath,
	external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	data, urlStr, format, err := s.readExternalArchiveData(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, err
	}

	dirAttr := DirAttr{
		TargetName: externalRelPath.Base(),
		Exact:      external.Exact,
	}
	sourceStateDir := &SourceStateDir{
		attr:          dirAttr,
		origin:        external,
		sourceRelPath: parentSourceRelPath.Join(NewSourceRelPath(dirAttr.SourceName())),
		targetStateEntry: &TargetStateDir{
			perm: fs.ModePerm &^ s.umask,
			sourceAttr: SourceAttr{
				External: true,
			},
		},
	}
	sourceStateEntries := map[RelPath][]SourceStateEntry{
		externalRelPath: {sourceStateDir},
	}

	patternSet := NewPatternSet()
	for _, includePattern := range external.Include {
		if err := patternSet.Add(includePattern, PatternSetInclude); err != nil {
			return nil, err
		}
	}
	for _, excludePattern := range external.Exclude {
		if err := patternSet.Add(excludePattern, PatternSetExclude); err != nil {
			return nil, err
		}
	}

	sourceRelPaths := make(map[RelPath]SourceRelPath)
	if err := WalkArchive(data, format, func(name string, fileInfo fs.FileInfo, r io.Reader, linkname string) error {
		// Perform matching against the name before stripping any components,
		// otherwise it is not possible to differentiate between
		// identically-named files at the same level.
		if patternSet.Match(name) == PatternSetMatchExclude {
			// In case that `name` is a directory which matched an explicit
			// exclude pattern, return fs.SkipDir to exclude not just the
			// directory itself but also everything it contains (recursively).
			if fileInfo.IsDir() && len(patternSet.ExcludePatterns) > 0 {
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
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		dirTargetRelPath, _ := targetRelPath.Split()
		dirSourceRelPath := sourceRelPaths[dirTargetRelPath]

		var sourceStateEntry SourceStateEntry
		switch {
		case fileInfo.IsDir():
			targetStateEntry := &TargetStateDir{
				perm: fileInfo.Mode().Perm() &^ s.umask,
				sourceAttr: SourceAttr{
					External: true,
				},
			}
			dirAttr := DirAttr{
				TargetName: fileInfo.Name(),
				Exact:      external.Exact,
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			sourceStateEntry = &SourceStateDir{
				attr:             dirAttr,
				origin:           external,
				sourceRelPath:    parentSourceRelPath.Join(dirSourceRelPath, NewSourceRelPath(dirAttr.SourceName())),
				targetStateEntry: targetStateEntry,
			}
		case fileInfo.Mode()&fs.ModeType == 0:
			contents, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}

			if !external.Archive.ExtractAppleDoubleFiles && isAppleDoubleFile(name, contents) {
				return nil
			}

			contentsFunc := eagerNoErr(contents)
			contentsSHA256Func := lazySHA256(contentsFunc)
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeFile,
				Empty:      fileInfo.Size() == 0,
				Executable: IsExecutable(fileInfo),
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			sourceRelPath := NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))
			targetStateEntry := &TargetStateFile{
				contentsFunc:       contentsFunc,
				contentsSHA256Func: contentsSHA256Func,
				empty:              fileAttr.Empty,
				perm:               fileAttr.perm() &^ s.umask,
				sourceAttr: SourceAttr{
					External: true,
				},
			}
			sourceStateEntry = &SourceStateFile{
				attr:               fileAttr,
				contentsFunc:       contentsFunc,
				contentsSHA256Func: contentsSHA256Func,
				origin:             external,
				sourceRelPath:      parentSourceRelPath.Join(dirSourceRelPath, sourceRelPath),
				targetStateEntry:   targetStateEntry,
			}
		case fileInfo.Mode()&fs.ModeType == fs.ModeSymlink:
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeSymlink,
			}
			sourceRelPath := NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix()))
			targetStateEntry := &TargetStateSymlink{
				linknameFunc: sync.OnceValues(func() (string, error) {
					return linkname, nil
				}),
				sourceAttr: SourceAttr{
					External: true,
				},
			}
			sourceStateEntry = &SourceStateFile{
				attr:             fileAttr,
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
		return nil, fmt.Errorf("%s: %s: %w", externalRelPath, urlStr, err)
	}

	return s.populateImplicitParentDirs(externalRelPath, external, sourceStateEntries), nil
}

// readExternalArchiveData reads an external archive's data and returns its data
// and format.
func (s *SourceState) readExternalArchiveData(
	ctx context.Context,
	externalRelPath RelPath,
	external *External,
	options *ReadOptions,
) ([]byte, string, ArchiveFormat, error) {
	data, urlStr, err := s.getExternalData(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, "", ArchiveFormatUnknown, err
	}

	externalURL, err := url.Parse(urlStr)
	if err != nil {
		err := fmt.Errorf("%s: %s: %w", externalRelPath, urlStr, err)
		return nil, urlStr, ArchiveFormatUnknown, err
	}
	urlPath := externalURL.Path
	if external.Encrypted {
		urlPath = strings.TrimSuffix(urlPath, s.encryption.EncryptedSuffix())
	}

	format := external.Format
	if format == ArchiveFormatUnknown {
		format = GuessArchiveFormat(urlPath, data)
	}

	return data, urlStr, format, nil
}

// readExternalArchiveFile reads a file from an external archive and returns its
// SourceStateEntries.
func (s *SourceState) readExternalArchiveFile(
	ctx context.Context,
	externalRelPath RelPath,
	parentSourceRelPath SourceRelPath,
	external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	if external.ArchivePath == "" {
		return nil, fmt.Errorf("%s: missing path", externalRelPath)
	}

	data, urlStr, format, err := s.readExternalArchiveData(ctx, externalRelPath, external, options)
	if err != nil {
		return nil, err
	}

	var sourceStateEntry SourceStateEntry
	if err := WalkArchive(data, format, func(name string, fileInfo fs.FileInfo, r io.Reader, linkname string) error {
		if external.StripComponents > 0 {
			components := strings.Split(name, "/")
			if len(components) <= external.StripComponents {
				return nil
			}
			name = path.Join(components[external.StripComponents:]...)
		}
		switch {
		case name == "":
			return nil
		case name != external.ArchivePath:
			// If this entry is a directory and it cannot contain the file we
			// are looking for then skip this directory.
			if fileInfo.IsDir() && !strings.HasPrefix(external.ArchivePath, name) {
				return fs.SkipDir
			}
			return nil
		case fileInfo.Mode()&fs.ModeType == 0:
			contents, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}

			if !external.Archive.ExtractAppleDoubleFiles && isAppleDoubleFile(name, contents) {
				return nil
			}

			contentsFunc := eagerNoErr(contents)
			contentsSHA256Func := lazySHA256(contentsFunc)
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeFile,
				Empty:      fileInfo.Size() == 0,
				Executable: IsExecutable(fileInfo) || external.Executable,
				Private:    isPrivate(fileInfo) || external.Private,
				ReadOnly:   isReadOnly(fileInfo) || external.ReadOnly,
			}
			sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
			targetStateEntry := &TargetStateFile{
				contentsFunc:       contentsFunc,
				contentsSHA256Func: contentsSHA256Func,
				empty:              fileAttr.Empty,
				perm:               fileAttr.perm() &^ s.umask,
				sourceAttr: SourceAttr{
					External: true,
				},
			}
			sourceStateEntry = &SourceStateFile{
				attr:               fileAttr,
				contentsFunc:       contentsFunc,
				contentsSHA256Func: contentsSHA256Func,
				origin:             external,
				sourceRelPath:      sourceRelPath,
				targetStateEntry:   targetStateEntry,
			}
			return fs.SkipAll
		case fileInfo.Mode()&fs.ModeType == fs.ModeSymlink:
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeSymlink,
			}
			sourceRelPath := parentSourceRelPath.Join(NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())))
			targetStateEntry := &TargetStateSymlink{
				linknameFunc: sync.OnceValues(func() (string, error) {
					return linkname, nil
				}),
				sourceAttr: SourceAttr{
					External: true,
				},
			}
			sourceStateEntry = &SourceStateFile{
				attr:             fileAttr,
				origin:           external,
				sourceRelPath:    sourceRelPath,
				targetStateEntry: targetStateEntry,
			}
			return fs.SkipAll
		default:
			return fmt.Errorf("%s: unsupported mode %o", name, fileInfo.Mode()&fs.ModeType)
		}
	}); err != nil {
		return nil, err
	}
	if sourceStateEntry == nil {
		return nil, fmt.Errorf("%s: path not found in %s", external.ArchivePath, urlStr)
	}

	return s.populateImplicitParentDirs(externalRelPath, external, map[RelPath][]SourceStateEntry{
		externalRelPath: {sourceStateEntry},
	}), nil
}

// readExternalDir returns all source state entries in an external_ dir.
func (s *SourceState) readExternalDir(
	rootSourceAbsPath AbsPath,
	rootSourceRelPath SourceRelPath,
	rootTargetRelPath RelPath,
) (map[RelPath][]SourceStateEntry, error) {
	sourceStateEntries := make(map[RelPath][]SourceStateEntry)
	walkFunc := func(absPath AbsPath, fileInfo fs.FileInfo, err error) error {
		switch {
		case err != nil:
			return err
		case absPath == rootSourceAbsPath:
			return nil
		}
		relPath := absPath.MustTrimDirPrefix(rootSourceAbsPath)
		targetRelPath := rootTargetRelPath.Join(relPath)
		if s.Ignore(targetRelPath) {
			if fileInfo.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		var sourceStateEntry SourceStateEntry
		switch fileInfo.Mode().Type() {
		case 0:
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeFile,
				Empty:      true,
				Executable: IsExecutable(fileInfo),
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			contentsFunc := sync.OnceValues(func() ([]byte, error) {
				return s.system.ReadFile(absPath)
			})
			contentsSHA256Func := lazySHA256(contentsFunc)
			sourceStateEntry = &SourceStateFile{
				origin:             SourceStateOriginAbsPath(absPath),
				attr:               fileAttr,
				contentsFunc:       contentsFunc,
				contentsSHA256Func: contentsSHA256Func,
				sourceRelPath:      rootSourceRelPath.Join(relPath.SourceRelPath()),
				targetStateEntry: &TargetStateFile{
					contentsFunc:       contentsFunc,
					contentsSHA256Func: contentsSHA256Func,
					empty:              true,
					perm:               fileAttr.perm() &^ s.umask,
				},
			}
		case fs.ModeDir:
			dirAttr := DirAttr{
				TargetName: fileInfo.Name(),
				Exact:      true,
				Private:    isPrivate(fileInfo),
				ReadOnly:   isReadOnly(fileInfo),
			}
			sourceStateEntry = &SourceStateDir{
				origin:        SourceStateOriginAbsPath(absPath),
				sourceRelPath: rootSourceRelPath.Join(relPath.SourceRelDirPath()),
				attr:          dirAttr,
				targetStateEntry: &TargetStateDir{
					perm: dirAttr.perm() &^ s.umask,
				},
			}
		case fs.ModeSymlink:
			fileAttr := FileAttr{
				TargetName: fileInfo.Name(),
				Type:       SourceFileTypeFile,
			}
			linknameFunc := sync.OnceValues(func() (string, error) {
				return s.system.Readlink(absPath)
			})
			contentsFunc := sync.OnceValues(func() ([]byte, error) {
				linkname, err := linknameFunc()
				if err != nil {
					return nil, err
				}
				return []byte(linkname), nil
			})
			sourceStateEntry = &SourceStateFile{
				origin:             SourceStateOriginAbsPath(absPath),
				attr:               fileAttr,
				contentsFunc:       contentsFunc,
				contentsSHA256Func: lazySHA256(contentsFunc),
				sourceRelPath:      rootSourceRelPath.Join(relPath.SourceRelPath()),
				targetStateEntry: &TargetStateSymlink{
					linknameFunc: linknameFunc,
				},
			}
		}
		sourceStateEntries[targetRelPath] = append(sourceStateEntries[targetRelPath], sourceStateEntry)
		return nil
	}
	if err := Walk(s.system, rootSourceAbsPath, walkFunc); err != nil {
		return nil, err
	}
	return sourceStateEntries, nil
}

// readExternalFile reads an external file and returns its SourceStateEntries.
func (s *SourceState) readExternalFile(
	ctx context.Context,
	externalRelPath RelPath,
	parentSourceRelPath SourceRelPath,
	external *External,
	options *ReadOptions,
) (map[RelPath][]SourceStateEntry, error) {
	contentsFunc := sync.OnceValues(func() ([]byte, error) {
		data, _, err := s.getExternalData(ctx, externalRelPath, external, options)
		return data, err
	})
	fileAttr := FileAttr{
		Empty:      true,
		Executable: external.Executable,
		Private:    external.Private,
		ReadOnly:   external.ReadOnly,
	}
	targetStateEntry := &TargetStateFile{
		contentsFunc:       contentsFunc,
		contentsSHA256Func: lazySHA256(contentsFunc),
		empty:              fileAttr.Empty,
		perm:               fileAttr.perm() &^ s.umask,
		sourceAttr: SourceAttr{
			External: true,
		},
	}
	sourceStateEntry := &SourceStateFile{
		origin: external,
		sourceRelPath: parentSourceRelPath.Join(
			NewSourceRelPath(fileAttr.SourceName(s.encryption.EncryptedSuffix())),
		),
		targetStateEntry: targetStateEntry,
	}
	return s.populateImplicitParentDirs(externalRelPath, external, map[RelPath][]SourceStateEntry{
		externalRelPath: {sourceStateEntry},
	}), nil
}

// readScriptsDir reads all scripts in scriptsDirAbsPath.
func (s *SourceState) readScriptsDir(ctx context.Context, scriptsDirAbsPath AbsPath) (map[RelPath][]SourceStateEntry, error) {
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
				return fs.SkipDir
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
			return &UnsupportedFileTypeError{
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
	actualStateEntry ActualStateEntry,
	destAbsPath AbsPath,
	fileInfo fs.FileInfo,
	parentSourceRelPath SourceRelPath,
	options *AddOptions,
) (SourceStateEntry, error) {
	switch actualStateEntry := actualStateEntry.(type) {
	case *ActualStateAbsent:
		return nil, fmt.Errorf("%s: not found", destAbsPath)
	case *ActualStateDir:
		return s.newSourceStateDirEntry(actualStateEntry, fileInfo, parentSourceRelPath, options), nil
	case *ActualStateFile:
		return s.newSourceStateFileEntryFromFile(actualStateEntry, fileInfo, parentSourceRelPath, options)
	case *ActualStateSymlink:
		return s.newSourceStateFileEntryFromSymlink(actualStateEntry, fileInfo, parentSourceRelPath, options)
	default:
		panic(fmt.Sprintf("%T: unsupported type", actualStateEntry))
	}
}

func (e *External) IsExternal() bool {
	return true
}

func (e *External) Path() AbsPath {
	return e.sourceAbsPath
}

func (e *External) OriginString() string {
	urlStr := cmp.Or(append([]string{e.URL}, e.URLs...)...)
	return urlStr + " defined in " + e.sourceAbsPath.String()
}

// canonicalSourceStateEntry returns the canonical SourceStateEntry for the
// given sourceStateEntries.
//
// This only applies to directories, where SourceStateImplicitDirs are
// considered equivalent to all SourceStateDirs.
func canonicalSourceStateEntry(sourceStateEntries []SourceStateEntry) (SourceStateEntry, bool) {
	// Find all directories to check for equivalence.
	var firstSourceStateDir *SourceStateDir
	sourceStateDirs := make([]SourceStateEntry, len(sourceStateEntries))
	for i, sourceStateEntry := range sourceStateEntries {
		switch sourceStateEntry := sourceStateEntry.(type) {
		case *SourceStateDir:
			firstSourceStateDir = sourceStateEntry
			sourceStateDirs[i] = sourceStateEntry
		case *SourceStateImplicitDir:
			sourceStateDirs[i] = sourceStateEntry
		default:
			return nil, false
		}
	}

	switch len(sourceStateDirs) {
	case 0:
		// If there are no SourceStateDirs then there are no equivalent directories.
		return nil, false
	case 1:
		return sourceStateDirs[0], true
	default:
		// Check for equivalence.
		for _, sourceStateDir := range sourceStateDirs {
			switch sourceStateDir := sourceStateDir.(type) {
			case *SourceStateDir:
				if sourceStateDir.attr != firstSourceStateDir.attr {
					return nil, false
				}
			case *SourceStateImplicitDir:
				// SourceStateImplicitDirs are considered equivalent to all other
				// directories.
			}
		}
		// If all directories are equivalent then return the first real
		// *SourceStateDir, if it exists.
		if firstSourceStateDir != nil {
			return firstSourceStateDir, true
		}
		// Otherwise, return the first entry which is a *SourceStateImplicitDir.
		return sourceStateDirs[0], true
	}
}

// isAppleDoubleFile returns true if the file looks like and has the
// expected signature of an AppleDouble file.
func isAppleDoubleFile(name string, contents []byte) bool {
	return strings.HasPrefix(path.Base(name), appleDoubleNamePrefix) && bytes.HasPrefix(contents, appleDoubleContentsPrefix)
}
