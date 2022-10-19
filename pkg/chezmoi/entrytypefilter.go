package chezmoi

import "io/fs"

// An EntryTypeFilter filters entries by type and source attributes. Any entry
// in the include set is included, otherwise if the entry is in the exclude set
// then it is excluded, otherwise it is included.
type EntryTypeFilter struct {
	Include *EntryTypeSet
	Exclude *EntryTypeSet
}

// NewEntryTypeFilter returns a new EntryTypeFilter with the given entry type
// bits.
func NewEntryTypeFilter(includeEntryTypeBits, excludeEntryTypeBits EntryTypeBits) *EntryTypeFilter {
	return &EntryTypeFilter{
		Include: NewEntryTypeSet(includeEntryTypeBits),
		Exclude: NewEntryTypeSet(excludeEntryTypeBits),
	}
}

// IncludeEntryTypes returns if entryTypeBits is included.
func (f *EntryTypeFilter) IncludeEntryTypeBits(entryTypeBits EntryTypeBits) bool {
	return f.Include.ContainsEntryTypeBits(entryTypeBits) && !f.Exclude.ContainsEntryTypeBits(entryTypeBits)
}

// IncludeFileInfo returns if fileInfo is included.
func (f *EntryTypeFilter) IncludeFileInfo(fileInfo fs.FileInfo) bool {
	return f.Include.ContainsFileInfo(fileInfo) && !f.Exclude.ContainsFileInfo(fileInfo)
}

// IncludeSourceStateEntry returns if sourceStateEntry is included.
func (f *EntryTypeFilter) IncludeSourceStateEntry(sourceStateEntry SourceStateEntry) bool {
	return f.Include.ContainsSourceStateEntry(sourceStateEntry) && !f.Exclude.ContainsSourceStateEntry(sourceStateEntry)
}

// IncludeTargetStateEntry returns if targetStateEntry is included.
func (f *EntryTypeFilter) IncludeTargetStateEntry(targetStateEntry TargetStateEntry) bool {
	return f.Include.ContainsTargetStateEntry(targetStateEntry) && !f.Exclude.ContainsTargetStateEntry(targetStateEntry)
}
