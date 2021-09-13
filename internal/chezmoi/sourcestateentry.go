package chezmoi

import (
	"encoding/hex"

	"github.com/rs/zerolog"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A SourceStateEntry represents the state of an entry in the source state.
type SourceStateEntry interface {
	zerolog.LogObjectMarshaler
	Evaluate() error
	Order() int
	SourceRelPath() SourceRelPath
	TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error)
}

// A SourceStateDir represents the state of a directory in the source state.
type SourceStateDir struct {
	Attr             DirAttr
	sourceRelPath    SourceRelPath
	targetStateEntry TargetStateEntry
}

// A SourceStateFile represents the state of a file in the source state.
type SourceStateFile struct {
	*lazyContents
	Attr                 FileAttr
	sourceRelPath        SourceRelPath
	targetStateEntryFunc targetStateEntryFunc
	targetStateEntry     TargetStateEntry
	targetStateEntryErr  error
}

// A SourceStateRemove represents that an entry should be removed.
type SourceStateRemove struct {
	targetRelPath RelPath
}

// A SourceStateRenameDir represents the renaming of a directory in the source
// state.
type SourceStateRenameDir struct {
	oldSourceRelPath SourceRelPath
	newSourceRelPath SourceRelPath
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateDir) Evaluate() error {
	return nil
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *SourceStateDir) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("sourceRelPath", s.sourceRelPath)
	e.Object("attr", s.Attr)
}

// Order returns s's order.
func (s *SourceStateDir) Order() int {
	return 0
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateDir) SourceRelPath() SourceRelPath {
	return s.sourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateDir) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return s.targetStateEntry, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateFile) Evaluate() error {
	_, err := s.ContentsSHA256()
	return err
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *SourceStateFile) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("sourceRelPath", s.sourceRelPath)
	e.Interface("attr", s.Attr)
	contents, contentsErr := s.Contents()
	e.Bytes("contents", chezmoilog.FirstFewBytes(contents))
	if contentsErr != nil {
		e.Str("contentsErr", contentsErr.Error())
	}
	e.Err(contentsErr)
	contentsSHA256, contentsSHA256Err := s.ContentsSHA256()
	e.Str("contentsSHA256", hex.EncodeToString(contentsSHA256))
	if contentsSHA256Err != nil {
		e.Str("contentsSHA256Err", contentsSHA256Err.Error())
	}
}

// Order returns s's order.
func (s *SourceStateFile) Order() int {
	return s.Attr.Order
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateFile) SourceRelPath() SourceRelPath {
	return s.sourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateFile) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	if s.targetStateEntryFunc != nil {
		s.targetStateEntry, s.targetStateEntryErr = s.targetStateEntryFunc(destSystem, destDirAbsPath)
		s.targetStateEntryFunc = nil
	}
	return s.targetStateEntry, s.targetStateEntryErr
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateRemove) Evaluate() error {
	return nil
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *SourceStateRemove) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("targetRelPath", s.targetRelPath)
}

// Order returns s's order.
func (s *SourceStateRemove) Order() int {
	return 0
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateRemove) SourceRelPath() SourceRelPath {
	return SourceRelPath{}
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateRemove) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &TargetStateRemove{}, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateRenameDir) Evaluate() error {
	return nil
}

// MarashalLogObject implements zerolog.LogObjectMarshaler.
func (s *SourceStateRenameDir) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("oldSourceRelPath", s.oldSourceRelPath)
	e.Stringer("newSourceRelPath", s.newSourceRelPath)
}

// Order returns s's order.
func (s *SourceStateRenameDir) Order() int {
	return -1
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateRenameDir) SourceRelPath() SourceRelPath {
	return s.newSourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateRenameDir) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &targetStateRenameDir{
		oldRelPath: s.oldSourceRelPath.RelPath(),
		newRelPath: s.newSourceRelPath.RelPath(),
	}, nil
}
