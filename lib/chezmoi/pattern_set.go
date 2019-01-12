package chezmoi

import "path/filepath"

// An PatternSet is a set of patterns.
type PatternSet map[string]struct{}

// NewPatternSet returns a new PatternSet.
func NewPatternSet() PatternSet {
	return PatternSet(make(map[string]struct{}))
}

// Add adds pattern to ps.
func (ps PatternSet) Add(pattern string) error {
	if _, err := filepath.Match(pattern, ""); err != nil {
		return nil
	}
	ps[pattern] = struct{}{}
	return nil
}

// Match returns if name matches any pattern in ps.
func (ps PatternSet) Match(name string) bool {
	for pattern := range ps {
		if ok, _ := filepath.Match(pattern, name); ok {
			return true
		}
	}
	return false
}
