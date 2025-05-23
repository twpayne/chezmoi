package cmd

import (
	"runtime"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/twpayne/go-vfs/v5"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestCatCmd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fails due to Windows paths on GitHub Actions")
	}
	for _, tc := range []struct {
		name        string
		root        any
		args        []string
		expectedStr string
	}{
		{
			name: "template_delimiters",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_template.tmpl": chezmoitest.JoinLines(
					`# chezmoi:template:left-delimiter=[[ right-delimiter=]]`,
					`[[ "ok" ]]`,
				),
			},
			args: []string{
				"/home/user/.template",
			},
			expectedStr: chezmoitest.JoinLines(
				"ok",
			),
		},
		{
			name: "json_indent",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_template.tmpl": chezmoitest.JoinLines(
					`# chezmoi:template:format-indent-width=3`,
					`{{ dict "a" (dict "b" "c") | toJson }}`,
				),
			},
			args: []string{
				"/home/user/.template",
			},
			expectedStr: chezmoitest.JoinLines(
				`{`,
				`   "a": {`,
				`      "b": "c"`,
				`   }`,
				`}`,
				``,
			),
		},
		{
			name: "yaml_indent",
			root: map[string]any{
				"/home/user/.local/share/chezmoi/dot_template.tmpl": chezmoitest.JoinLines(
					`# chezmoi:template:format-indent-width=3`,
					`{{ dict "a" (dict "b" "c") | toYaml }}`,
				),
			},
			args: []string{
				"/home/user/.template",
			},
			expectedStr: chezmoitest.JoinLines(
				`a:`,
				`   b: c`,
				``,
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.WithTestFS(t, tc.root, func(fileSystem vfs.FS) {
				stdout := strings.Builder{}
				c := newTestConfig(t, fileSystem, withStdout(&stdout))
				assert.NoError(t, c.execute(append([]string{"cat"}, tc.args...)))
				assert.Equal(t, tc.expectedStr, stdout.String())
			})
		})
	}
}
