package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmpl, err := ParseTemplate(tc.name, []byte(tc.dataStr), nil, TemplateOptions{})
			require.NoError(t, err)
			actual, err := tmpl.Execute(nil)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStr, string(actual))
		})
	}
}
