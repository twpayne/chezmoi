package chezmoi

import "testing"

func TestAutoTemplate(t *testing.T) {
	for _, tc := range []struct {
		name        string
		contentsStr string
		data        map[string]interface{}
		wantStr     string
	}{
		{
			name:        "simple",
			contentsStr: "email = hello@example.com\n",
			data: map[string]interface{}{
				"email": "hello@example.com",
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
			contentsStr: "email = hello@example.com\n",
			data: map[string]interface{}{
				"personal": map[string]interface{}{
					"email": "hello@example.com",
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
		/*
			// FIXME this test currently fails because we match on word
			// boundaries and ^/ is not a word boundary.
			{
				contentsStr: "/home/user",
				data: map[string]interface{}{
					"homedir": "/home/user",
				},
				wantStr: "{{ .homedir }}",
			},
		*/
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := autoTemplate([]byte(tc.contentsStr), tc.data)
			gotStr := string(got)
			if gotErr != nil || gotStr != tc.wantStr {
				t.Errorf("autoTemplate([]byte(%q), %v) == %q, %v, want %q, <nil>", tc.contentsStr, tc.data, gotStr, gotErr, tc.wantStr)
			}
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
		if got := inWord(tc.s, tc.i); got != tc.want {
			t.Errorf("inWord(%q, %d) == %v, want %v", tc.s, tc.i, got, tc.want)
		}
	}
}
