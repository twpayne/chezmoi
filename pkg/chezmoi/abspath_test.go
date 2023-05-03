package chezmoi

import (
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoitest"
)

func TestNewAbsPathFromExtPath(t *testing.T) {
	wd, err := os.Getwd()
	assert.NoError(t, err)
	wdAbsPath := NewAbsPath(wd)
	assert.NoError(t, err)
	homeDirAbsPath, err := NormalizePath(chezmoitest.HomeDir())
	assert.NoError(t, err)

	for _, tc := range []struct {
		name     string
		extPath  string
		expected AbsPath
	}{
		{
			name:     "empty",
			expected: wdAbsPath,
		},
		{
			name:     "file",
			extPath:  "file",
			expected: wdAbsPath.JoinString("file"),
		},
		{
			name:     "tilde",
			extPath:  "~",
			expected: homeDirAbsPath,
		},
		{
			name:     "tilde_home_file",
			extPath:  "~/file",
			expected: homeDirAbsPath.JoinString("file"),
		},
		{
			name:     "tilde_home_file_windows",
			extPath:  `~\file`,
			expected: homeDirAbsPath.JoinString("file"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)

			actual, err := NewAbsPathFromExtPath(tc.extPath, homeDirAbsPath)
			assert.NoError(t, err)
			normalizedActual, err := NormalizePath(actual.String())
			assert.NoError(t, err)
			expected, err := NormalizePath(tc.expected.String())
			assert.NoError(t, err)
			assert.Equal(t, expected, normalizedActual)
		})
	}
}
