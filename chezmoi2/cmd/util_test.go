package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpperSnakeCaseToCamelCaseMap(t *testing.T) {
	actual := upperSnakeCaseToCamelCaseMap(map[string]string{
		"BUG_REPORT_URL": "",
		"ID":             "",
	})
	assert.Equal(t, map[string]string{
		"bugReportURL": "",
		"id":           "",
	}, actual)
}
