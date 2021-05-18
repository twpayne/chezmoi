package chezmoi

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	vfs "github.com/twpayne/go-vfs/v3"
)

// A stringSet is a set of strings.
type stringSet map[string]struct{}

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
			exclude, err := doublestar.Match(prefix+excludePattern, match)
			if err != nil {
				return nil, err
			}
			if exclude {
				delete(allMatches, match)
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

// newStringSet returns a new StringSet containing elements.
func newStringSet(elements ...string) stringSet {
	s := make(stringSet)
	s.add(elements...)
	return s
}

// add adds elements to s.
func (s stringSet) add(elements ...string) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// contains returns true if s contains element.
func (s stringSet) contains(element string) bool {
	_, ok := s[element]
	return ok
}

// elements returns all the elements of s.
func (s stringSet) elements() []string {
	elements := make([]string, 0, len(s))
	for element := range s {
		elements = append(elements, element)
	}
	return elements
}
