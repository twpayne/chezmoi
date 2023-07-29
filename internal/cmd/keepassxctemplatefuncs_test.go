package cmd

import (
	"strconv"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestKeepassxcParseOutput(t *testing.T) {
	for i, tc := range []struct {
		output   []byte
		expected map[string]string
	}{
		{
			expected: map[string]string{},
		},
		{
			output: []byte(chezmoitest.JoinLines(
				"Title: test",
				"UserName: test",
				"Password: test",
				"URL:",
				"Notes: account: 123456789",
				"2021-11-27 [expires: 2023-02-25]",
				"main = false",
			)),
			expected: map[string]string{
				"Title":    "test",
				"UserName": "test",
				"Password": "test",
				"URL":      "",
				"Notes": strings.Join([]string{
					"account: 123456789",
					"2021-11-27 [expires: 2023-02-25]",
					"main = false",
				}, "\n"),
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := keepassxcParseOutput(tc.output)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
