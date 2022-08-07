package cmd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuoteListTemplateFunc(t *testing.T) {
	c, err := newConfig()
	require.NoError(t, err)
	actual := c.quoteListTemplateFunc([]any{
		[]byte{65},
		"b",
		errors.New("error"),
		1,
		true,
	})
	assert.Equal(t, []string{
		`"A"`,
		`"b"`,
		`"error"`,
		`"1"`,
		`"true"`,
	}, actual)
}
