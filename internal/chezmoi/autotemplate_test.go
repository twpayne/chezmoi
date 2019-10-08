package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoTemplate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		contentsStr string
		data        map[string]interface{}
		wantStr     string
	}{
		{
			name:        "simple",
			contentsStr: "email = john.smith@company.com\n",
			data: map[string]interface{}{
				"email": "john.smith@company.com",
			},
			wantStr: "email = {{ .email }}\n",
		},
		{
			name:        "longest_first",
			contentsStr: "name = John Smith\nfirstName = John\n",
			data: map[string]interface{}{
				"name":      "John Smith",
				"firstName": "John",
			},
			wantStr: "name = {{ .name }}\nfirstName = {{ .firstName }}\n",
		},
		{
			name:        "alphabetical_first",
			contentsStr: "name = John Smith\n",
			data: map[string]interface{}{
				"alpha": "John Smith",
				"beta":  "John Smith",
				"gamma": "John Smith",
			},
			wantStr: "name = {{ .alpha }}\n",
		},
		{
			name:        "nested_values",
			contentsStr: "email = john.smith@company.com\n",
			data: map[string]interface{}{
				"personal": map[string]interface{}{
					"email": "john.smith@company.com",
				},
			},
			wantStr: "email = {{ .personal.email }}\n",
		},
		{
			name:        "only_replace_words",
			contentsStr: "darwinian evolution",
			data: map[string]interface{}{
				"os": "darwin",
			},
			wantStr: "darwinian evolution", // not "{{ .os }}ian evolution"
		},
		{
			name:        "longest_match_first",
			contentsStr: "/home/user",
			data: map[string]interface{}{
				"homedir": "/home/user",
			},
			wantStr: "{{ .homedir }}",
		},
		{
			name:        "longest_match_first_prefix",
			contentsStr: "HOME=/home/user",
			data: map[string]interface{}{
				"homedir": "/home/user",
			},
			wantStr: "HOME={{ .homedir }}",
		},
		{
			name:        "longest_match_first_suffix",
			contentsStr: "/home/user/something",
			data: map[string]interface{}{
				"homedir": "/home/user",
			},
			wantStr: "{{ .homedir }}/something",
		},
		{
			name:        "longest_match_first_prefix_and_suffix",
			contentsStr: "HOME=/home/user/something",
			data: map[string]interface{}{
				"homedir": "/home/user",
			},
			wantStr: "HOME={{ .homedir }}/something",
		},
		{
			name:        "words_only",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "a",
			},
			wantStr: "aaa aa {{ .alpha }} aa aaa aa {{ .alpha }} aa aaa",
		},
		{
			name:        "words_only_2",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "aa",
			},
			wantStr: "aaa {{ .alpha }} a {{ .alpha }} aaa {{ .alpha }} a {{ .alpha }} aaa",
		},
		{
			name:        "words_only_3",
			contentsStr: "aaa aa a aa aaa aa a aa aaa",
			data: map[string]interface{}{
				"alpha": "aaa",
			},
			wantStr: "{{ .alpha }} aa a aa {{ .alpha }} aa a aa {{ .alpha }}",
		},
		{
			name:        "skip_empty",
			contentsStr: "a",
			data: map[string]interface{}{
				"empty": "",
			},
			wantStr: "a",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantStr, string(autoTemplate([]byte(tc.contentsStr), tc.data)))
		})
	}
}

func TestInWord(t *testing.T) {
	for _, tc := range []struct {
		s    string
		i    int
		want bool
	}{
		{s: "", i: 0, want: false},
		{s: "a", i: 0, want: false},
		{s: "a", i: 1, want: false},
		{s: "ab", i: 0, want: false},
		{s: "ab", i: 1, want: true},
		{s: "ab", i: 2, want: false},
		{s: "abc", i: 0, want: false},
		{s: "abc", i: 1, want: true},
		{s: "abc", i: 2, want: true},
		{s: "abc", i: 3, want: false},
		{s: " abc ", i: 0, want: false},
		{s: " abc ", i: 1, want: false},
		{s: " abc ", i: 2, want: true},
		{s: " abc ", i: 3, want: true},
		{s: " abc ", i: 4, want: false},
		{s: " abc ", i: 5, want: false},
		{s: "/home/user", i: 0, want: false},
		{s: "/home/user", i: 1, want: false},
		{s: "/home/user", i: 2, want: true},
		{s: "/home/user", i: 3, want: true},
		{s: "/home/user", i: 4, want: true},
		{s: "/home/user", i: 5, want: false},
		{s: "/home/user", i: 6, want: false},
		{s: "/home/user", i: 7, want: true},
		{s: "/home/user", i: 8, want: true},
		{s: "/home/user", i: 9, want: true},
		{s: "/home/user", i: 10, want: false},
	} {
		assert.Equal(t, tc.want, inWord(tc.s, tc.i))
	}
}
