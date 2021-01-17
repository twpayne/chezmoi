package chezmoi

import (
	"bytes"
	"os"
)

// An EntryStateType is an entry state type.
type EntryStateType string

// Entry state types.
const (
	EntryStateTypeAbsent  EntryStateType = "absent"
	EntryStateTypePresent EntryStateType = "present"
	EntryStateTypeDir     EntryStateType = "dir"
	EntryStateTypeFile    EntryStateType = "file"
	EntryStateTypeSymlink EntryStateType = "symlink"
	EntryStateTypeScript  EntryStateType = "script"
)

// An EntryState represents the state of an entry. A nil EntryState is
// equivalent to EntryStateTypeAbsent.
type EntryState struct {
	Type           EntryStateType `json:"type" toml:"type" yaml:"type"`
	Mode           os.FileMode    `json:"mode,omitempty" toml:"mode,omitempty" yaml:"mode,omitempty"`
	ContentsSHA256 hexBytes       `json:"contentsSHA256,omitempty" toml:"contentsSHA256,omitempty" yaml:"contentsSHA256,omitempty"`
}

// Equal returns true if s is equal to other.
func (s *EntryState) Equal(other *EntryState, umask os.FileMode) bool {
	return s.Type == other.Type &&
		s.Mode&^umask == other.Mode&^umask &&
		bytes.Equal(s.ContentsSHA256, other.ContentsSHA256)
}

// Equivalent returns true if s is equivalent to other.
func (s *EntryState) Equivalent(other *EntryState, umask os.FileMode) bool {
	switch {
	case s == nil:
		return other == nil || other.Type == EntryStateTypeAbsent
	case other == nil:
		return s.Type == EntryStateTypeAbsent
	case s.Type == EntryStateTypeFile:
		return other.Type == EntryStateTypePresent || s.Equal(other, umask)
	case s.Type == EntryStateTypePresent:
		return other.Type == EntryStateTypeFile || other.Type == EntryStateTypePresent
	default:
		return s.Equal(other, umask)
	}
}
