package chezmoi

import (
	"encoding/hex"
	"os/exec"

	"github.com/rs/zerolog"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoilog"
)

// A SourceStateEntry represents the state of an entry in the source state.
type SourceStateEntry interface {
	zerolog.LogObjectMarshaler
	Evaluate() error
	External() bool
	Order() ScriptOrder
	Origin() string
	SourceRelPath() SourceRelPath
	TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error)
}

// A SourceStateCommand represents a command that should be run.
type SourceStateCommand struct {
	cmd           *exec.Cmd
	external      bool
	origin        string
	forceRefresh  bool
	refreshPeriod Duration
}

// A SourceStateDir represents the state of a directory in the source state.
type SourceStateDir struct {
	Attr             DirAttr
	external         bool
	origin           string
	sourceRelPath    SourceRelPath
	targetStateEntry TargetStateEntry
}

// A SourceStateFile represents the state of a file in the source state.
type SourceStateFile struct {
	*lazyContents
	Attr                 FileAttr
	external             bool
	origin               string
	sourceRelPath        SourceRelPath
	targetStateEntryFunc targetStateEntryFunc
	targetStateEntry     TargetStateEntry
	targetStateEntryErr  error
}

// A SourceStateRemove represents that an entry should be removed.
type SourceStateRemove struct {
	sourceRelPath SourceRelPath
	targetRelPath RelPath
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateCommand) Evaluate() error {
	return nil
}

// External returns if s is from an external.
func (s *SourceStateCommand) External() bool {
	return s.external
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (s *SourceStateCommand) MarshalZerologObject(e *zerolog.Event) {
	e.EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: s.cmd})
	e.Str("origin", s.origin)
}

// Order returns s's order.
func (s *SourceStateCommand) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateCommand) Origin() string {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateCommand) SourceRelPath() SourceRelPath {
	return emptySourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateCommand) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &TargetStateModifyDirWithCmd{
		cmd:           s.cmd,
		forceRefresh:  s.forceRefresh,
		refreshPeriod: s.refreshPeriod,
	}, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateDir) Evaluate() error {
	return nil
}

// External returns if s is from an external.
func (s *SourceStateDir) External() bool {
	return s.external
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (s *SourceStateDir) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("sourceRelPath", s.sourceRelPath)
	e.Object("attr", s.Attr)
}

// Order returns s's order.
func (s *SourceStateDir) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateDir) Origin() string {
	return s.origin
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

// External returns if s is from an external.
func (s *SourceStateFile) External() bool {
	return s.external
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
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
func (s *SourceStateFile) Order() ScriptOrder {
	return s.Attr.Order
}

// Origin returns s's origin.
func (s *SourceStateFile) Origin() string {
	return s.origin
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

// External returns if s is from an external.
func (s *SourceStateRemove) External() bool {
	return false
}

// MarshalZerologObject implements zerolog.LogObjectMarshaler.
func (s *SourceStateRemove) MarshalZerologObject(e *zerolog.Event) {
	e.Stringer("targetRelPath", s.targetRelPath)
}

// Order returns s's order.
func (s *SourceStateRemove) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateRemove) Origin() string {
	return s.sourceRelPath.String()
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateRemove) SourceRelPath() SourceRelPath {
	return SourceRelPath{}
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateRemove) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &TargetStateRemove{}, nil
}
