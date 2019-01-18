package cmd

import "testing"

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, want := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		if got := upperSnakeCaseToCamelCase(s); got != want {
			t.Errorf("upperSnakeCaseToCamelCase(%q) == %q, want %q", s, got, want)
		}
	}
}
