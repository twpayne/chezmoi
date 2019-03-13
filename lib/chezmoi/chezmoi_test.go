package chezmoi

import (
	"errors"
	"testing"
)

func TestReturnTemplateError(t *testing.T) {
	funcs := map[string]interface{}{
		"returnTemplateError": func() string {
			ReturnTemplateFuncError(errors.New("error"))
			return "foo"
		},
	}
	for name, dataString := range map[string]string{
		"syntax_error":         "{{",
		"unknown_field":        "{{ .Unknown }}",
		"unknown_func":         "{{ func }}",
		"func_returning_error": "{{ returnTemplateError }}",
	} {
		t.Run(name, func(t *testing.T) {
			ts := NewTargetState("/home/user", 0, "/home/user/.chezmoi", nil, funcs, "")
			if got, err := ts.executeTemplateData(name, []byte(dataString)); err == nil {
				t.Errorf("ts.executeTemplate(%q, %q) == %q, <nil>, want _, !<nil>", name, dataString, got)
			}
		})
	}
}
