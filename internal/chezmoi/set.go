package chezmoi

// A set is a set of elements.
type set[T comparable] map[T]struct{}

// newSet returns a new StringSet containing elements.
func newSet[T comparable](elements ...T) set[T] {
	s := make(set[T])
	s.add(elements...)
	return s
}

// add adds elements to s.
func (s set[T]) add(elements ...T) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// contains returns true if s contains element.
func (s set[T]) contains(element T) bool {
	_, ok := s[element]
	return ok
}

// elements returns all the elements of s.
func (s set[T]) elements() []T {
	elements := make([]T, 0, len(s))
	for element := range s {
		elements = append(elements, element)
	}
	return elements
}

// remove removes an element from s.
func (s set[T]) remove(element T) {
	delete(s, element)
}
