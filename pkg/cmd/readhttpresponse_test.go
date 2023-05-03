package cmd

import (
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestMakeNightriderFrames(t *testing.T) {
	for i, tc := range []struct {
		shape    string
		padding  rune
		width    int
		expected []string
	}{
		{
			shape:   "+",
			padding: ' ',
			width:   1,
			expected: []string{
				"+",
			},
		},
		{
			shape:   "+",
			padding: ' ',
			width:   2,
			expected: []string{
				"+ ",
				" +",
			},
		},
		{
			shape:   "+",
			padding: ' ',
			width:   3,
			expected: []string{
				"+  ",
				" + ",
				"  +",
				" + ",
			},
		},
		{
			shape:   "<=>",
			padding: ' ',
			width:   1,
			expected: []string{
				"<",
			},
		},
		{
			shape:   "<=>",
			padding: ' ',
			width:   4,
			expected: []string{
				"<=> ",
				" <=>",
			},
		},
		{
			shape:   "<=>",
			padding: ' ',
			width:   5,
			expected: []string{
				"<=>  ",
				" <=> ",
				"  <=>",
				" <=> ",
			},
		},
		{
			shape:   "<=>",
			padding: ' ',
			width:   6,
			expected: []string{
				"<=>   ",
				" <=>  ",
				"  <=> ",
				"   <=>",
				"  <=> ",
				" <=>  ",
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := makeNightriderFrames(tc.shape, tc.padding, tc.width)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
