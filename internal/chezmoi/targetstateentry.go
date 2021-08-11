package chezmoi

import (
	"bytes"
	"encoding/hex"
	"io/fs"
	"runtime"
	"time"
)

// A TargetStateEntry represents the state of an entry in the target state.
type TargetStateEntry interface {
	Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error)
	EntryState(umask fs.FileMode) (*EntryState, error)
	Evaluate() error
	SkipApply(persistentState PersistentState) (bool, error)
}

// A TargetStateDir represents the state of a directory in the target state.
type TargetStateDir struct {
	perm fs.FileMode
}

// A TargetStateFile represents the state of a file in the target state.
type TargetStateFile struct {
	*lazyContents
	empty     bool
	overwrite bool
	perm      fs.FileMode
}

// A TargetStateRemove represents the absence of an entry in the target state.
type TargetStateRemove struct{}

// A TargetStateScript represents the state of a script.
type TargetStateScript struct {
	*lazyContents
	name        RelPath
	interpreter *Interpreter
	once        bool
}

// A TargetStateSymlink represents the state of a symlink in the target state.
type TargetStateSymlink struct {
	*lazyLinkname
}

// A targetStateRenameDir represents the renaming of a directory in the target
// state.
type targetStateRenameDir struct {
	oldRelPath RelPath
	newRelPath RelPath
}

// A scriptState records the state of a script that has been run.
type scriptState struct {
	Name  string    `json:"name" toml:"name" yaml:"name"`
	RunAt time.Time `json:"runAt" toml:"runAt" yaml:"runAt"`
}

// Apply updates actualStateEntry to match t. It does not recurse.
func (t *TargetStateDir) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
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
func (t *TargetStateDir) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateFile) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
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

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateFile) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply updates actualStateEntry to match t.
func (t *TargetStateRemove) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
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

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateRemove) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply runs t.
func (t *TargetStateScript) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
	contentsSHA256, err := t.ContentsSHA256()
	if err != nil {
		return false, err
	}
	key := []byte(hex.EncodeToString(contentsSHA256))
	if t.once {
		switch scriptState, err := persistentState.Get(scriptStateBucket, key); {
		case err != nil:
			return false, err
		case scriptState != nil:
			return false, nil
		}
	}
	contents, err := t.Contents()
	if err != nil {
		return false, err
	}
	runAt := time.Now().UTC()
	if !isEmpty(contents) {
		if err := system.RunScript(t.name, actualStateEntry.Path().Dir(), contents, t.interpreter); err != nil {
			return false, err
		}
	}
	return true, persistentStateSet(persistentState, scriptStateBucket, key, &scriptState{
		Name:  string(t.name),
		RunAt: runAt,
	})
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
func (t *TargetStateSymlink) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
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

// SkipApply implements TargetState.SkipApply.
func (t *TargetStateSymlink) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}

// Apply renames actualStateEntry.
func (t *targetStateRenameDir) Apply(system System, persistentState PersistentState, actualStateEntry ActualStateEntry) (bool, error) {
	dir := actualStateEntry.Path().Dir()
	return true, system.Rename(dir.Join(t.oldRelPath), dir.Join(t.newRelPath))
}

// EntryState returns t's entry state.
func (t *targetStateRenameDir) EntryState(umask fs.FileMode) (*EntryState, error) {
	return nil, nil
}

// Evaluate does nothing.
func (t *targetStateRenameDir) Evaluate() error {
	return nil
}

// SkipApply implements TargetState.SkipApply.
func (t *targetStateRenameDir) SkipApply(persistentState PersistentState) (bool, error) {
	return false, nil
}
