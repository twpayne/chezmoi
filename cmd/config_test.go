package cmd

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-xdg/v3"
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

//nolint:unparam
func newTestBaseDirectorySpecification(homeDir string) *xdg.BaseDirectorySpecification {
	return &xdg.BaseDirectorySpecification{
		ConfigHome: filepath.Join(homeDir, ".config"),
		DataHome:   filepath.Join(homeDir, ".local"),
		CacheHome:  filepath.Join(homeDir, ".cache"),
		RuntimeDir: filepath.Join(homeDir, ".run"),
	}
}
