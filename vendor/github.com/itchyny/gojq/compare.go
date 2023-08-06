package gojq

import (
	"math"
	"math/big"
)

// Compare l and r, and returns jq-flavored comparison value.
// The result will be 0 if l == r, -1 if l < r, and +1 if l > r.
// This comparison is used by built-in operators and functions.
func Compare(l, r any) int {
	return compare(l, r)
}

func compare(l, r any) int {
	return binopTypeSwitch(l, r,
		compareInt,
		func(l, r float64) any {
			switch {
			case l < r || math.IsNaN(l):
				return -1
			case l == r:
				return 0
			default:
				return 1
			}
		},
		func(l, r *big.Int) any {
			return l.Cmp(r)
		},
		func(l, r string) any {
			switch {
			case l < r:
				return -1
			case l == r:
				return 0
			default:
				return 1
			}
		},
		func(l, r []any) any {
			n := len(l)
			if len(r) < n {
				n = len(r)
			}
			for i := 0; i < n; i++ {
				if cmp := compare(l[i], r[i]); cmp != 0 {
					return cmp
				}
			}
			return compareInt(len(l), len(r))
		},
		func(l, r map[string]any) any {
			lk, rk := funcKeys(l), funcKeys(r)
			if cmp := compare(lk, rk); cmp != 0 {
				return cmp
			}
			for _, k := range lk.([]any) {
				if cmp := compare(l[k.(string)], r[k.(string)]); cmp != 0 {
					return cmp
				}
			}
			return 0
		},
		func(l, r any) any {
			return compareInt(typeIndex(l), typeIndex(r))
		},
	).(int)
}

func compareInt(l, r int) any {
	switch {
	case l < r:
		return -1
	case l == r:
		return 0
	default:
		return 1
	}
}

func typeIndex(v any) int {
	switch v := v.(type) {
	default:
		return 0
	case bool:
		if !v {
			return 1
		}
		return 2
	case int, float64, *big.Int:
		return 3
	case string:
		return 4
	case []any:
		return 5
	case map[string]any:
		return 6
	}
}
