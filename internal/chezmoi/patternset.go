package chezmoi

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rs/zerolog"
	vfs "github.com/twpayne/go-vfs/v4"
)

// An patternSet is a set of patterns.
type patternSet struct {
	includePatterns stringSet
	excludePatterns stringSet
}

// newPatternSet returns a new patternSet.
func newPatternSet() *patternSet {
	return &patternSet{
		includePatterns: newStringSet(),
		excludePatterns: newStringSet(),
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (ps *patternSet) MarshalZerologObject(e *zerolog.Event) {
	if ps == nil {
		return
	}
	e.Strs("includePatterns", ps.includePatterns.elements())
	e.Strs("excludePatterns", ps.excludePatterns.elements())
}

// add adds a pattern to ps.
func (ps *patternSet) add(pattern string, include bool) error {
	if ok := doublestar.ValidatePattern(pattern); !ok {
		return fmt.Errorf("%s: invalid pattern", pattern)
	}
	if include {
		ps.includePatterns.add(pattern)
	} else {
		ps.excludePatterns.add(pattern)
	}
	return nil
}

// glob returns all matches in fileSystem.
func (ps *patternSet) glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	// FIXME use AbsPath and RelPath
	allMatches := newStringSet()
	for includePattern := range ps.includePatterns {
		matches, err := doublestar.Glob(fileSystem, prefix+includePattern)
		if err != nil {
			return nil, err
		}
		allMatches.add(matches...)
	}
	for match := range allMatches {
		for excludePattern := range ps.excludePatterns {
			exclude, err := doublestar.Match(path.Clean(prefix+excludePattern), match)
			if err != nil {
				return nil, err
			}
			if exclude {
				allMatches.remove(match)
			}
		}
	}
	matchesSlice := allMatches.elements()
	for i, match := range matchesSlice {
		matchesSlice[i] = mustTrimPrefix(filepath.ToSlash(match), prefix)
	}
	sort.Strings(matchesSlice)
	return matchesSlice, nil
}

// match returns if name matches any pattern in ps.
func (ps *patternSet) match(name string) bool {
	for pattern := range ps.excludePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return false
		}
	}
	for pattern := range ps.includePatterns {
		if ok, _ := doublestar.Match(pattern, name); ok {
			return true
		}
	}
	return false
}

// mergeWithPrefixAndStripComponents merges the elements of other into ps with
// prefix and a number of components stripped.
func (ps *patternSet) mergeWithPrefixAndStripComponents(other *patternSet, prefix string, stripComponents int) {
	ps.includePatterns.mergeWithPrefixAndStripComponents(other.includePatterns, prefix, stripComponents)
	ps.excludePatterns.mergeWithPrefixAndStripComponents(other.excludePatterns, prefix, stripComponents)
}
