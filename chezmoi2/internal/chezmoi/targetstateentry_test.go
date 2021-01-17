package chezmoi

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"

	"github.com/twpayne/chezmoi/chezmoi2/internal/chezmoitest"
)

func TestTargetStateEntryApplyAndEqual(t *testing.T) {
	targetStates := map[string]TargetStateEntry{
		"absent": &TargetStateAbsent{},
		"dir": &TargetStateDir{
			perm: 0o777,
		},
		"file": &TargetStateFile{
			perm: 0o666,
			lazyContents: &lazyContents{
				contents: []byte("# contents of file"),
			},
		},
		"file_empty": &TargetStateFile{
			perm: 0o666,
		},
		"file_executable": &TargetStateFile{
			perm: 0o777,
			lazyContents: &lazyContents{
				contents: []byte("#!/bin/sh\n"),
			},
		},
		"present": &TargetStatePresent{
			perm: 0o666,
		},
		"symlink": &TargetStateSymlink{
			lazyLinkname: &lazyLinkname{
				linkname: "target",
			},
		},
	}

	actualStates := map[string]map[string]interface{}{
		"absent": {
			"/home/user": &vfst.Dir{Perm: 0o777},
		},
		"dir": {
			"/home/user/target": &vfst.Dir{Perm: 0o777},
		},
		"file": {
			"/home/user/target": "# contents of file",
		},
		"file_empty": {
			"/home/user/target": "",
		},
		"file_executable": {
			"/home/user/target": &vfst.File{
				Perm:     0o777,
				Contents: []byte("!/bin/sh\n"),
			},
		},
		"symlink": {
			"/home/user": map[string]interface{}{
				"symlink-target": "",
				"target":         &vfst.Symlink{Target: "symlink-target"},
			},
		},
		"symlink_broken": {
			"/home/user/target": &vfst.Symlink{Target: "symlink-target"},
		},
	}

	targetStateKeys := make([]string, 0, len(targetStates))
	for targetStateKey := range targetStates {
		targetStateKeys = append(targetStateKeys, targetStateKey)
	}
	sort.Strings(targetStateKeys)

	actualDestDirStateKeys := make([]string, 0, len(actualStates))
	for actualDestDirStateKey := range actualStates {
		actualDestDirStateKeys = append(actualDestDirStateKeys, actualDestDirStateKey)
	}
	sort.Strings(actualDestDirStateKeys)

	testData := struct {
		TargetStateKey        []string
		ActualDestDirStateKey []string
	}{
		TargetStateKey:        targetStateKeys,
		ActualDestDirStateKey: actualDestDirStateKeys,
	}
	var testCases []struct {
		TargetStateKey        string
		ActualDestDirStateKey string
	}
	require.NoError(t, combinator.Generate(&testCases, testData))

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("target_%s_actual_%s", tc.TargetStateKey, tc.ActualDestDirStateKey), func(t *testing.T) {
			targetState := targetStates[tc.TargetStateKey]
			actualState := actualStates[tc.ActualDestDirStateKey]

			chezmoitest.WithTestFS(t, actualState, func(fs vfs.FS) {
				s := NewRealSystem(fs)

				// Read the initial destination state entry from fs.
				actualStateEntry, err := NewActualStateEntry(s, "/home/user/target", nil, nil)
				require.NoError(t, err)

				// Apply the target state entry.
				require.NoError(t, targetState.Apply(s, nil, actualStateEntry, GetUmask()))

				// Verify that the actual state entry matches the desired
				// state.
				vfst.RunTests(t, fs, "", vfst.TestPath("/home/user/target", targetStateTest(t, targetState)...))

				// Read the updated destination state entry from fs and
				// verify that it is equal to the target state entry.
				newActualStateEntry, err := NewActualStateEntry(s, "/home/user/target", nil, nil)
				require.NoError(t, err)
				equal, err := targetState.Equal(newActualStateEntry, GetUmask())
				require.NoError(t, err)
				require.True(t, equal)
			})
		})
	}
}

func targetStateTest(t *testing.T, ts TargetStateEntry) []vfst.PathTest {
	t.Helper()
	switch ts := ts.(type) {
	case *TargetStateAbsent:
		return []vfst.PathTest{
			vfst.TestDoesNotExist,
		}
	case *TargetStateDir:
		return []vfst.PathTest{
			vfst.TestIsDir,
			vfst.TestModePerm(ts.perm &^ GetUmask()),
		}
	case *TargetStateFile:
		expectedContents, err := ts.Contents()
		require.NoError(t, err)
		return []vfst.PathTest{
			vfst.TestModeIsRegular,
			vfst.TestContents(expectedContents),
			vfst.TestModePerm(ts.perm &^ GetUmask()),
		}
	case *TargetStatePresent:
		return []vfst.PathTest{
			vfst.TestModeIsRegular,
			vfst.TestModePerm(ts.perm &^ GetUmask()),
		}
	case *TargetStateRenameDir:
		// FIXME test for presence of newName
		return []vfst.PathTest{
			vfst.TestDoesNotExist,
		}
	case *TargetStateScript:
		return nil // FIXME how to verify scripts?
	case *TargetStateSymlink:
		expectedLinkname, err := ts.Linkname()
		require.NoError(t, err)
		return []vfst.PathTest{
			vfst.TestModeType(os.ModeSymlink),
			vfst.TestSymlinkTarget(expectedLinkname),
		}
	default:
		return nil
	}
}
