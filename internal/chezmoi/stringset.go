package chezmoi

import "strings"

// A stringSet is a set of strings.
type stringSet map[string]struct{}

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

// mergeWithPrefixAndStripComponents merges the elements of other into s with
// prefix and a number of components stripped.
func (s stringSet) mergeWithPrefixAndStripComponents(other stringSet, prefix string, stripComponents int) {
	for pattern := range other {
		components := strings.Split(pattern, "/")
		if len(components) <= stripComponents {
			continue
		}
		strippedComponentPattern := prefix + strings.Join(components[stripComponents:], "/")
		s[strippedComponentPattern] = struct{}{}
	}
}

// remove removes an element from s.
func (s stringSet) remove(element string) {
	delete(s, element)
}
