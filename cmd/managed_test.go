package cmd

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestManagedCmd(t *testing.T) {
	for _, tc := range []struct {
		include             []string
		expectedTargetNames []string
	}{
		{
			include: []string{"dirs", "files", "symlinks"},
			expectedTargetNames: []string{
				"/home/user/dir",
				"/home/user/dir/file1",
				"/home/user/dir/subdir",
				"/home/user/dir/subdir/file2",
				"/home/user/symlink",
			},
		},
		{
			include: []string{"d", "f", "s"},
			expectedTargetNames: []string{
				"/home/user/dir",
				"/home/user/dir/file1",
				"/home/user/dir/subdir",
				"/home/user/dir/subdir/file2",
				"/home/user/symlink",
			},
		},
		{
			include: []string{"dirs"},
			expectedTargetNames: []string{
				"/home/user/dir",
				"/home/user/dir/subdir",
			},
		},
		{
			include: []string{"files"},
			expectedTargetNames: []string{
				"/home/user/dir/file1",
				"/home/user/dir/subdir/file2",
			},
		},
		{
			include: []string{"symlinks"},
			expectedTargetNames: []string{
				"/home/user/symlink",
			},
		},
		{
			include: []string{"f", "s"},
			expectedTargetNames: []string{
				"/home/user/dir/file1",
				"/home/user/dir/subdir/file2",
				"/home/user/symlink",
			},
		},
	} {
		t.Run(strings.Join(tc.include, "_"), func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file1":        "contents",
					"dir/subdir/file2": "contents",
					"symlink_symlink":  "target",
				},
			})
			require.NoError(t, err)
			defer cleanup()
			stdout := &bytes.Buffer{}
			c := newTestConfig(
				fs,
				withStdout(stdout),
				withManaged(managedCmdConfig{
					include: tc.include,
				}),
			)
			assert.NoError(t, c.runManagedCmd(nil, nil))
			posixTargetNames, err := extractPOSIXTargetNames(stdout.Bytes())
			require.NoError(t, err)
			assert.Equal(t, tc.expectedTargetNames, posixTargetNames)
		})
	}
}

// extractPOSIXTargetNames extracts all target names from b and coverts them to
// POSIX-like names.
func extractPOSIXTargetNames(b []byte) ([]string, error) {
	var targetNames []string
	s := bufio.NewScanner(bytes.NewBuffer(b))
	for s.Scan() {
		targetNames = append(targetNames, posixify(s.Text()))
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return targetNames, nil
}

// posixify returns a POSIX-like path based on path, stripping any volume name
// and converting backward slashes.
func posixify(path string) string {
	return filepath.ToSlash(strings.TrimPrefix(path, filepath.VolumeName(path)))
}

func withManaged(managed managedCmdConfig) configOption {
	return func(c *Config) {
		c.managed = managed
	}
}
