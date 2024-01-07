package cmd

import (
	"errors"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/internal/chezmoiassert"
	"github.com/twpayne/chezmoi/v2/internal/chezmoitest"
)

func TestCommentTemplateFunc(t *testing.T) {
	prefix := "# "
	for i, tc := range []struct {
		s        string
		expected string
	}{
		{
			s:        "",
			expected: "",
		},
		{
			s:        "line",
			expected: "# line",
		},
		{
			s:        "\n",
			expected: "# \n",
		},
		{
			s:        "\n\n",
			expected: "# \n# \n",
		},
		{
			s:        "line1\nline2",
			expected: "# line1\n# line2",
		},
		{
			s:        "line1\nline2\n",
			expected: "# line1\n# line2\n",
		},
		{
			s:        "line1\n\nline3\n",
			expected: "# line1\n# \n# line3\n",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &Config{}
			assert.Equal(t, tc.expected, c.commentTemplateFunc(prefix, tc.s))
		})
	}
}

func TestDeleteValueAtPathTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name        string
		dict        map[string]any
		path        string
		expected    any
		expectedErr string
	}{
		{
			name:        "empty",
			expectedErr: "empty path",
		},
		{
			name: "outer",
			dict: map[string]any{
				"key": "value",
			},
			path:     "key",
			expected: map[string]any{},
		},
		{
			name: "inner",
			dict: map[string]any{
				"key1": map[string]any{
					"key2a": "value2a",
					"key2b": "value2b",
				},
			},
			path: "key1.key2a",
			expected: map[string]any{
				"key1": map[string]any{
					"key2b": "value2b",
				},
			},
		},
		{
			name: "missing",
			dict: map[string]any{
				"key": "value",
			},
			path: "missingKey",
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			name: "missing_inner",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": 0,
				},
			},
			path: "key1.key3",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": 0,
				},
			},
		},
		{
			name: "missing_depth2",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": map[string]any{
						"key3": 0,
					},
				},
			},
			path: "key1.key2.missingKey",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": map[string]any{
						"key3": 0,
					},
				},
			},
		},
		{
			name: "not_an_inner_dict",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": 0,
				},
			},
			path: "key1.key2.key3",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": 0,
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			if tc.expectedErr == "" {
				actual := c.deleteValueAtPathTemplateFunc(tc.path, tc.dict)
				assert.Equal(t, tc.expected, actual)
			} else {
				chezmoiassert.PanicsWithErrorString(t, tc.expectedErr, func() {
					c.deleteValueAtPathTemplateFunc(tc.path, tc.dict)
				})
			}
		})
	}
}

func TestFromJson(t *testing.T) {
	c, err := newConfig()
	assert.NoError(t, err)
	for i, tc := range []struct {
		expected any
		s        string
	}{
		{
			s:        `{"key":1}`,
			expected: map[string]any{"key": int64(1)},
		},
		{
			s:        `{"key":2.2}`,
			expected: map[string]any{"key": 2.2},
		},
		{
			s:        `{"key":[1,2.2,3]}`,
			expected: map[string]any{"key": []any{int64(1), 2.2, int64(3)}},
		},
		{
			s:        `{"key":1}`,
			expected: map[string]any{"key": int64(1)},
		},
		{
			s:        `{"key":1E400}`,
			expected: map[string]any{"key": "1E400"},
		},
		{
			s:        `{"key":3.141592653589793238462643383279}`,
			expected: map[string]any{"key": 3.141592653589793238462643383279},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, c.fromJsonTemplateFunc(tc.s))
		})
	}
}

func TestPruneEmptyDicts(t *testing.T) {
	for _, tc := range []struct {
		dict     map[string]any
		expected map[string]any
		name     string
	}{
		{
			name:     "nil",
			dict:     nil,
			expected: nil,
		},
		{
			name:     "empty",
			dict:     map[string]any{},
			expected: map[string]any{},
		},
		{
			name: "nested_empty",
			dict: map[string]any{
				"key1": map[string]any{},
				"key2": 0,
			},
			expected: map[string]any{
				"key2": 0,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, (&Config{}).pruneEmptyDictsTemplateFunc(tc.dict))
		})
	}
}

