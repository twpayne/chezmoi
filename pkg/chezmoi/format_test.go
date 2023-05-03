package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFormats(t *testing.T) {
	assert.NotZero(t, FormatsByName["json"])
	assert.NotZero(t, FormatsByName["jsonc"])
	assert.NotZero(t, FormatsByName["toml"])
	assert.NotZero(t, FormatsByName["yaml"])
	assert.Zero(t, FormatsByName["yml"])
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
			assert.NoError(t, err)
			var actualValue value
			assert.NoError(t, format.Unmarshal(data, &actualValue))
			assert.Equal(t, v, actualValue)
		})
	}
}
