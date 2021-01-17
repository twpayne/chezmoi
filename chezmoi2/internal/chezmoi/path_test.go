package chezmoi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

func TestNewAbsPathFromExtPath(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	wdAbsPath := AbsPath(wd)
	require.NoError(t, err)
	homeDirAbsPath, err := NormalizePath(chezmoitest.HomeDir())
	require.NoError(t, err)

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
			expected: wdAbsPath.Join("file"),
		},
		{
			name:     "tilde",
			extPath:  "~",
			expected: homeDirAbsPath,
		},
		{
			name:     "tilde_home_file",
			extPath:  "~/file",
			expected: homeDirAbsPath + "/file",
		},
		{
			name:     "tilde_home_file_windows",
			extPath:  `~\file`,
			expected: homeDirAbsPath + "/file",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			chezmoitest.SkipUnlessGOOS(t, tc.name)

			actual, err := NewAbsPathFromExtPath(tc.extPath, homeDirAbsPath)
			require.NoError(t, err)
			normalizedActual, err := NormalizePath(string(actual))
			require.NoError(t, err)
			expected, err := NormalizePath(string(tc.expected))
			require.NoError(t, err)
			assert.Equal(t, expected, normalizedActual)
		})
	}
}
