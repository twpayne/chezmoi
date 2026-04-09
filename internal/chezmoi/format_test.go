package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestFormatJSONSingleValue(t *testing.T) {
	var value any
	assert.NoError(t, FormatJSON.Unmarshal([]byte(`{}`), &value))
	assert.NoError(t, FormatJSON.Unmarshal([]byte(`{} `), &value))
	assert.Error(t, FormatJSON.Unmarshal([]byte(`{} 1`), &value))
}

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

func TestFormatEdgeCases(t *testing.T) {
	type T struct {
		Key string `json:"key" yaml:"key"`
	}
	for _, tc := range []struct {
		name        string
		format      Format
		data        string
		expected    T
		expectedErr string
	}{
		{
			name:     "jsonc",
			format:   FormatJSONC,
			data:     `{"key":"value"} // comment` + "\n",
			expected: T{Key: "value"},
		},
		{
			name:        "jsonc_empty",
			format:      FormatJSONC,
			expectedErr: "parsing value: unexpected EOF",
		},
		{
			name:     "jsonc_simple",
			format:   FormatJSONC,
			data:     `{"key":"value"}`,
			expected: T{Key: "value"},
		},
		{
			name:        "jsonc_trailing_value",
			format:      FormatJSONC,
			data:        `{"key":"value"}1`,
			expectedErr: "invalid character '1' after top-level value",
		},
		{
			name:        "jsonc_unknown_field",
			format:      FormatJSONC,
			data:        `{"unknown":"value"}`,
			expectedErr: `json: unknown field "unknown"`,
		},
		{
			name:        "jsonc_unexpected_eof",
			format:      FormatJSONC,
			data:        `{`,
			expectedErr: "parsing value: unexpected EOF",
		},
		{
			name:        "jsonc_whitespace",
			format:      FormatJSONC,
			data:        "\n",
			expectedErr: "parsing value: unexpected EOF",
		},
		{
			name:     "json",
			format:   FormatJSON,
			data:     `{"key":"value"}`,
			expected: T{Key: "value"},
		},
		{
			name:        "json_empty",
			format:      FormatJSON,
			expectedErr: "EOF",
		},
		{
			name:     "json_simple",
			format:   FormatJSON,
			data:     `{"key":"value"}`,
			expected: T{Key: "value"},
		},
		{
			name:        "json_unknown_field",
			format:      FormatJSON,
			data:        `{"unknown":"value"}`,
			expectedErr: `json: unknown field "unknown"`,
		},
		{
			name:        "json_unexpected_eof",
			format:      FormatJSON,
			data:        `{`,
			expectedErr: "unexpected EOF",
		},
		{
			name:        "json_whitespace",
			format:      FormatJSON,
			data:        "\n",
			expectedErr: "EOF",
		},
		{
			name:     "toml",
			format:   FormatTOML,
			data:     `key = "value"`,
			expected: T{Key: "value"},
		},
		{
			name:     "toml_1.1",
			format:   FormatTOML,
			data:     `key = "null byte: \x00; letter a: \x61"`,
			expected: T{Key: "null byte: \x00; letter a: a"},
		},
		{
			name:   "toml_empty",
			format: FormatTOML,
		},
		{
			name:     "toml_simple",
			format:   FormatTOML,
			data:     `key = "value"`,
			expected: T{Key: "value"},
		},
		{
			name:        "toml_unexpected_eof",
			format:      FormatTOML,
			data:        `[`,
			expectedErr: "unexpected end of table name",
		},
		{
			name:   "toml_whitespace",
			format: FormatTOML,
			data:   "\n",
		},
		{
			name:     "yaml",
			format:   FormatYAML,
			data:     `key: value`,
			expected: T{Key: "value"},
		},
		{
			name:   "yaml_empty",
			format: FormatYAML,
		},
		{
			name:     "yaml_simple",
			format:   FormatYAML,
			data:     "key: value",
			expected: T{Key: "value"},
		},
		{
			name:        "yaml_unknown_field",
			format:      FormatYAML,
			data:        "unknown: value",
			expectedErr: "unknown: value",
		},
		{
			name:        "yaml_unexpected_eof",
			format:      FormatYAML,
			data:        `{`,
			expectedErr: "could not find flow mapping end token '}'",
		},
		{
			name:   "yaml_whitespace",
			format: FormatYAML,
			data:   "\n",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var actual T
			err := tc.format.Unmarshal([]byte(tc.data), &actual)
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}
