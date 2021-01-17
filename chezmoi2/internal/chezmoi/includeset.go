package chezmoi

// FIXME Add IncludeEncrypted

import (
	"fmt"
	"os"
	"strings"
)

// An IncludeSet controls what types of entries to include. It parses and prints
// as a comma-separated list of strings, but is internally represented as a
// bitmask. *IncludeSet implements the github.com/spf13/pflag.Value interface.
type IncludeSet struct {
	bits IncludeBits
}

// An IncludeBits is a bitmask of entries to include.
type IncludeBits int

// Include bits.
const (
	IncludeAbsent IncludeBits = 1 << iota
	IncludeDirs
	IncludeFiles
	IncludeScripts
	IncludeSymlinks

	// IncludeAll is all include bits.
	IncludeAll IncludeBits = IncludeAbsent | IncludeDirs | IncludeFiles | IncludeScripts | IncludeSymlinks

	// includeNone is no include bits.
	includeNone IncludeBits = 0
)

// includeBits is a map from human-readable strings to IncludeBits.
var includeBits = map[string]IncludeBits{
	"a":        IncludeAbsent,
	"absent":   IncludeAbsent,
	"all":      IncludeAll,
	"d":        IncludeDirs,
	"dirs":     IncludeDirs,
	"f":        IncludeFiles,
	"files":    IncludeFiles,
	"scripts":  IncludeScripts,
	"s":        IncludeSymlinks,
	"symlinks": IncludeSymlinks,
}

// NewIncludeSet returns a new IncludeSet.
func NewIncludeSet(bits IncludeBits) *IncludeSet {
	return &IncludeSet{
		bits: bits,
	}
}

// IncludeActualStateEntry returns true if actualStateEntry should be included.
func (s *IncludeSet) IncludeActualStateEntry(actualStateEntry ActualStateEntry) bool {
	switch actualStateEntry.(type) {
	case *ActualStateDir:
		return s.bits&IncludeDirs != 0
	case *ActualStateFile:
		return s.bits&IncludeFiles != 0
	case *ActualStateSymlink:
		return s.bits&IncludeSymlinks != 0
	default:
		return false
	}
}

// IncludeFileInfo returns true if info should be included.
func (s *IncludeSet) IncludeFileInfo(info os.FileInfo) bool {
	switch {
	case info.IsDir():
		return s.bits&IncludeDirs != 0
	case info.Mode().IsRegular():
		return s.bits&IncludeFiles != 0
	case info.Mode()&os.ModeType == os.ModeSymlink:
		return s.bits&IncludeSymlinks != 0
	default:
		return false
	}
}

// IncludeTargetStateEntry returns true if targetStateEntry should be included.
func (s *IncludeSet) IncludeTargetStateEntry(targetStateEntry TargetStateEntry) bool {
	switch targetStateEntry.(type) {
	case *TargetStateAbsent:
		return s.bits&IncludeAbsent != 0
	case *TargetStateDir:
		return s.bits&IncludeDirs != 0
	case *TargetStateFile:
		return s.bits&IncludeFiles != 0
	case *TargetStatePresent:
		return s.bits&IncludeFiles != 0
	case *TargetStateRenameDir:
		return s.bits&IncludeDirs != 0
	case *TargetStateScript:
		return s.bits&IncludeScripts != 0
	case *TargetStateSymlink:
		return s.bits&IncludeSymlinks != 0
	default:
		return false
	}
}

// Set implements github.com/spf13/pflag.Value.Set.
func (s *IncludeSet) Set(str string) error {
	if str == "none" {
		s.bits = includeNone
		return nil
	}

	var bits IncludeBits
	for _, element := range strings.Split(str, ",") {
		if element == "" {
			continue
		}
		exclude := false
		if strings.HasPrefix(element, "!") {
			exclude = true
			element = element[1:]
		}
		bit, ok := includeBits[element]
		if !ok {
			return fmt.Errorf("%s: unknown include element", element)
		}
		if exclude {
			bits &^= bit
		} else {
			bits |= bit
		}
	}
	s.bits = bits
	return nil
}

func (s *IncludeSet) String() string {
	//nolint:exhaustive
	switch s.bits {
	case IncludeAll:
		return "all"
	case includeNone:
		return "none"
	}
	var elements []string
	for i, element := range []string{
		"absent",
		"dirs",
		"files",
		"scripts",
		"symlinks",
	} {
		if s.bits&(1<<i) != 0 {
			elements = append(elements, element)
		}
	}
	return strings.Join(elements, ",")
}

// Type implements github.com/spf13/pflag.Value.Type.
func (s *IncludeSet) Type() string {
	return "include set"
}