func TestSetValueAtPathTemplateFunc(t *testing.T) {
	for _, tc := range []struct {
		name        string
		path        any
		value       any
		dict        any
		expected    any
		expectedErr string
	}{
		{
			name:  "simple",
			path:  "key",
			value: "value",
			dict:  make(map[string]any),
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "create_map",
			path:  "key",
			value: "value",
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			name:  "modify_map",
			path:  "key2",
			value: "value2",
			dict: map[string]any{
				"key1": "value1",
			},
			expected: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "create_nested_map",
			path:  "key1.key2",
			value: "value",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": "value",
				},
			},
		},
		{
			name:  "modify_nested_map",
			path:  "key1.key2",
			value: "value",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
					"key3": "value3",
				},
				"key2": "value2",
			},
			expected: map[string]any{
				"key1": map[string]any{
					"key2": "value",
					"key3": "value3",
				},
				"key2": "value2",
			},
		},
		{
			name:  "replace_map",
			path:  "key1",
			value: "value1",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
				},
			},
			expected: map[string]any{
				"key1": "value1",
			},
		},
		{
			name:  "replace_nested_map",
			path:  "key1.key2",
			value: "value2",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": map[string]any{
						"key3": "value3",
					},
				},
			},
			expected: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
				},
			},
		},
		{
			name:  "replace_nested_value",
			path:  "key1.key2.key3",
			value: "value3",
			dict: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
				},
			},
			expected: map[string]any{
				"key1": map[string]any{
					"key2": map[string]any{
						"key3": "value3",
					},
				},
			},
		},
		{
			name: "string_list_path",
			path: []string{
				"key1",
				"key2",
			},
			value: "value2",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
				},
			},
		},
		{
			name: "any_list_path",
			path: []any{
				"key1",
				"key2",
			},
			value: "value2",
			expected: map[string]any{
				"key1": map[string]any{
					"key2": "value2",
				},
			},
		},
		{
			name:        "invalid_path",
			path:        0,
			expectedErr: "0: invalid path type int",
		},
		{
			name: "invalid_path_element",
			path: []any{
				0,
			},
			expectedErr: "0: invalid path element type int",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c Config
			if tc.expectedErr == "" {
				actual := c.setValueAtPathTemplateFunc(tc.path, tc.value, tc.dict)
				assert.Equal(t, tc.expected, actual)
			} else {
				chezmoiassert.PanicsWithErrorString(t, tc.expectedErr, func() {
					c.setValueAtPathTemplateFunc(tc.path, tc.value, tc.dict)
				})
			}
		})
	}
}

func TestFromIniTemplateFunc(t *testing.T) {
	for i, tc := range []struct {
		expected map[string]any
		text     string
	}{
		{
			text: chezmoitest.JoinLines(
				`key = value`,
			),
			expected: map[string]any{
				"key": "value",
			},
		},
		{
			text: chezmoitest.JoinLines(
				`[section]`,
				`sectionKey = sectionValue`,
			),
			expected: map[string]any{
				"section": map[string]any{
					"sectionKey": "sectionValue",
				},
			},
		},
		{
			text: chezmoitest.JoinLines(
				`key = value`,
				`[section]`,
				`sectionKey = sectionValue`,
			),
			expected: map[string]any{
				"key": "value",
				"section": map[string]any{
					"sectionKey": "sectionValue",
				},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &Config{}
			assert.Equal(t, tc.expected, c.fromIniTemplateFunc(tc.text))
		})
	}
}

