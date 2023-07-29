package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestTemplateParseAndExecute(t *testing.T) {
	for _, tc := range []struct {
		name        string
		dataStr     string
		expectedStr string
	}{
		{
			name: "missing_key",
			dataStr: chezmoitest.JoinLines(
				"# chezmoi:template:missing-key=invalid",
				"{{ .missing }}",
			),
			expectedStr: chezmoitest.JoinLines(
				"<no value>",
			),
		},
		{
			name: "delimiters",
			dataStr: chezmoitest.JoinLines(
				"# chezmoi:template:left-delimiter=[[ right-delimiter=]]",
				"[[ 0 ]]",
			),
			expectedStr: chezmoitest.JoinLines(
				"0",
			),
		},
		{
			name: "line_ending_crlf",
			dataStr: "" +
				"unix\n" +
				"\n" +
				"windows\r\n" +
				"\r\n" +
				"# chezmoi:template:line-ending=crlf\n",
			expectedStr: "" +
				"unix\r\n" +
				"\r\n" +
				"windows\r\n" +
				"\r\n",
		},
		{
			name: "line_ending_lf",
			dataStr: "" +
				"unix\n" +
				"\n" +
				"windows\r\n" +
				"\r\n" +
				"# chezmoi:template:line-ending=lf\n",
			expectedStr: chezmoitest.JoinLines(
				"unix",
				"",
				"windows",
				"",
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tc.name, []byte(tc.dataStr), nil, TemplateOptions{})
			assert.NoError(t, err)
			actual, err := tmpl.Execute(nil)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStr, string(actual))
		})
	}
}
