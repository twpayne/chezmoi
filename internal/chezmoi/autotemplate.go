package chezmoi

import (
	"regexp"
	"slices"
	"sort"
	"strings"
)

// A templateVariable is a template variable. It is used instead of a
// map[string]string so that we can control order.
type templateVariable struct {
	components []string
	value      string
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
	sort.Slice(variables, func(i, j int) bool {
		// First sort by value length, longest first.
		valueI := variables[i].value
		valueJ := variables[j].value
		switch {
		case len(valueI) > len(valueJ):
			return true
		case len(valueI) == len(valueJ):
			// Second sort by value name depth, shallowest first.
			componentsI := variables[i].components
			componentsJ := variables[j].components
			switch {
			case len(componentsI) < len(componentsJ):
				return true
			case len(componentsI) == len(componentsJ):
				// Thirdly, sort by component names in alphabetical order.
				return slices.Compare(componentsI, componentsJ) < 0
			default:
				return false
			}
		default:
			return false
		}
	})

	// Replace variables in order.
	//
	// This naive approach will generate incorrect templates if the variable
	// names match variable values. The algorithm here is probably O(N^2), we
	// can do better.
	for _, variable := range variables {
		if variable.value == "" {
			continue
		}

		index := strings.Index(contentsStr, variable.value)
		for index != -1 && index != len(contentsStr) {
			if !inWord(contentsStr, index) && !inWord(contentsStr, index+len(variable.value)) {
				// Replace variable.value which is on word boundaries at both
				// ends.
				replacement := "{{ ." + strings.Join(variable.components, ".") + " }}"
				contentsStr = contentsStr[:index] + replacement + contentsStr[index+len(variable.value):]
				index += len(replacement)
				replacements = true
			} else {
				// Otherwise, keep looking. Consume at least one byte so we make
				// progress.
				index++
			}

			// Look for the next occurrence of variable.value.
			j := strings.Index(contentsStr[index:], variable.value)
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
func appendVariables(
	variables []templateVariable, parent []string, data map[string]any,
) []templateVariable {
	for name, value := range data {
		switch value := value.(type) {
		case string:
			variable := templateVariable{
				components: append(slices.Clone(parent), name),
				value:      value,
			}
			variables = append(variables, variable)
		case map[string]any:
			variables = appendVariables(variables, append(parent, name), value)
		}
	}
	return variables
}

// extractVariables extracts all template variables from data.
func extractVariables(data map[string]any) []templateVariable {
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
