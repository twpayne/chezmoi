package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoTemplate(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		contentsStr          string
		data                 map[string]interface{}
		expected             string
		expectedReplacements bool
	}{
		{
			name:        "simple",
			contentsStr: "email = you@example.com\n",
			data: map[string]interface{}{
				"email": "you@example.com",
			},
			expected:             "email = {{ .email }}\n",
			expectedReplacements: true,
		},
		{
			name:        "longest_first",
			contentsStr: "name = John Smith\nfirstName = John\n",
			data: map[string]interface{}{
				"name":      "John Smith",
				"firstName": "John",
			},
			expected: "" +
				"name = {{ .name }}\n" +
				"firstName = {{ .firstName }}\n",
			expectedReplacements: true,
		},
		{
			name:        "alphabetical_first",
			contentsStr: "name = John Smith\n",
			data: map[string]interface{}{
				"alpha": "John Smith",
				"beta":  "John Smith",
				"gamma": "John Smith",
			},
			expected:             "name = {{ .alpha }}\n",
			expectedReplacements: true,
		},
		{
			name:        "nested_values",
			contentsStr: "email = you@example.com\n",
			data: map[string]interface{}{
				"personal": map[string]interface{}{
					"email": "you@example.com",
				},
			},
			expected:             "email = {{ .personal.email }}\n",
			expectedReplacements: true,
		},
		{
			name:        "only_replace_words",
			contentsStr: "darwinian evolution",
			data: map[string]interface{}{
				"os": "darwin",
			},
			expected: "darwinian evolution", // not "{{ .os }}ian evolution"
		},
		{
			name:        "longest_match_first",
			contentsStr: "/home/user",
			data: map[string]interface{}{
				"homeDir": "/home/user",
			},
			expected:             "{{ .homeDir }}",
			expectedReplacements: true,
		},
		{
			name:        "longest_match_first_prefix",
			contentsStr: "HOME=/home/user",
			data: map[string]interface{}{
				"homeDir": "/home/user",
			},
			expected:             "HOME={{ .homeDir }}",
			expectedReplacements: true,
		},
		{
			name:        "longest_match_first_suffix",
			contentsStr: "/home/user/something",
			data: map[string]interface{}{
				"homeDir": "/home/user",
			},
			expected:             "{{ .homeDir }}/something",
			expectedReplacements: true,
		},
		{
			name:        "longest_match_first_prefix_and_suffix",
			contentsStr: "HOME=/home/user/something",
			data: map[string]interface{}{
				"homeDir": "/home/user",
			},
			expected:             "HOME={{ .homeDir }}/something",
			expectedReplacements: true,
		},
		{
			name:        "words_only",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "a",
			},
			expected:             "aaa aa {{ .alpha }} aa aaa aa {{ .alpha }} aa aaa",
			expectedReplacements: true,
		},
		{
			name:        "words_only_2",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "aa",
			},
			expected:             "aaa {{ .alpha }} a {{ .alpha }} aaa {{ .alpha }} a {{ .alpha }} aaa",
			expectedReplacements: true,
		},
		{
			name:        "words_only_3",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "aaa",
			},
			expected:             "{{ .alpha }} aa a aa {{ .alpha }} aa a aa {{ .alpha }}",
			expectedReplacements: true,
		},
		{
			name:        "skip_empty",
			contentsStr: "a",
			data: map[string]interface{}{
				"empty": "",
			},
			expected: "a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actualTemplate, actualReplacements := autoTemplate([]byte(tc.contentsStr), tc.data)
			assert.Equal(t, tc.expected, string(actualTemplate))
			assert.Equal(t, tc.expectedReplacements, actualReplacements)
		})
	}
}

func TestInWord(t *testing.T) {
	for _, tc := range []struct {
		s        string
		i        int
		expected bool
	}{
		{s: "", i: 0, expected: false},
		{s: "a", i: 0, expected: false},
		{s: "a", i: 1, expected: false},
		{s: "ab", i: 0, expected: false},
		{s: "ab", i: 1, expected: true},
		{s: "ab", i: 2, expected: false},
		{s: "abc", i: 0, expected: false},
		{s: "abc", i: 1, expected: true},
		{s: "abc", i: 2, expected: true},
		{s: "abc", i: 3, expected: false},
		{s: " abc ", i: 0, expected: false},
		{s: " abc ", i: 1, expected: false},
		{s: " abc ", i: 2, expected: true},
		{s: " abc ", i: 3, expected: true},
		{s: " abc ", i: 4, expected: false},
		{s: " abc ", i: 5, expected: false},
		{s: "/home/user", i: 0, expected: false},
		{s: "/home/user", i: 1, expected: false},
		{s: "/home/user", i: 2, expected: true},
		{s: "/home/user", i: 3, expected: true},
		{s: "/home/user", i: 4, expected: true},
		{s: "/home/user", i: 5, expected: false},
		{s: "/home/user", i: 6, expected: false},
		{s: "/home/user", i: 7, expected: true},
		{s: "/home/user", i: 8, expected: true},
		{s: "/home/user", i: 9, expected: true},
		{s: "/home/user", i: 10, expected: false},
	} {
		assert.Equal(t, tc.expected, inWord(tc.s, tc.i))
	}
}
