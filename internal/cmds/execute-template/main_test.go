package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSlugify(t *testing.T) {
	for _, tc := range []struct {
		s        string
		expected string
	}{
		{
			s:        "add",
			expected: "add",
		},
		{
			s:        "*command*`.post.args`",
			expected: "command-post-args",
		},
	} {
		t.Run(tc.s, func(t *testing.T) {
			assert.Equal(t, tc.expected, slugify(tc.s))
		})
	}
}
