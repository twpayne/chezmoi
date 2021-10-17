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

// A StringSet is a set of strings.
type StringSet map[string]struct{}

// An patternSet is a set of patterns.
type patternSet struct {
	includePatterns StringSet
	excludePatterns StringSet
}

// newPatternSet returns a new patternSet.
func newPatternSet() *patternSet {
	return &patternSet{
		includePatterns: NewStringSet(),
		excludePatterns: NewStringSet(),
	}
}

// MarshalZerologObject implements
// github.com/rs/zerolog.LogObjectMarshaler.MarshalZerologObject.
func (ps *patternSet) MarshalZerologObject(e *zerolog.Event) {
	if ps == nil {
		return
	}
	e.Strs("includePatterns", ps.includePatterns.Elements())
	e.Strs("excludePatterns", ps.excludePatterns.Elements())
}

// add adds a pattern to ps.
func (ps *patternSet) add(pattern string, include bool) error {
	if ok := doublestar.ValidatePattern(pattern); !ok {
		return fmt.Errorf("%s: invalid pattern", pattern)
	}
	if include {
		ps.includePatterns.Add(pattern)
	} else {
		ps.excludePatterns.Add(pattern)
	}
	return nil
}

// glob returns all matches in fileSystem.
func (ps *patternSet) glob(fileSystem vfs.FS, prefix string) ([]string, error) {
	// FIXME use AbsPath and RelPath
	allMatches := NewStringSet()
	for includePattern := range ps.includePatterns {
		matches, err := doublestar.Glob(fileSystem, prefix+includePattern)
		if err != nil {
			return nil, err
		}
		allMatches.Add(matches...)
	}
	for match := range allMatches {
		for excludePattern := range ps.excludePatterns {
			exclude, err := doublestar.Match(path.Clean(prefix+excludePattern), match)
			if err != nil {
				return nil, err
			}
			if exclude {
				allMatches.Delete(match)
			}
		}
	}
	matchesSlice := allMatches.Elements()
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

// NewStringSet returns a new StringSet containing elements.
func NewStringSet(elements ...string) StringSet {
	s := make(StringSet)
	s.Add(elements...)
	return s
}

// Add adds elements to s.
func (s StringSet) Add(elements ...string) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// Contains returns true if s Contains element.
func (s StringSet) Contains(element string) bool {
	_, ok := s[element]
	return ok
}

// Delete deletes element from s.
func (s StringSet) Delete(element string) {
	delete(s, element)
}

// Element returns an arbitrary element from s or the empty string if s is
// empty.
func (s StringSet) Element() string {
	for element := range s {
		return element
	}
	return ""
}

// Elements returns all the Elements of s.
func (s StringSet) Elements() []string {
	elements := make([]string, 0, len(s))
	for element := range s {
		elements = append(elements, element)
	}
	return elements
}
