package chezmoi

import (
	"sort"
	"strings"
)

type templateVariable struct {
	name  string
	value string
}

type byValueLength []templateVariable

func (b byValueLength) Len() int { return len(b) }
func (b byValueLength) Less(i, j int) bool {
	switch {
	case len(b[i].value) < len(b[j].value):
		return true
	case len(b[i].value) == len(b[j].value):
		// Fallback to name
		return b[i].name > b[j].name
	default:
		return false
	}
}
func (b byValueLength) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func autoTemplate(contents []byte, data map[string]interface{}) []byte {
	// FIXME this naive approach will generate incorrect templates if the
	// variable names match variable values
	// FIXME the algorithm here is probably O(N^2), we can do better
	variables := extractVariables(nil, nil, data)
	sort.Sort(sort.Reverse(byValueLength(variables)))
	contentsStr := string(contents)
	for _, variable := range variables {
		if variable.value == "" {
			continue
		}
		index := strings.Index(contentsStr, variable.value)
		for index != -1 && index != len(contentsStr) {
			if !inWord(contentsStr, index) && !inWord(contentsStr, index+len(variable.value)) {
				// Replace variable.value which is on word boundaries at both
				// ends.
				replacement := "{{ ." + variable.name + " }}"
				contentsStr = contentsStr[:index] + replacement + contentsStr[index+len(variable.value):]
				index += len(replacement)
			} else {
				// Otherwise, keep looking. Consume at least one byte so we
				// make progress.
				index++
			}
			// Look for the next occurrence of variable.value.
			j := strings.Index(contentsStr[index:], variable.value)
			if j == -1 {
				// No more occurrences found, so terminate the loop.
				break
			} else {
				// Advance to the next occurrence.
				index += j
			}
		}
	}
	return []byte(contentsStr)
}

func extractVariables(variables []templateVariable, parent []string, data map[string]interface{}) []templateVariable {
	for name, value := range data {
		switch value := value.(type) {
		case string:
			variables = append(variables, templateVariable{
				name:  strings.Join(append(parent, name), "."),
				value: value,
			})
		case map[string]interface{}:
			variables = extractVariables(variables, append(parent, name), value)
		}
	}
	return variables
}

// inWord returns true if splitting s at position i would split a word.
func inWord(s string, i int) bool {
	return i > 0 && i < len(s) && isWord(s[i-1]) && isWord(s[i])
}

// isWord returns true if b is a word byte.
func isWord(b byte) bool {
	return '0' <= b && b <= '9' || 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z'
}
