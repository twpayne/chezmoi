package chezmoi

import (
	"bytes"
	"encoding/hex"
	"os"
	"runtime"
	"time"
)

// A TargetStateEntry represents the state of an entry in the target state.
type TargetStateEntry interface {
	Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error
	EntryState(umask os.FileMode) (*EntryState, error)
	Evaluate() error
	SkipApply(persistentState PersistentState) (bool, error)
}

// A TargetStateAbsent represents the absence of an entry in the target state.
type TargetStateAbsent struct{}

// A TargetStateDir represents the state of a directory in the target state.
type TargetStateDir struct {
	perm os.FileMode
}

// A TargetStateFile represents the state of a file in the target state.
type TargetStateFile struct {
	*lazyContents
	perm os.FileMode
}

// A TargetStatePresent represents the presence of an entry in the target state.
type TargetStatePresent struct {
	*lazyContents
	perm os.FileMode
}

// A TargetStateRenameDir represents the renaming of a directory in the target
// state.
type TargetStateRenameDir struct {
	oldRelPath RelPath
	newRelPath RelPath
}

// A TargetStateScript represents the state of a script.
type TargetStateScript struct {
	*lazyContents
	name RelPath
	once bool
}

// A TargetStateSymlink represents the state of a symlink in the target state.
type TargetStateSymlink struct {
	*lazyLinkname
}

// A scriptState records the state of a script that has been run.
type scriptState struct {
	Name  string    `json:"name" toml:"name" yaml:"name"`
	RunAt time.Time `json:"runAt" toml:"runAt" yaml:"runAt"`
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateAbsent) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	if _, ok := actualStateEntry.(*ActualStateAbsent); ok {
		return nil
	}
	return system.RemoveAll(actualStateEntry.Path())
}

// EntryState returns t's entry state.
func (t *TargetStateAbsent) EntryState(umask os.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeAbsent,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateAbsent) Evaluate() error {
	return nil
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateAbsent) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply updates actualStateEntry to match t. It does not recurse.
func (t *TargetStateDir) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	if actualStateDir, ok := actualStateEntry.(*ActualStateDir); ok {
		if runtime.GOOS == "windows" || actualStateDir.perm == t.perm&^umask {
			return nil
		}
		return system.Chmod(actualStateDir.Path(), t.perm&^umask)
	}
	if err := actualStateEntry.Remove(system); err != nil {
		return err
	}
	return system.Mkdir(actualStateEntry.Path(), t.perm&^umask)
}

// EntryState returns t's entry state.
func (t *TargetStateDir) EntryState(umask os.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypeDir,
		Mode: os.ModeDir | t.perm&^umask,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateDir) Evaluate() error {
	return nil
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateDir) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateFile) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	if actualStateFile, ok := actualStateEntry.(*ActualStateFile); ok {
		// Compare file contents using only their SHA256 sums. This is so that
		// we can compare last-written states without storing the full contents
		// of each file written.
		actualContentsSHA256, err := actualStateFile.ContentsSHA256()
		if err != nil {
			return err
		}
		contentsSHA256, err := t.ContentsSHA256()
		if err != nil {
			return err
		}
		if bytes.Equal(actualContentsSHA256, contentsSHA256) {
			if runtime.GOOS == "windows" || actualStateFile.perm == t.perm&^umask {
				return nil
			}
			return system.Chmod(actualStateFile.Path(), t.perm&^umask)
		}
	} else if err := actualStateEntry.Remove(system); err != nil {
		return err
	}
	contents, err := t.Contents()
	if err != nil {
		return err
	}
	return system.WriteFile(actualStateEntry.Path(), contents, t.perm&^umask)
}