func TestKeysFromPath(t *testing.T) {
	for _, tc := range []struct {
		path            any
		expectedErr     error
		name            string
		expectedLastKey string
		expectedKeys    []string
	}{
		{
			name:            "string_key",
			path:            "key",
			expectedKeys:    []string{},
			expectedLastKey: "key",
		},
		{
			name:            "string_period_separated_keys",
			path:            "key1.key2",
			expectedKeys:    []string{"key1"},
			expectedLastKey: "key2",
		},
		{
			name:            "string_period_separated_nested_keys",
			path:            "key1.key2.key3",
			expectedKeys:    []string{"key1", "key2"},
			expectedLastKey: "key3",
		},
		{
			name:        "string_empty",
			path:        "",
			expectedErr: errEmptyPath,
		},
		{
			name: "string_period_separated_empty_key",
			path: "key1..key3",
			expectedErr: emptyPathElementError{
				index: 1,
			},
		},
		{
			name:            "string_slice_one_key",
			path:            []string{"key1"},
			expectedKeys:    []string{},
			expectedLastKey: "key1",
		},
		{
			name:            "string_slice_two_keys",
			path:            []string{"key1", "key2"},
			expectedKeys:    []string{"key1"},
			expectedLastKey: "key2",
		},
		{
			name:            "string_slice_multiple_keys",
			path:            []string{"key1", "key2", "key3"},
			expectedKeys:    []string{"key1", "key2"},
			expectedLastKey: "key3",
		},
		{
			name:        "string_slice_empty",
			path:        []string{},
			expectedErr: errEmptyPath,
		},
		{
			name: "string_slice_empty_key",
			path: []string{""},
			expectedErr: emptyPathElementError{
				index: 0,
			},
		},
		{
			name: "string_slice_empty_key_second",
			path: []string{"key", ""},
			expectedErr: emptyPathElementError{
				index: 1,
			},
		},
		{
			name:        "any_slice_nil",
			expectedErr: errEmptyPath,
		},
		{
			name:        "any_slice_empty",
			path:        []any{},
			expectedErr: errEmptyPath,
		},
		{
			name:            "any_slice_one_key",
			path:            []any{"key"},
			expectedKeys:    []string{},
			expectedLastKey: "key",
		},
		{
			name:            "any_slice_two_keys",
			path:            []any{"key1", "key2"},
			expectedKeys:    []string{"key1"},
			expectedLastKey: "key2",
		},
		{
			name: "any_slice_invalid_key",
			path: []any{0},
			expectedErr: invalidPathElementTypeError{
				element: 0,
			},
		},
		{
			name: "any_slice_empty_key",
			path: []any{""},
			expectedErr: emptyPathElementError{
				index: 0,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualKeys, actualLastKey, err := keysFromPath(tc.path)
			if tc.expectedErr != nil {
				assert.Error(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedKeys, actualKeys)
				assert.Equal(t, tc.expectedLastKey, actualLastKey)
			}
		})
	}
}

func TestNestedMapAtPath(t *testing.T) {
	for _, tc := range []struct {
		path            any
		expectedErr     error
		m               map[string]any
		expectedMap     map[string]any
		name            string
		expectedLastKey string
	}{
		{
			name: "simple",
			m: map[string]any{
				"key": "value",
			},
			path: "key",
			expectedMap: map[string]any{
				"key": "value",
			},
			expectedLastKey: "key",
		},
		{
			name: "nested_map",
			m: map[string]any{
				"key1": map[string]any{
					"key2": "value",
				},
			},
			path: "key1.key2",
			expectedMap: map[string]any{
				"key2": "value",
			},
			expectedLastKey: "key2",
		},
		{
			name: "not_a_map",
			m: map[string]any{
				"key1": "value",
			},
			path: "key1.key2",
		},
		{
			name: "nested_not_a_map",
			m: map[string]any{
				"key1": map[string]any{
					"key2": "value",
				},
			},
			path: "key1.key2.key3",
		},
		{
			name:        "empty_path",
			expectedErr: errEmptyPath,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualMap, actualLastKey, err := nestedMapAtPath(tc.m, tc.path)
			if tc.expectedErr != nil {
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMap, actualMap)
				assert.Equal(t, tc.expectedLastKey, actualLastKey)
			}
		})
	}
}

