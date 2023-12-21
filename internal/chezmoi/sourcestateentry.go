package chezmoi

import (
	"encoding/hex"
	"os/exec"

	"github.com/rs/zerolog"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A SourceAttr contains attributes of the source.
type SourceAttr struct {
	Condition ScriptCondition
	Encrypted bool
	External  bool
	Template  bool
}

// A SourceStateOrigin represents the origin of a source state.
type SourceStateOrigin interface {
	Path() AbsPath
	OriginString() string
}

// A SourceStateOriginAbsPath is an absolute path.
type SourceStateOriginAbsPath AbsPath

// A SourceStateEntry represents the state of an entry in the source state.
type SourceStateEntry interface {
	zerolog.LogObjectMarshaler
	Evaluate() error
	Order() ScriptOrder
	Origin() SourceStateOrigin
	SourceRelPath() SourceRelPath
	TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error)
}

// A SourceStateCommand represents a command that should be run.
type SourceStateCommand struct {
	cmd           *exec.Cmd
	origin        SourceStateOrigin
	forceRefresh  bool
	refreshPeriod Duration
	sourceAttr    SourceAttr
}

// A SourceStateDir represents the state of a directory in the source state.
type SourceStateDir struct {
	Attr             DirAttr
	origin           SourceStateOrigin
	sourceRelPath    SourceRelPath
	targetStateEntry TargetStateEntry
}

// A SourceStateFile represents the state of a file in the source state.
type SourceStateFile struct {
	*lazyContents
	Attr                 FileAttr
	origin               SourceStateOrigin
	sourceRelPath        SourceRelPath
	targetStateEntryFunc targetStateEntryFunc
	targetStateEntry     TargetStateEntry
	targetStateEntryErr  error
}

// A SourceStateImplicitDir represents the state of a directory that is implicit
// in the source state, typically because it is a parent directory of an
// external. Implicit directories have no attributes and are considered
// equivalent to any other directory.
type SourceStateImplicitDir struct {
	origin           SourceStateOrigin
	targetStateEntry TargetStateEntry
}

// A SourceStateRemove represents that an entry should be removed.
type SourceStateRemove struct {
	origin        SourceStateOrigin
	sourceRelPath SourceRelPath
	targetRelPath RelPath
}

// A SourceStateOriginRemove is used for removes. The source of the remove is
// not currently tracked. The remove could come from an exact_ directory, a
// non-empty_ file with empty contents, or one of many patterns in many
// .chezmoiignore files.
//
// FIXME remove this when the sources of all removes are tracked.
type SourceStateOriginRemove struct{}

// Evaluate evaluates s and returns any error.
func (s *SourceStateCommand) Evaluate() error {
	return nil
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (s *SourceStateCommand) MarshalZerologObject(e *zerolog.Event) {
	e.EmbedObject(chezmoilog.OSExecCmdLogObject{Cmd: s.cmd})
	e.Str("origin", s.origin.OriginString())
}

// Order returns s's order.
func (s *SourceStateCommand) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateCommand) Origin() SourceStateOrigin {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateCommand) SourceRelPath() SourceRelPath {
	return emptySourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateCommand) TargetStateEntry(
	destSystem System,
	destDirAbsPath AbsPath,
) (TargetStateEntry, error) {
	return &TargetStateModifyDirWithCmd{
		cmd:           s.cmd,
		forceRefresh:  s.forceRefresh,
		refreshPeriod: s.refreshPeriod,
		sourceAttr:    s.sourceAttr,
	}, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateDir) Evaluate() error {
	return nil
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
func (s *SourceStateDir) Origin() SourceStateOrigin {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateDir) SourceRelPath() SourceRelPath {
	return s.sourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateDir) TargetStateEntry(
	destSystem System,
	destDirAbsPath AbsPath,
) (TargetStateEntry, error) {
	return s.targetStateEntry, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateFile) Evaluate() error {
	_, err := s.ContentsSHA256()
	return err
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
func (s *SourceStateFile) Origin() SourceStateOrigin {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateFile) SourceRelPath() SourceRelPath {
	return s.sourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateFile) TargetStateEntry(
	destSystem System,
	destDirAbsPath AbsPath,
) (TargetStateEntry, error) {
	if s.targetStateEntryFunc != nil {
		s.targetStateEntry, s.targetStateEntryErr = s.targetStateEntryFunc(
			destSystem,
			destDirAbsPath,
		)
		s.targetStateEntryFunc = nil
	}
	return s.targetStateEntry, s.targetStateEntryErr
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateImplicitDir) Evaluate() error {
	return nil
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (s *SourceStateImplicitDir) MarshalZerologObject(e *zerolog.Event) {
}

// Order returns s's order.
func (s *SourceStateImplicitDir) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateImplicitDir) Origin() SourceStateOrigin {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateImplicitDir) SourceRelPath() SourceRelPath {
	return emptySourceRelPath
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateImplicitDir) TargetStateEntry(
	destSystem System,
	destDirAbsPath AbsPath,
) (TargetStateEntry, error) {
	return s.targetStateEntry, nil
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
func (s *SourceStateRemove) Order() ScriptOrder {
	return ScriptOrderDuring
}

// Origin returns s's origin.
func (s *SourceStateRemove) Origin() SourceStateOrigin {
	return s.origin
}

// SourceRelPath returns s's source relative path.
func (s *SourceStateRemove) SourceRelPath() SourceRelPath {
	return SourceRelPath{}
}

// TargetStateEntry returns s's target state entry.
func (s *SourceStateRemove) TargetStateEntry(
	destSystem System,
	destDirAbsPath AbsPath,
) (TargetStateEntry, error) {
	return &TargetStateRemove{}, nil
}

// Path returns s's path.
func (s SourceStateOriginAbsPath) Path() AbsPath {
	return AbsPath(s)
}

// OriginString returns s's origin.
func (s SourceStateOriginAbsPath) OriginString() string {
	return AbsPath(s).String()
}

// Path returns s's path.
func (s SourceStateOriginRemove) Path() AbsPath {
	return EmptyAbsPath
}

// OriginString returns s's origin.
func (s SourceStateOriginRemove) OriginString() string {
	return "remove"
}
