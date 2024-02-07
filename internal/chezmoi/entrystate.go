package chezmoi

import (
	"bytes"
	"io/fs"
	"log/slog"
	"runtime"

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

// LogValue implements log/slog.LogValuer.LogValue.
func (s *EntryState) LogValue() slog.Value {
	if s == nil {
		return slog.Value{}
	}
	attrs := []slog.Attr{
		slog.String("Type", string(s.Type)),
		slog.Int("Mode", int(s.Mode)),
		chezmoilog.Stringer("ContentsSHA256", s.ContentsSHA256),
	}
	if len(s.contents) != 0 {
		attrs = append(attrs, chezmoilog.FirstFewBytes("contents", s.contents))
	}
	if s.overwrite {
		attrs = append(attrs, slog.Bool("overwrite", s.overwrite))
	}
	return slog.GroupValue(attrs...)
}

// Overwrite returns true if s should be overwritten by default.
func (s *EntryState) Overwrite() bool {
	return s.overwrite
}
