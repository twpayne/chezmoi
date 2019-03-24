package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperSnakeCaseToCamelCase(t *testing.T) {
	for s, want := range map[string]string{
		"BUG_REPORT_URL":   "bugReportURL",
		"ID":               "id",
		"ID_LIKE":          "idLike",
		"NAME":             "name",
		"VERSION_CODENAME": "versionCodename",
		"VERSION_ID":       "versionID",
	} {
		assert.Equal(t, want, upperSnakeCaseToCamelCase(s))
	}
}
