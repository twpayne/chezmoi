package chezmoi

// recursiveCopy returns a recursive copy of v.
func recursiveCopy(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v
	}
	c := make(map[string]interface{})
	for key, value := range m {
		if mapValue, ok := value.(map[string]interface{}); ok {
			c[key] = recursiveCopy(mapValue)
		} else {
			c[key] = value
		}
	}
	return c
}

// recursiveMerge recursively merges maps in source into dest.
func recursiveMerge(dest, source map[string]interface{}) {
	for key, sourceValue := range source {
		destValue, ok := dest[key]
		if !ok {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		destMap, ok := destValue.(map[string]interface{})
		if !ok || destMap == nil {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		sourceMap, ok := sourceValue.(map[string]interface{})
		if !ok {
			dest[key] = recursiveCopy(sourceValue)
			continue
		}
		recursiveMerge(destMap, sourceMap)
	}
}
