package chezmoi

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/fs"
	"runtime"
	"time"
)

// A TargetStateEntry represents the state of an entry in the target state.
type TargetStateEntry interface {
	Apply(
		system System,
		persistentState PersistentState,
		actualStateEntry ActualStateEntry,
	) (bool, error)
	EntryState(umask fs.FileMode) (*EntryState, error)
	Evaluate() error
	SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error)
	SourceAttr() SourceAttr
}

// A TargetStateModifyDirWithCmd represents running a command that modifies
// a directory.
type TargetStateModifyDirWithCmd struct {
	cmd           *lazyCommand
	forceRefresh  bool
	refreshPeriod Duration
	sourceAttr    SourceAttr
}

// A TargetStateDir represents the state of a directory in the target state.
type TargetStateDir struct {
	perm       fs.FileMode
	sourceAttr SourceAttr
}

// A TargetStateFile represents the state of a file in the target state.
type TargetStateFile struct {
	*lazyContents
	empty      bool
	overwrite  bool
	perm       fs.FileMode
	sourceAttr SourceAttr
}

// A TargetStateRemove represents the absence of an entry in the target state.
type TargetStateRemove struct{}

// A TargetStateScript represents the state of a script.
type TargetStateScript struct {
	*lazyContents
	name          RelPath
	interpreter   *Interpreter
	condition     ScriptCondition
	sourceAttr    SourceAttr
	sourceRelPath SourceRelPath
}

// A TargetStateSymlink represents the state of a symlink in the target state.
type TargetStateSymlink struct {
	*lazyLinkname
	sourceAttr SourceAttr
}

// A modifyDirWithCmdState records the state of a directory modified by a
// command.
type modifyDirWithCmdState struct {
	Name  AbsPath   `json:"name"  yaml:"name"`
	RunAt time.Time `json:"runAt" yaml:"runAt"`
}

// A scriptState records the state of a script that has been run.
type scriptState struct {
	Name  RelPath   `json:"name"  yaml:"name"`
	RunAt time.Time `json:"runAt" yaml:"runAt"`
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateModifyDirWithCmd) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	if _, ok := actualStateEntry.(*ActualStateDir); !ok {
		if err := actualStateEntry.Remove(system); err != nil {
			return false, err
		}
	}

	runAt := time.Now().UTC()
	if err := system.RunCmd(t.cmd.Command()); err != nil {
		return false, fmt.Errorf("%s: %w", actualStateEntry.Path(), err)
	}

	modifyDirWithCmdStateKey := []byte(actualStateEntry.Path().String())
	if err := PersistentStateSet(
		persistentState, GitRepoExternalStateBucket, modifyDirWithCmdStateKey, &modifyDirWithCmdState{
			Name:  actualStateEntry.Path(),
			RunAt: runAt,
		}); err != nil {
		return false, err
	}

	return true, nil
}

// EntryState returns t's entry state.
func (t *TargetStateModifyDirWithCmd) EntryState(umask fs.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeDir,
		Mode: fs.ModeDir | fs.ModePerm&^umask,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateModifyDirWithCmd) Evaluate() error {
	return nil
}

// SkipApply implements TargetStateEntry.SkipApply.
func (t *TargetStateModifyDirWithCmd) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	if t.forceRefresh {
		return false, nil
	}
	modifyDirWithCmdKey := []byte(targetAbsPath.String())
	switch modifyDirWithCmdStateBytes, err := persistentState.Get(GitRepoExternalStateBucket, modifyDirWithCmdKey); {
	case err != nil:
		return false, err
	case modifyDirWithCmdStateBytes == nil:
		return false, nil
	default:
		var modifyDirWithCmdState modifyDirWithCmdState
		if err := stateFormat.Unmarshal(modifyDirWithCmdStateBytes, &modifyDirWithCmdState); err != nil {
			return false, err
		}
		if t.refreshPeriod == 0 {
			return true, nil
		}
		return time.Since(modifyDirWithCmdState.RunAt) < time.Duration(t.refreshPeriod), nil
	}
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateModifyDirWithCmd) SourceAttr() SourceAttr {
	return t.sourceAttr
}

