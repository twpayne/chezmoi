package chezmoi

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReturnTemplateError(t *testing.T) {
	funcs := map[string]interface{}{
		"returnTemplateError": func() string {
			panic(errors.New("error"))
		},
	}
	for name, dataString := range map[string]string{
		"syntax_error":         "{{",
		"unknown_field":        "{{ .Unknown }}",
		"unknown_func":         "{{ func }}",
		"func_returning_error": "{{ returnTemplateError }}",
	} {
		t.Run(name, func(t *testing.T) {
			ts := NewTargetState(
				WithDestDir("/home/user"),
				WithSourceDir("/home/user/.chezmoi"),
				WithTemplateFuncs(funcs),
			)
			_, err := ts.ExecuteTemplateData(name, []byte(dataString))
			assert.Error(t, err)
		})
	}
}
