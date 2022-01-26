package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeLinkname(t *testing.T) {
	for _, tc := range []struct {
		linkname string
		expected string
	}{
		{
			linkname: "rel",
			expected: "rel",
		},
		{
			linkname: "rel/forward",
			expected: "rel/forward",
		},
		{
			linkname: "rel\\backward",
			expected: "rel/backward",
		},
		{
			linkname: "rel/forward\\backward",
			expected: "rel/forward/backward",
		},
		{
			linkname: "/abs/forward",
			expected: "/abs/forward",
		},
		{
			linkname: "\\abs\\backward",
			expected: "/abs/backward",
		},
		{
			linkname: "/abs/forward\\backward",
			expected: "/abs/forward/backward",
		},
		{
			linkname: "c:/abs/forward",
			expected: "C:/abs/forward",
		},
		{
			linkname: "c:\\abs\\backward",
			expected: "C:/abs/backward",
		},
		{
			linkname: "c:/abs/forward\\backward",
			expected: "C:/abs/forward/backward",
		},
	} {
		t.Run(tc.linkname, func(t *testing.T) {
			assert.Equal(t, tc.expected, normalizeLinkname(tc.linkname))
		})
	}
}