// Apply updates actualStateEntry to match t. It does not recurse.
func (t *TargetStateDir) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	if actualStateDir, ok := actualStateEntry.(*ActualStateDir); ok {
		if runtime.GOOS == "windows" || actualStateDir.perm == t.perm {
			return false, nil
		}
		return true, system.Chmod(actualStateDir.Path(), t.perm)
	}
	if err := actualStateEntry.Remove(system); err != nil {
		return false, err
	}
	return true, system.Mkdir(actualStateEntry.Path(), t.perm)
}

// EntryState returns t's entry state.
func (t *TargetStateDir) EntryState(umask fs.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeDir,
		Mode: fs.ModeDir | t.perm&^umask,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateDir) Evaluate() error {
	return nil
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateDir) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	return false, nil
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateDir) SourceAttr() SourceAttr {
	return t.sourceAttr
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateFile) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	contents, err := t.Contents()
	if err != nil {
		return false, err
	}
	if !t.empty && isEmpty(contents) {
		if _, ok := actualStateEntry.(*ActualStateAbsent); ok {
			return false, nil
		}
		return true, system.RemoveAll(actualStateEntry.Path())
	}
	if actualStateFile, ok := actualStateEntry.(*ActualStateFile); ok {
		// Compare file contents using only their SHA256 sums. This is so that
		// we can compare last-written states without storing the full contents
		// of each file written.
		actualContentsSHA256, err := actualStateFile.ContentsSHA256()
		if err != nil {
			return false, err
		}
		contentsSHA256, err := t.ContentsSHA256()
		if err != nil {
			return false, err
		}
		if bytes.Equal(actualContentsSHA256, contentsSHA256) {
			if runtime.GOOS == "windows" || actualStateFile.perm == t.perm {
				return false, nil
			}
			return true, system.Chmod(actualStateFile.Path(), t.perm)
		}
	} else if err := actualStateEntry.Remove(system); err != nil {
		return false, err
	}
	return true, system.WriteFile(actualStateEntry.Path(), contents, t.perm)
}

// EntryState returns t's entry state.
func (t *TargetStateFile) EntryState(umask fs.FileMode) (*EntryState, error) {
	contents, err := t.Contents()
	if err != nil {
		return nil, err
	}
	if !t.empty && isEmpty(contents) {
		return &EntryState{
			Type: EntryStateTypeRemove,
		}, nil
	}
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return nil, err
	}
	return &EntryState{
		Type:           EntryStateTypeFile,
		Mode:           t.perm &^ umask,
		ContentsSHA256: HexBytes(contentsSHA256),
		contents:       contents,
		overwrite:      t.overwrite,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateFile) Evaluate() error {
	_, err := t.ContentsSHA256()
	return err
}

// Perm returns t's perm.
func (t *TargetStateFile) Perm(umask fs.FileMode) fs.FileMode {
	return t.perm &^ umask
}

// SkipApply implements TargetStateEntry.SkipApply.
func (t *TargetStateFile) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	return false, nil
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateFile) SourceAttr() SourceAttr {
	return t.sourceAttr
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateRemove) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	if _, ok := actualStateEntry.(*ActualStateAbsent); ok {
		return false, nil
	}
	return true, system.RemoveAll(actualStateEntry.Path())
}

// EntryState returns t's entry state.
func (t *TargetStateRemove) EntryState(umask fs.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeRemove,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateRemove) Evaluate() error {
	return nil
}

// SkipApply implements TargetStateEntry.SkipApply.
func (t *TargetStateRemove) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	return false, nil
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateRemove) SourceAttr() SourceAttr {
	return SourceAttr{}
}

// Apply runs t.
func (t *TargetStateScript) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	skipApply, err := t.SkipApply(persistentState, actualStateEntry.Path())
	if err != nil {
		return false, err
	}
	if skipApply {
		return false, nil
	}

	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return false, err
	}

	contents, err := t.Contents()
	if err != nil {
		return false, err
	}
	runAt := time.Now().UTC()
	if !isEmpty(contents) {
		if err := system.RunScript(t.name, actualStateEntry.Path().Dir(), contents, RunScriptOptions{
			Condition:     t.condition,
			Interpreter:   t.interpreter,
			SourceRelPath: t.sourceRelPath,
		}); err != nil {
			return false, err
		}
	}

	scriptStateKey := []byte(hex.EncodeToString(contentsSHA256))
	if err := PersistentStateSet(persistentState, ScriptStateBucket, scriptStateKey, &scriptState{
		Name:  t.name,
		RunAt: runAt,
	}); err != nil {
		return false, err
	}

	entryStateKey := actualStateEntry.Path().Bytes()
	if err := PersistentStateSet(persistentState, EntryStateBucket, entryStateKey, &EntryState{
		Type:           EntryStateTypeScript,
		ContentsSHA256: HexBytes(contentsSHA256),
	}); err != nil {
		return false, err
	}

	return true, nil
}

