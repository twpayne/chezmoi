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
	} {
		got := autoTemplate([]byte(tc.contentsStr), tc.data)
		gotStr := string(got)
		if gotStr != tc.wantStr {
			t.Errorf("autoTemplate([]byte(%q), %v) == %q, want %q", tc.contentsStr, tc.data, gotStr, tc.wantStr)
		}
	}
}
