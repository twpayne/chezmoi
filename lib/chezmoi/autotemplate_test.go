package chezmoi

import "testing"

func TestAutoTemplate(t *testing.T) {
	for _, tc := range []struct {
		contentsStr string
		data        map[string]interface{}
		wantStr     string
	}{
		{
			contentsStr: "email = hello@example.com\n",
			data: map[string]interface{}{
				"email": "hello@example.com",
			},
			wantStr: "email = {{ .email }}\n",
		},
		{
			contentsStr: "name = John Smith\nfirstName = John\n",
			data: map[string]interface{}{
				"name":      "John Smith",
				"firstName": "John",
			},
			wantStr: "name = {{ .name }}\nfirstName = {{ .firstName }}\n",
		},
		{
			contentsStr: "email = hello@example.com\n",
			data: map[string]interface{}{
				"personal": map[string]interface{}{
					"email": "hello@example.com",
				},
			},
			wantStr: "email = {{ .personal.email }}\n",
		},
		{
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
		got := autoTemplate([]byte(tc.contentsStr), tc.data)
		gotStr := string(got)
		if gotStr != tc.wantStr {
			t.Errorf("autoTemplate([]byte(%q), %v) == %q, want %q", tc.contentsStr, tc.data, gotStr, tc.wantStr)
		}
	}
}