// EntryState returns t's entry state.
func (t *TargetStateScript) EntryState(umask fs.FileMode) (*EntryState, error) {
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return nil, err
	}
	return &EntryState{
		Type:           EntryStateTypeScript,
		ContentsSHA256: HexBytes(contentsSHA256),
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateScript) Evaluate() error {
	_, err := t.ContentsSHA256()
	return err
}

// SkipApply implements TargetStateEntry.SkipApply.
func (t *TargetStateScript) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	switch contents, err := t.Contents(); {
	case err != nil:
		return false, err
	case len(contents) == 0:
		return true, nil
	}
	switch t.condition {
	case ScriptConditionAlways:
		return false, nil
	case ScriptConditionOnce:
		contentsSHA256, err := t.ContentsSHA256()
		if err != nil {
			return false, err
		}
		scriptStateKey := []byte(hex.EncodeToString(contentsSHA256))
		switch scriptState, err := persistentState.Get(ScriptStateBucket, scriptStateKey); {
		case err != nil:
			return false, err
		case scriptState != nil:
			return true, nil
		}
	case ScriptConditionOnChange:
		entryStateKey := []byte(targetAbsPath.String())
		switch entryStateBytes, err := persistentState.Get(EntryStateBucket, entryStateKey); {
		case err != nil:
			return false, err
		case entryStateBytes != nil:
			var entryState EntryState
			if err := stateFormat.Unmarshal(entryStateBytes, &entryState); err != nil {
				return false, err
			}
			contentsSHA256, err := t.ContentsSHA256()
			if err != nil {
				return false, err
			}
			if bytes.Equal(entryState.ContentsSHA256.Bytes(), contentsSHA256) {
				return true, nil
			}
		}
	}
	return false, nil
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateScript) SourceAttr() SourceAttr {
	return t.sourceAttr
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateSymlink) Apply(
	system System,
	persistentState PersistentState,
	actualStateEntry ActualStateEntry,
) (bool, error) {
	linkname, err := t.Linkname()
	if err != nil {
		return false, err
	}
	if linkname == "" {
		if _, ok := actualStateEntry.(*ActualStateAbsent); ok {
			return false, nil
		}
		return true, system.RemoveAll(actualStateEntry.Path())
	}
	if actualStateSymlink, ok := actualStateEntry.(*ActualStateSymlink); ok {
		actualLinkname, err := actualStateSymlink.Linkname()
		if err != nil {
			return false, err
		}
		linkname, err := t.Linkname()
		if err != nil {
			return false, err
		}
		if normalizeLinkname(actualLinkname) == normalizeLinkname(linkname) {
			return false, nil
		}
	}
	if err := actualStateEntry.Remove(system); err != nil {
		return false, err
	}
	return true, system.WriteSymlink(linkname, actualStateEntry.Path())
}

// EntryState returns t's entry state.
func (t *TargetStateSymlink) EntryState(umask fs.FileMode) (*EntryState, error) {
	linkname, err := t.Linkname()
	if err != nil {
		return nil, err
	}
	if linkname == "" {
		return &EntryState{
			Type: EntryStateTypeRemove,
		}, nil
	}
	linknameSHA256, err := t.LinknameSHA256()
	if err != nil {
		return nil, err
	}
	return &EntryState{
		Type:           EntryStateTypeSymlink,
		ContentsSHA256: linknameSHA256,
		contents:       []byte(linkname),
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateSymlink) Evaluate() error {
	_, err := t.Linkname()
	return err
}

// SkipApply implements TargetStateEntry.SkipApply.
func (t *TargetStateSymlink) SkipApply(persistentState PersistentState, targetAbsPath AbsPath) (bool, error) {
	return false, nil
}

// SourceAttr implements TargetStateEntry.SourceAttr.
func (t *TargetStateSymlink) SourceAttr() SourceAttr {
	return t.sourceAttr
}
