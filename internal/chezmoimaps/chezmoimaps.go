// Package chezmoimaps implements common map functions.
package chezmoimaps

import (
	"cmp"
	"slices"
)

// Keys returns the keys of the map m.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// SortedKeys returns the keys of the map m in order.
func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	keys := Keys(m)
	slices.Sort(keys)
	return keys
}
