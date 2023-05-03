package chezmoi

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestSourceRelPath(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		sourceStatePath       SourceRelPath
		expectedDirPath       SourceRelPath
		expectedTargetRelPath RelPath
	}{
		{
			name:            "empty",
			expectedDirPath: NewSourceRelDirPath("."),
		},
		{
			name:                  "dir",
			sourceStatePath:       NewSourceRelDirPath("dir"),
			expectedDirPath:       NewSourceRelDirPath("."),
			expectedTargetRelPath: NewRelPath("dir"),
		},
		{
			name:                  "exact_dir",
			sourceStatePath:       NewSourceRelDirPath("exact_dir"),
			expectedDirPath:       NewSourceRelDirPath("."),
			expectedTargetRelPath: NewRelPath("dir"),
		},
		{
			name:                  "exact_dir_private_dir",
			sourceStatePath:       NewSourceRelDirPath("exact_dir/private_dir"),
			expectedDirPath:       NewSourceRelDirPath("exact_dir"),
			expectedTargetRelPath: NewRelPath("dir/dir"),
		},
		{
			name:                  "file",
			sourceStatePath:       NewSourceRelPath("file"),
			expectedDirPath:       NewSourceRelDirPath("."),
			expectedTargetRelPath: NewRelPath("file"),
		},
		{
			name:                  "dot_file",
			sourceStatePath:       NewSourceRelPath("dot_file"),
			expectedDirPath:       NewSourceRelDirPath("."),
			expectedTargetRelPath: NewRelPath(".file"),
		},
		{
			name:                  "exact_dir_executable_file",
			sourceStatePath:       NewSourceRelPath("exact_dir/executable_file"),
			expectedDirPath:       NewSourceRelDirPath("exact_dir"),
			expectedTargetRelPath: NewRelPath("dir/file"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedDirPath, tc.sourceStatePath.Dir())
			assert.Equal(t, tc.expectedTargetRelPath, tc.sourceStatePath.TargetRelPath(""))
		})
	}
}
