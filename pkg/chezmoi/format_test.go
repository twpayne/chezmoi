package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormats(t *testing.T) {
	assert.Contains(t, FormatsByName, "json")
	assert.Contains(t, FormatsByName, "jsonc")
	assert.Contains(t, FormatsByName, "toml")
	assert.Contains(t, FormatsByName, "yaml")
	assert.NotContains(t, FormatsByName, "yml")
}

func TestFormatRoundTrip(t *testing.T) {
	type value struct {
		Bool   bool
		Bytes  []byte
		Int    int
		Object map[string]any
		String string
	}

	for _, format := range []Format{
		formatJSONC{},
		formatJSON{},
		formatTOML{},
		formatYAML{},
	} {
		t.Run(format.Name(), func(t *testing.T) {
			v := value{
				Bool:  true,
				Bytes: []byte("bytes"),
				Int:   1,
				Object: map[string]any{
					"key": "value",
				},
				String: "string",
			}
			data, err := format.Marshal(v)
			require.NoError(t, err)
			var actualValue value
			require.NoError(t, format.Unmarshal(data, &actualValue))
			assert.Equal(t, v, actualValue)
		})
	}
}
