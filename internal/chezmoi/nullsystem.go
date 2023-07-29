package chezmoi

type NullSystem struct {
	emptySystemMixin
	noUpdateSystemMixin
}

// UnderlyingSystem implements System.UnderlyingSystem.
func (s *NullSystem) UnderlyingSystem() System {
	return s
}