func TestToIniTemplateFunc(t *testing.T) {
	for i, tc := range []struct {
		data     map[string]any
		expected string
	}{
		{
			data: map[string]any{
				"bool":   true,
				"float":  1.0,
				"int":    1,
				"string": "string",
			},
			expected: chezmoitest.JoinLines(
				`bool = true`,
				`float = 1.000000`,
				`int = 1`,
				`string = string`,
			),
		},
		{
			data: map[string]any{
				"bool":   "true",
				"float":  "1.0",
				"int":    "1",
				"string": "string string", //nolint:dupword
			},
			expected: chezmoitest.JoinLines(
				`bool = "true"`,
				`float = "1.0"`,
				`int = "1"`,
				`string = "string string"`,
			),
		},
		{
			data: map[string]any{
				"key": "value",
				"section": map[string]any{
					"subKey": "subValue",
				},
			},
			expected: chezmoitest.JoinLines(
				`key = value`,
				``,
				`[section]`,
				`subKey = subValue`,
			),
		},
		{
			data: map[string]any{
				"section": map[string]any{
					"subsection": map[string]any{
						"subSubKey": "subSubValue",
					},
				},
			},
			expected: chezmoitest.JoinLines(
				``,
				`[section]`,
				``,
				`[section.subsection]`,
				`subSubKey = subSubValue`,
			),
		},
		{
			data: map[string]any{
				"key": "value",
				"section": map[string]any{
					"subKey": "subValue",
					"subsection": map[string]any{
						"subSubKey": "subSubValue",
					},
				},
			},
			expected: chezmoitest.JoinLines(
				`key = value`,
				``,
				`[section]`,
				`subKey = subValue`,
				``,
				`[section.subsection]`,
				`subSubKey = subSubValue`,
			),
		},
		{
			data: map[string]any{
				"section1": map[string]any{
					"subKey1": "subValue1",
					"subsection1a": map[string]any{
						"subSubKey1a": "subSubValue1a",
					},
					"subsection1b": map[string]any{
						"subSubKey1b": "subSubValue1b",
					},
				},
				"section2": map[string]any{
					"subKey2": "subValue2",
					"subsection2a": map[string]any{
						"subSubKey2a": "subSubValue2a",
					},
					"subsection2b": map[string]any{
						"subSubKey2b": "subSubValue2b",
					},
				},
			},
			expected: chezmoitest.JoinLines(
				``,
				`[section1]`,
				`subKey1 = subValue1`,
				``,
				`[section1.subsection1a]`,
				`subSubKey1a = subSubValue1a`,
				``,
				`[section1.subsection1b]`,
				`subSubKey1b = subSubValue1b`,
				``,
				`[section2]`,
				`subKey2 = subValue2`,
				``,
				`[section2.subsection2a]`,
				`subSubKey2a = subSubValue2a`,
				``,
				`[section2.subsection2b]`,
				`subSubKey2b = subSubValue2b`,
			),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			c := &Config{}
			actual := c.toIniTemplateFunc(tc.data)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestNeedsQuote(t *testing.T) {
	for i, tc := range []struct {
		s        string
		expected bool
	}{
		{
			s:        "",
			expected: true,
		},
		{
			s:        "\\",
			expected: true,
		},
		{
			s:        "\a",
			expected: true,
		},
		{
			s:        "abc",
			expected: false,
		},
		{
			s:        "true",
			expected: true,
		},
		{
			s:        "1",
			expected: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tc.expected, needsQuote(tc.s))
		})
	}
}

func TestQuoteListTemplateFunc(t *testing.T) {
	c, err := newConfig()
	assert.NoError(t, err)
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
