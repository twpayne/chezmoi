package chezmoitest

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestJoinLines(t *testing.T) {
	for i, tc := range []struct {
		lines    []string
		expected string
	}{
		{
			lines:    nil,
			expected: "\n",
		},
		{
			lines:    []string{""},
			expected: "\n",
		},
		{
			lines:    []string{"a"},
			expected: "a\n",
		},
		{
			lines:    []string{"a", "b"},
			expected: "a\nb\n",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, JoinLines(tc.lines...))
		})
	}
}
