package chezmoi

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

// remove removes an element from s.
func (s stringSet) remove(element string) {
	delete(s, element)
}