// EntryState returns t's entry state.
func (t *TargetStateFile) EntryState(umask os.FileMode) (*EntryState, error) {
	contents, err := t.Contents()
	if err != nil {
		return nil, err
	}
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return nil, err
	}
	return &EntryState{
		Type:           EntryStateTypeFile,
		Mode:           t.perm &^ umask,
		ContentsSHA256: hexBytes(contentsSHA256),
		contents:       contents,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateFile) Evaluate() error {
	_, err := t.ContentsSHA256()
	return err
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateFile) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply updates actualStateEntry to match t.
func (t *TargetStatePresent) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	if actualStateFile, ok := actualStateEntry.(*ActualStateFile); ok {
		if runtime.GOOS == "windows" || actualStateFile.perm == t.perm&^umask {
			return nil
		}
		return system.Chmod(actualStateFile.Path(), t.perm&^umask)
	} else if err := actualStateEntry.Remove(system); err != nil {
		return err
	}
	contents, err := t.Contents()
	if err != nil {
		return err
	}
	return system.WriteFile(actualStateEntry.Path(), contents, t.perm&^umask)
}

// EntryState returns t's entry state.
func (t *TargetStatePresent) EntryState(umask os.FileMode) (*EntryState, error) {
	return &EntryState{
		Type: EntryStateTypePresent,
		Mode: t.perm &^ umask,
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStatePresent) Evaluate() error {
	_, err := t.ContentsSHA256()
	return err
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStatePresent) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply renames actualStateEntry.
func (t *TargetStateRenameDir) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	dir := actualStateEntry.Path().Dir()
	return system.Rename(dir.Join(t.oldRelPath), dir.Join(t.newRelPath))
}

// EntryState returns t's entry state.
func (t *TargetStateRenameDir) EntryState(umask os.FileMode) (*EntryState, error) {
	return nil, nil
}

// Evaluate does nothing.
func (t *TargetStateRenameDir) Evaluate() error {
	return nil
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateRenameDir) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply runs t.
func (t *TargetStateScript) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return err
	}
	key := []byte(hex.EncodeToString(contentsSHA256))
	if t.once {
		switch scriptState, err := persistentState.Get(scriptStateBucket, key); {
		case err != nil:
			return err
		case scriptState != nil:
			return nil
		}
	}
	contents, err := t.Contents()
	if err != nil {
		return err
	}
	runAt := time.Now().UTC()
	if !isEmpty(contents) {
		if err := system.RunScript(t.name, actualStateEntry.Path().Dir(), contents); err != nil {
			return err
		}
	}
	return persistentStateSet(persistentState, scriptStateBucket, key, &scriptState{
		Name:  string(t.name),
		RunAt: runAt,
	})
}

// EntryState returns t's entry state.
func (t *TargetStateScript) EntryState(umask os.FileMode) (*EntryState, error) {
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return nil, err
	}
	return &EntryState{
		Type:           EntryStateTypeScript,
		ContentsSHA256: hexBytes(contentsSHA256),
	}, nil
}

// Evaluate evaluates t.
func (t *TargetStateScript) Evaluate() error {
	_, err := t.ContentsSHA256()
	return err
}

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateScript) SkipApply(persistentState PersistentState) (bool, error) {
	if !t.once {
		return false, nil
	}
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return false, err
	}
	key := []byte(hex.EncodeToString(contentsSHA256))
	switch scriptState, err := persistentState.Get(scriptStateBucket, key); {
	case err != nil:
		return false, err
	case scriptState != nil:
		return true, nil
	default:
		return false, nil
	}
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateSymlink) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry, umask os.FileMode) error {
	if actualStateSymlink, ok := actualStateEntry.(*ActualStateSymlink); ok {
		actualLinkname, err := actualStateSymlink.Linkname()
		if err != nil {
			return err
		}
		linkname, err := t.Linkname()
		if err != nil {
			return err
		}
		if actualLinkname == linkname {
			return nil
		}
	}
	linkname, err := t.Linkname()
	if err != nil {
		return err
	}
	if err := actualStateEntry.Remove(system); err != nil {
		return err
	}
	return system.WriteSymlink(linkname, actualStateEntry.Path())
}

// EntryState returns t's entry state.
func (t *TargetStateSymlink) EntryState(umask os.FileMode) (*EntryState, error) {
	linkname, err := t.Linkname()
	if err != nil {
		return nil, err
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

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateSymlink) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}
