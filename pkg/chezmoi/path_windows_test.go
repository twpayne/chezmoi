package chezmoi

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"
)

func TestAbsPathTrimDirPrefix(t *testing.T) {
	for i, tc := range []struct {
		absPath          AbsPath
		dirPrefixAbsPath AbsPath
		expected         RelPath
	}{
		{
			absPath:          NewAbsPath("/home/user/.config"),
			dirPrefixAbsPath: NewAbsPath("/home/user"),
			expected:         NewRelPath(".config"),
		},
		{
			absPath:          NewAbsPath("H:/.config"),
			dirPrefixAbsPath: NewAbsPath("H:"),
			expected:         NewRelPath(".config"),
		},
		{
			absPath:          NewAbsPath("H:/.config"),
			dirPrefixAbsPath: NewAbsPath("H:/"),
			expected:         NewRelPath(".config"),
		},
		{
			absPath:          NewAbsPath("H:/home/user/.config"),
			dirPrefixAbsPath: NewAbsPath("H:/home/user"),
			expected:         NewRelPath(".config"),
		},
		{
			absPath:          NewAbsPath(`//server/user/.config`),
			dirPrefixAbsPath: NewAbsPath(`//server/user`),
			expected:         NewRelPath(".config"),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual, err := tc.absPath.TrimDirPrefix(tc.dirPrefixAbsPath)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

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
