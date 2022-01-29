package chezmoi

type relPathSet map[RelPath]struct{}

func newRelPathSet(relPaths []RelPath) relPathSet {
	s := make(relPathSet)
	for _, relPath := range relPaths {
		s[relPath] = struct{}{}
	}
	return s
}

func (s relPathSet) contains(relPath RelPath) bool {
	_, ok := s[relPath]
	return ok
}
