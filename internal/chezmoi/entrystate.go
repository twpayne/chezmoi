package chezmoi

import (
	"bytes"
	"io/fs"
	"runtime"

	"github.com/rs/zerolog"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// An EntryStateType is an entry state type.
type EntryStateType string

// Entry state types.
const (
	EntryStateTypeDir     EntryStateType = "dir"
	EntryStateTypeFile    EntryStateType = "file"
	EntryStateTypeSymlink EntryStateType = "symlink"
	EntryStateTypeRemove  EntryStateType = "remove"
	EntryStateTypeScript  EntryStateType = "script"
)

// An EntryState represents the state of an entry. A nil EntryState is
// equivalent to EntryStateTypeAbsent.
type EntryState struct {
	Type           EntryStateType `json:"type"                     yaml:"type"`
	Mode           fs.FileMode    `json:"mode,omitempty"           yaml:"mode,omitempty"`
	ContentsSHA256 HexBytes       `json:"contentsSHA256,omitempty" yaml:"contentsSHA256,omitempty"` //nolint:tagliatelle
	contents       []byte
	overwrite      bool
}

// Contents returns s's contents, if available.
func (s *EntryState) Contents() []byte {
	return s.contents
}

// Equal returns true if s is equal to other.
func (s *EntryState) Equal(other *EntryState) bool {
	if s.Type != other.Type {
		return false
	}
	if runtime.GOOS != "windows" && s.Mode.Perm() != other.Mode.Perm() {
		return false
	}
	return bytes.Equal(s.ContentsSHA256, other.ContentsSHA256)
}

// Equivalent returns true if s is equivalent to other.
func (s *EntryState) Equivalent(other *EntryState) bool {
	switch {
	case s == nil:
		return other == nil || other.Type == EntryStateTypeRemove
	case other == nil:
		return s.Type == EntryStateTypeRemove
	default:
		return s.Equal(other)
	}
}

// Overwrite returns true if s should be overwritten by default.
func (s *EntryState) Overwrite() bool {
	return s.overwrite
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (s *EntryState) MarshalZerologObject(e *zerolog.Event) {
	if s == nil {
		return
	}
	e.Str("Type", string(s.Type))
	e.Int("Mode", int(s.Mode))
	e.Stringer("ContentsSHA256", s.ContentsSHA256)
	if len(s.contents) != 0 {
		e.Bytes("contents", chezmoilog.FirstFewBytes(s.contents))
	}
	if s.overwrite {
		e.Bool("overwrite", s.overwrite)
	}
}
