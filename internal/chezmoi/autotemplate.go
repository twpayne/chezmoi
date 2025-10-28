package chezmoi

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
)

// A TemplateVariable is a template variable. It is used instead of a
// map[string]string so that we can control order.
type TemplateVariable struct {
	Components []string
	Value      string
}

var templateMarkerRx = regexp.MustCompile(`\{{2,}|\}{2,}`)

// autoTemplate converts contents into a template by escaping template markers
// and replacing values in data with their keys. It returns the template and if
// any replacements were made.
func autoTemplate(contents []byte, data map[string]any) ([]byte, bool) {
	contentsStr := string(contents)
	replacements := false

	// Replace template markers.
	replacedTemplateMarkersStr := templateMarkerRx.ReplaceAllString(contentsStr, `{{ "$0" }}`)
	if replacedTemplateMarkersStr != contentsStr {
		contentsStr = replacedTemplateMarkersStr
		replacements = true
	}

	// Determine the priority order of replacements.
	//
	// Replace longest values first. If there are multiple matches for the same
	// length of value, then choose the shallowest first so that .variable is
	// preferred over .chezmoi.config.data.variable. If there are multiple
	// matches at the same depth, chose the variable that comes first
	// alphabetically.
	variables := extractVariables(data)
	slices.SortFunc(variables, func(a, b TemplateVariable) int {
		// First sort by value length, longest first.
		if compare := -cmp.Compare(len(a.Value), len(b.Value)); compare != 0 {
			return compare
		}
		// Second sort by value name depth, shallowest first.
		if compare := cmp.Compare(len(a.Components), len(b.Components)); compare != 0 {
			return compare
		}
		// Thirdly, sort by component names in alphabetical order.
		return slices.Compare(a.Components, b.Components)
	})

	// Replace variables in order.
	//
	// This naive approach will generate incorrect templates if the variable
	// names match variable values. The algorithm here is probably O(N^2), we
	// can do better.
	for _, variable := range variables {
		if variable.Value == "" {
			continue
		}

		index := strings.Index(contentsStr, variable.Value)
		for index != -1 && index != len(contentsStr) {
			if !inWord(contentsStr, index) && !inWord(contentsStr, index+len(variable.Value)) {
				// Replace variable.value which is on word boundaries at both
				// ends.
				replacement := "{{ ." + strings.Join(variable.Components, ".") + " }}"
				contentsStr = contentsStr[:index] + replacement + contentsStr[index+len(variable.Value):]
				index += len(replacement)
				replacements = true
			} else {
				// Otherwise, keep looking. Consume at least one byte so we make
				// progress.
				index++
			}

			// Look for the next occurrence of variable.value.
			j := strings.Index(contentsStr[index:], variable.Value)
			if j == -1 {
				// No more occurrences found, so terminate the loop.
				break
			}
			// Advance to the next occurrence.
			index += j
		}
	}

	return []byte(contentsStr), replacements
}

// appendVariables appends all template variables in data to variables
// and returns variables. data is assumed to be rooted at parent.
func appendVariables(variables []TemplateVariable, parent []string, data map[string]any) []TemplateVariable {
	for name, value := range data {
		switch value := value.(type) {
		case string:
			variable := TemplateVariable{
				Components: append(slices.Clone(parent), name),
				Value:      value,
			}
			variables = append(variables, variable)
		case map[string]any:
			variables = appendVariables(variables, append(parent, name), value)
		}
	}
	return variables
}

// extractVariables extracts all template variables from data.
func extractVariables(data map[string]any) []TemplateVariable {
	return appendVariables(nil, nil, data)
}

// inWord returns true if splitting s at position i would split a word.
func inWord(s string, i int) bool {
	return i > 0 && i < len(s) && isWord(s[i-1]) && isWord(s[i])
}

// isWord returns true if b is a word byte.
func isWord(b byte) bool {
	return '0' <= b && b <= '9' || 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z'
}
