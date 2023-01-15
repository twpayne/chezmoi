package cmd

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func init() {
	// github.com/twpayne/chezmoi/v2/pkg/chezmoi reads the umask before
	// github.com/twpayne/chezmoi/v2/pkg/chezmoitest sets it, so update it.
	chezmoi.Umask = chezmoitest.Umask
}

func TestDeDuplicateError(t *testing.T) {
	for i, tc := range []struct {
		errStr   string
		expected string
	}{
		{
			errStr:   "",
			expected: "",
		},
		{
			errStr:   "a",
			expected: "a",
		},
		{
			errStr:   "a: a",
			expected: "a",
		},
		{
			errStr:   "a: b",
			expected: "a: b",
		},
		{
			errStr:   "a: a: b", //nolint:dupword
			expected: "a: b",
		},
		{
			errStr:   "a: b: b",
			expected: "a: b",
		},
		{
			errStr:   "a: b: c: b: a: d",
			expected: "a: b: c: d",
		},
		{
			errStr:   "a: b: a: b: c",
			expected: "a: b: c",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			actual := deDuplicateError(errors.New(tc.errStr))
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMustGetLongHelpPanics(t *testing.T) {
	assert.Panics(t, func() {
		mustLongHelp("non-existent-command")
	})
}
