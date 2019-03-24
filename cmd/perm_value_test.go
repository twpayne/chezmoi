package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermValue(t *testing.T) {
	for s, expected := range map[string]string{
		"0":   "000",
		"644": "644",
		"755": "755",
	} {
		var p permValue
		assert.NoError(t, p.Set(s))
		assert.Equal(t, expected, p.String())
		assert.Equal(t, "int", p.Type())
	}
}
