// Package chezmoiset implements a generic set type.
package chezmoiset

// A Set is a set of elements.
type Set[T comparable] map[T]struct{}

// New returns a new set containing elements.
func New[T comparable](elements ...T) Set[T] {
	s := make(Set[T])
	s.Add(elements...)
	return s
}

// NewWithCapacity returns a new empty set with the given capacity.
func NewWithCapacity[T comparable](capacity int) Set[T] {
	return make(Set[T], capacity)
}

// Add adds elements to s.
func (s Set[T]) Add(elements ...T) {
	for _, element := range elements {
		s[element] = struct{}{}
	}
}

// AddSet adds all elements from other to s.
func (s Set[T]) AddSet(other Set[T]) {
	for element := range other {
		s[element] = struct{}{}
	}
}

// AnyElement returns an arbitrary element from s. It is typically used when s
// is known to contain exactly one element.
func (s Set[T]) AnyElement() T {
	for element := range s {
		return element
	}
	var zero T
	return zero
}

// Contains returns true if s contains element.
func (s Set[T]) Contains(element T) bool {
	_, ok := s[element]
	return ok
}

// Elements returns all the elements of s.
func (s Set[T]) Elements() []T {
	elements := make([]T, 0, len(s))
	for element := range s {
		elements = append(elements, element)
	}
	return elements
}

// Remove removes elements from s.
func (s Set[T]) Remove(elements ...T) {
	for _, element := range elements {
		delete(s, element)
	}
}
