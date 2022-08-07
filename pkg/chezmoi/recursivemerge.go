package chezmoi

// recursiveCopy returns a recursive copy of v.
func recursiveCopy(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}
	c := make(map[string]any)
	for key, value := range m {
		if mapValue, ok := value.(map[string]any); ok {
			c[key] = recursiveCopy(mapValue)
		} else {
			c[key] = value
		}
	}
	return c
}

// RecursiveMerge recursively merges maps in source into dest.
func RecursiveMerge(dest, source map[string]any) {
	for key, sourceValue := range source {
		destValue, ok := dest[key]
		if !ok {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		destMap, ok := destValue.(map[string]any)
		if !ok || destMap == nil {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		sourceMap, ok := sourceValue.(map[string]any)
		if !ok {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		RecursiveMerge(destMap, sourceMap)
	}
}
