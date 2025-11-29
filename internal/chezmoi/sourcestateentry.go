package chezmoi

import (
	"encoding/hex"
	"log/slog"
	"os/exec"

	"chezmoi.io/chezmoi/internal/chezmoilog"
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
	IsExternal() bool
	Path() AbsPath
	OriginString() string
}

// A SourceStateOriginAbsPath is an absolute path.
type SourceStateOriginAbsPath AbsPath

// A SourceStateEntry represents the state of an entry in the source state.
type SourceStateEntry interface {
	slog.LogValuer
	Evaluate() error
	Order() ScriptOrder
	Origin() SourceStateOrigin
	SourceRelPath() SourceRelPath
	TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error)
}

// A SourceStateCommand represents a command that should be run.
type SourceStateCommand struct {
	cmdFunc       func() *exec.Cmd
	origin        SourceStateOrigin
	forceRefresh  bool
	refreshPeriod Duration
	sourceAttr    SourceAttr
}

// A SourceStateDir represents the state of a directory in the source state.
type SourceStateDir struct {
	attr             DirAttr
	origin           SourceStateOrigin
	sourceRelPath    SourceRelPath
	targetStateEntry TargetStateEntry
}

// A SourceStateFile represents the state of a file in the source state.
type SourceStateFile struct {
	attr                 FileAttr
	contentsFunc         func() ([]byte, error)
	contentsSHA256Func   func() ([32]byte, error)
	origin               SourceStateOrigin
	sourceRelPath        SourceRelPath
	targetStateEntryFunc TargetStateEntryFunc
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

// LogValue implements log/slog.LogValuer.LogValue.
func (s *SourceStateCommand) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("cmd", chezmoilog.OSExecCmdLogValuer{Cmd: s.cmdFunc()}),
		slog.String("origin", s.origin.OriginString()),
	)
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
func (s *SourceStateCommand) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &TargetStateModifyDirWithCmd{
		cmdFunc:       s.cmdFunc,
		forceRefresh:  s.forceRefresh,
		refreshPeriod: s.refreshPeriod,
		sourceAttr:    s.sourceAttr,
	}, nil
}

// Attr returns s's attributes.
func (s *SourceStateDir) Attr() DirAttr {
	return s.attr
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateDir) Evaluate() error {
	return nil
}

// LogValue implements log/slog.LogValuer.LogValue.
func (s *SourceStateDir) LogValue() slog.Value {
	return slog.GroupValue(
		chezmoilog.Stringer("sourceRelPath", s.sourceRelPath),
		slog.Any("attr", s.attr),
	)
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
func (s *SourceStateDir) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return s.targetStateEntry, nil
}

// Attr returns s's attributes.
func (s *SourceStateFile) Attr() FileAttr {
	return s.attr
}

// Contents returns s's contents.
func (s *SourceStateFile) Contents() ([]byte, error) {
	return s.contentsFunc()
}

// ContentsSHA256 returns the SHA256 sum of s's contents.
func (s *SourceStateFile) ContentsSHA256() ([32]byte, error) {
	return s.contentsSHA256Func()
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateFile) Evaluate() error {
	if _, err := s.Contents(); err != nil {
		return err
	}
	if _, err := s.ContentsSHA256(); err != nil {
		return err
	}
	return nil
}

// LogValue implements log/slog.LogValuer.LogValue.
func (s *SourceStateFile) LogValue() slog.Value {
	attrs := []slog.Attr{
		chezmoilog.Stringer("sourceRelPath", s.sourceRelPath),
		slog.Any("attr", s.attr),
	}
	contents, contentsErr := s.Contents()
	attrs = append(attrs, chezmoilog.FirstFewBytes("contents", contents))
	if contentsErr != nil {
		attrs = append(attrs, slog.Any("contentsErr", contentsErr))
	}
	contentsSHA256, contentsSHA256Err := s.ContentsSHA256()
	attrs = append(attrs, slog.String("contentsSHA256", hex.EncodeToString(contentsSHA256[:])))
	if contentsSHA256Err != nil {
		attrs = append(attrs, slog.Any("contentsSHA256Err", contentsSHA256Err))
	}
	return slog.GroupValue(attrs...)
}

// Order returns s's order.
func (s *SourceStateFile) Order() ScriptOrder {
	return s.attr.Order
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
func (s *SourceStateFile) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	if s.targetStateEntryFunc != nil {
		s.targetStateEntry, s.targetStateEntryErr = s.targetStateEntryFunc(destSystem, destDirAbsPath)
		s.targetStateEntryFunc = nil
	}
	return s.targetStateEntry, s.targetStateEntryErr
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateImplicitDir) Evaluate() error {
	return nil
}

// LogValue implements log/slog.LogValuer.LogValue.
func (s *SourceStateImplicitDir) LogValue() slog.Value {
	return slog.GroupValue()
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
func (s *SourceStateImplicitDir) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return s.targetStateEntry, nil
}

// Evaluate evaluates s and returns any error.
func (s *SourceStateRemove) Evaluate() error {
	return nil
}

// LogValue implements log/slog.LogValuer.LogValue.
func (s *SourceStateRemove) LogValue() slog.Value {
	return slog.GroupValue(
		chezmoilog.Stringer("targetRelPath", s.targetRelPath),
	)
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
func (s *SourceStateRemove) TargetStateEntry(destSystem System, destDirAbsPath AbsPath) (TargetStateEntry, error) {
	return &TargetStateRemove{}, nil
}

// IsExternal returns if s is an external.
func (SourceStateOriginAbsPath) IsExternal() bool {
	return false
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

// IsExternal returns if s is an external.
func (SourceStateOriginRemove) IsExternal() bool {
	return false
}
