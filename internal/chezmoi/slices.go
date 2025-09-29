package chezmoi

// Distinct returns a slice containing the distinct elements in input slice.
// Ordering is preserved in relation to first appearance.
func Distinct[T comparable](input []T) []T {
	result := make([]T, 0, len(input))
	visited := make(map[T]struct{}, len(input))

	for _, s := range input {
		// Check if element was already visited.
		if _, ok := visited[s]; ok {
			continue
		}

		visited[s] = struct{}{}
		result = append(result, s)
	}

	return result
}
