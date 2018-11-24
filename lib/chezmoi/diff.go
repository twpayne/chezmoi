package chezmoi

type section struct {
	ctype byte
	s     []string
}

// Inspired by https://github.com/paulgb/simplediff
func diff(before, after []string) []section {
	beforeMap := make(map[string][]int)
	for i, s := range before {
		beforeMap[s] = append(beforeMap[s], i)
	}
	overlap := make([]int, len(before))
	// Track start/len of largest overlapping match in old/new
	var startBefore, startAfter, subLen int
	for iafter, s := range after {
		o := make([]int, len(before))
		for _, ibefore := range beforeMap[s] {
			idx := 1
			if ibefore > 0 && overlap[ibefore-1] > 0 {
				idx = overlap[ibefore-1] + 1
			}
			o[ibefore] = idx
			if idx > subLen {
				// largest substring so far, store indices
				subLen = o[ibefore]
				startBefore = ibefore - subLen + 1
				startAfter = iafter - subLen + 1
			}
		}
		overlap = o
	}
	if subLen == 0 {
		// No common substring, issue - and +
		r := make([]section, 0)
		if len(before) > 0 {
			r = append(r, section{'-', before})
		}
		if len(after) > 0 {
			r = append(r, section{'+', after})
		}
		return r
	}
	// common substring unchanged, recurse on before/after substring
	r := diff(before[0:startBefore], after[0:startAfter])
	r = append(r, section{' ', after[startAfter : startAfter+subLen]})
	r = append(r, diff(before[startBefore+subLen:], after[startAfter+subLen:])...)
	return r
}
