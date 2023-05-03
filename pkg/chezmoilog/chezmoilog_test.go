package chezmoilog

import (
	"errors"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestOutput(t *testing.T) {
	nonNilError := errors.New("")
	for i, tc := range []struct {
		data     []byte
		err      error
		expected []byte
	}{
		{
			data:     nil,
			err:      nil,
			expected: nil,
		},
		{
			data:     newByteSlice(0),
			err:      nil,
			expected: newByteSlice(0),
		},
		{
			data:     newByteSlice(16),
			err:      nil,
			expected: newByteSlice(16),
		},
		{
			data:     newByteSlice(2 * few),
			err:      nil,
			expected: append(newByteSlice(few), []byte("...")...),
		},
		{
			data:     newByteSlice(0),
			err:      nonNilError,
			expected: newByteSlice(0),
		},
		{
			data:     newByteSlice(few),
			err:      nonNilError,
			expected: newByteSlice(few),
		},
		{
			data:     newByteSlice(2 * few),
			err:      nonNilError,
			expected: newByteSlice(2 * few),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, Output(tc.data, tc.err))
		})
	}
}

func newByteSlice(n int) []byte {
	s := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, byte(i))
	}
	return s
}
