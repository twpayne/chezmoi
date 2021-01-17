package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormats(t *testing.T) {
	assert.Contains(t, Formats, "json")
	assert.Contains(t, Formats, "toml")
	assert.Contains(t, Formats, "yaml")
}
