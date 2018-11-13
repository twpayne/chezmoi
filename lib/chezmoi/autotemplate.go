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

func autoTemplate(contents []byte, data map[string]interface{}) []byte {
	// FIXME this naive approach will generate incorrect templates if the
	// variable names match variable values
	variables := extractVariables(nil, nil, data)
	sort.Sort(sort.Reverse(byValueLength(variables)))
	contentsStr := string(contents)
	for _, variable := range variables {
		contentsStr = strings.Replace(contentsStr, variable.value, "{{ ."+variable.name+" }}", -1)
	}
	return []byte(contentsStr)
}
