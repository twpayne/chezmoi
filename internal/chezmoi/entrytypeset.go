package chezmoi

import (
	"fmt"
	"io/fs"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

// An EntryTypeSet is a set of entry types. It parses and prints as a
// comma-separated list of strings, but is internally represented as a bitmask.
// *EntryTypeSet implements the github.com/spf13/pflag.Value interface.
type EntryTypeSet struct {
	bits EntryTypeBits
}

// An EntryTypeBits is a bitmask of entry types.
type EntryTypeBits int

// Entry type bits.
const (
	EntryTypeDirs EntryTypeBits = 1 << iota
	EntryTypeFiles
	EntryTypeRemove
	EntryTypeScripts
	EntryTypeSymlinks
	EntryTypeEncrypted
	EntryTypeExternals
	EntryTypeTemplates
	EntryTypeAlways

	// EntryTypesAll is all entry types.
	EntryTypesAll EntryTypeBits = EntryTypeDirs |
		EntryTypeFiles |
		EntryTypeRemove |
		EntryTypeScripts |
		EntryTypeSymlinks |
		EntryTypeEncrypted |
		EntryTypeExternals |
		EntryTypeTemplates |
		EntryTypeAlways

	// EntryTypesNone is no entry types.
	EntryTypesNone EntryTypeBits = 0
)

var (
	// entryTypeBits is a map from human-readable strings to EntryTypeBits.
	entryTypeBits = map[string]EntryTypeBits{
		"all":       EntryTypesAll,
		"always":    EntryTypeAlways,
		"dirs":      EntryTypeDirs,
		"files":     EntryTypeFiles,
		"remove":    EntryTypeRemove,
		"scripts":   EntryTypeScripts,
		"symlinks":  EntryTypeSymlinks,
		"encrypted": EntryTypeEncrypted,
		"externals": EntryTypeExternals,
		"templates": EntryTypeTemplates,
	}

	entryTypeStrings = slices.Sorted(maps.Keys(entryTypeBits))

	entryTypeCompletions = []string{
		"all",
		"always",
		"dirs",
		"encrypted",
		"externals",
		"files",
		"noalways",
		"nodirs",
		"noencrypted",
		"noexternals",
		"nofiles",
		"none",
		"noremove",
		"noscripts",
		"nosymlinks",
		"notemplates",
		"remove",
		"scripts",
		"symlinks",
		"templates",
	}
)

// NewEntryTypeSet returns a new IncludeSet.
func NewEntryTypeSet(bits EntryTypeBits) *EntryTypeSet {
	return &EntryTypeSet{
		bits: bits,
	}
}

// Bits returns s's bits.
func (s *EntryTypeSet) Bits() EntryTypeBits {
	return s.bits
}

// ContainsEntryTypeBits returns if s includes b.
func (s *EntryTypeSet) ContainsEntryTypeBits(b EntryTypeBits) bool {
	return s.bits&b != 0
}

// ContainsFileInfo returns true if fileInfo is a member.
func (s *EntryTypeSet) ContainsFileInfo(fileInfo fs.FileInfo) bool {
	switch {
	case fileInfo.IsDir():
		return s.bits&EntryTypeDirs != 0
	case fileInfo.Mode().IsRegular():
		return s.bits&EntryTypeFiles != 0
	case fileInfo.Mode().Type() == fs.ModeSymlink:
		return s.bits&EntryTypeSymlinks != 0
	default:
		return false
	}
}

// ContainsSourceStateEntry returns true if sourceStateEntry is a member.
func (s *EntryTypeSet) ContainsSourceStateEntry(sourceStateEntry SourceStateEntry) bool {
	_, isExternal := sourceStateEntry.Origin().(*External)
	switch sourceStateEntry := sourceStateEntry.(type) {
	case *SourceStateCommand:
		switch {
		case s.bits&EntryTypeExternals != 0 && isExternal:
			return true
		case s.bits&EntryTypeDirs != 0:
			return true
		default:
			return false
		}
	case *SourceStateDir, *SourceStateImplicitDir:
		switch {
		case s.bits&EntryTypeExternals != 0 && isExternal:
			return true
		case s.bits&EntryTypeDirs != 0:
			return true
		default:
			return false
		}
	case *SourceStateFile:
		switch sourceAttr := sourceStateEntry.Attr; {
		case s.bits&EntryTypeExternals != 0 && isExternal:
			return true
		case s.bits&EntryTypeEncrypted != 0 && sourceAttr.Encrypted:
			return true
		case s.bits&EntryTypeTemplates != 0 && sourceAttr.Template:
			return true
		case s.bits&EntryTypeFiles != 0 && sourceAttr.Type == SourceFileTypeCreate:
			return true
		case s.bits&EntryTypeFiles != 0 && sourceAttr.Type == SourceFileTypeFile:
			return true
		case s.bits&EntryTypeFiles != 0 && sourceAttr.Type == SourceFileTypeModify:
			return true
		case s.bits&EntryTypeRemove != 0 && sourceAttr.Type == SourceFileTypeRemove:
			return true
		case s.bits&EntryTypeScripts != 0 && sourceAttr.Type == SourceFileTypeScript:
			return true
		case s.bits&EntryTypeSymlinks != 0 && sourceAttr.Type == SourceFileTypeSymlink:
			return true
		case s.bits&EntryTypeAlways != 0 && sourceAttr.Condition == ScriptConditionAlways:
			return true
		default:
			return false
		}
	case *SourceStateRemove:
		switch {
		case s.bits&EntryTypeExternals != 0 && isExternal:
			return true
		case s.bits&EntryTypeRemove != 0:
			return true
		default:
			return false
		}
	default:
		panic(fmt.Sprintf("%T: unsupported type", sourceStateEntry))
	}
}

// ContainsTargetStateEntry returns true if targetStateEntry is a member.
func (s *EntryTypeSet) ContainsTargetStateEntry(targetStateEntry TargetStateEntry) bool {
	sourceAttr := targetStateEntry.SourceAttr()
	switch targetStateEntry.(type) {
	case *TargetStateDir:
		switch {
		case s.bits&EntryTypeExternals != 0 && sourceAttr.External:
			return true
		case s.bits&EntryTypeDirs != 0:
			return true
		default:
			return false
		}
	case *TargetStateFile:
		switch {
		case s.bits&EntryTypeEncrypted != 0 && sourceAttr.Encrypted:
			return true
		case s.bits&EntryTypeExternals != 0 && sourceAttr.External:
			return true
		case s.bits&EntryTypeTemplates != 0 && sourceAttr.Template:
			return true
		case s.bits&EntryTypeFiles != 0:
			return true
		default:
			return false
		}
	case *TargetStateModifyDirWithCmd:
		switch {
		case s.bits&EntryTypeExternals != 0 && sourceAttr.External:
			return true
		case s.bits&EntryTypeDirs != 0:
			return true
		default:
			return false
		}
	case *TargetStateRemove:
		return s.bits&EntryTypeRemove != 0
	case *TargetStateScript:
		switch {
		case s.bits&EntryTypeEncrypted != 0 && sourceAttr.Encrypted:
			return true
		case s.bits&EntryTypeTemplates != 0 && sourceAttr.Template:
			return true
		case s.bits&EntryTypeAlways != 0 && sourceAttr.Condition == ScriptConditionAlways:
			return true
		case s.bits&EntryTypeScripts != 0:
			return true
		default:
			return false
		}
	case *TargetStateSymlink:
		switch {
		case s.bits&EntryTypeEncrypted != 0 && sourceAttr.Encrypted:
			return true
		case s.bits&EntryTypeExternals != 0 && sourceAttr.External:
			return true
		case s.bits&EntryTypeTemplates != 0 && sourceAttr.Template:
			return true
		case s.bits&EntryTypeSymlinks != 0:
			return true
		default:
			return false
		}
	default:
		panic(fmt.Sprintf("%T: unsupported type", targetStateEntry))
	}
}

// MarshalJSON implements encoding/json.Marshaler.MarshalJSON.
func (s *EntryTypeSet) MarshalJSON() ([]byte, error) {
	switch s.bits {
	case EntryTypesAll:
		return []byte(`["all"]`), nil
	case EntryTypesNone:
		return []byte("[]"), nil
	default:
		var elements []string
		for _, key := range entryTypeStrings {
			if bit := entryTypeBits[key]; s.bits&bit == bit {
				elements = append(elements, `"`+key+`"`)
			}
		}
		return []byte("[" + strings.Join(elements, ",") + "]"), nil
	}
}

// MarshalYAML implements gopkg.in/yaml.v3.Marshaler.
func (s *EntryTypeSet) MarshalYAML() (any, error) {
	if s.bits == EntryTypesAll {
		return []string{"all"}, nil
	}
	var result []string
	for _, key := range entryTypeStrings {
		if bit := entryTypeBits[key]; s.bits&bit == bit {
			result = append(result, key)
		}
	}
	return result, nil
}

// Set implements github.com/spf13/pflag.Value.Set.
func (s *EntryTypeSet) Set(str string) error {
	if str == "none" {
		s.bits = EntryTypesNone
		return nil
	}
	return s.SetSlice(strings.Split(str, ","))
}

// SetSlice sets s from a []string.
func (s *EntryTypeSet) SetSlice(ss []string) error {
	bits := EntryTypesNone
	for i, element := range ss {
		if element == "" {
			continue
		}
		element, exclude := strings.CutPrefix(element, "no")
		bit, ok := entryTypeBits[element]
		if !ok {
			return fmt.Errorf("%s: unknown entry type", element)
		}
		if i == 0 && exclude {
			bits = EntryTypesAll
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

// String implements github.com/spf13/pflag.Value.String.
func (s *EntryTypeSet) String() string {
	if s == nil {
		return "none"
	}
	switch s.bits {
	case EntryTypesAll:
		return "all"
	case EntryTypesNone:
		return "none"
	}
	var entryTypeStrs []string
	for _, entryTypeStr := range entryTypeStrings {
		bits := entryTypeBits[entryTypeStr]
		if s.bits&bits == bits {
			entryTypeStrs = append(entryTypeStrs, entryTypeStr)
		}
	}
	return strings.Join(entryTypeStrs, ",")
}

// Type implements github.com/spf13/pflag.Value.Type.
func (s *EntryTypeSet) Type() string {
	return "types"
}

// StringSliceToEntryTypeSetHookFunc is a
// github.com/mitchellh/mapstructure.DecodeHookFunc that parses an EntryTypeSet
// from a []string.
func StringSliceToEntryTypeSetHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if to != reflect.TypeOf(EntryTypeSet{}) {
			return data, nil
		}
		elemsAny, ok := data.([]any)
		if !ok {
			return nil, fmt.Errorf("expected a []string, got a %T", data)
		}
		elemStrs := make([]string, len(elemsAny))
		for i, elemAny := range elemsAny {
			elemStr, ok := elemAny.(string)
			if !ok {
				return nil, fmt.Errorf("expected a []string, got a %T element", elemAny)
			}
			elemStrs[i] = elemStr
		}
		s := NewEntryTypeSet(EntryTypesNone)
		if err := s.SetSlice(elemStrs); err != nil {
			return nil, err
		}
		return s, nil
	}
}

// EntryTypeSetFlagCompletionFunc completes EntryTypeSet flags.
func EntryTypeSetFlagCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var completions []string
	entryTypes := strings.Split(toComplete, ",")
	lastEntryType := entryTypes[len(entryTypes)-1]
	var prefix string
	if len(entryTypes) > 0 {
		prefix = toComplete[:len(toComplete)-len(lastEntryType)]
	}
	for _, completion := range entryTypeCompletions {
		if strings.HasPrefix(completion, lastEntryType) {
			completions = append(completions, prefix+completion)
		}
	}
	return completions, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}
