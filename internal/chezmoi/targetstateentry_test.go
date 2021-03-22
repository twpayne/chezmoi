package chezmoi

import (
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/muesli/combinator"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs/v2"
	"github.com/twpayne/go-vfs/v2/vfst"

	"github.com/twpayne/chezmoi/internal/chezmoitest"
)

func TestTargetStateEntryApplyAndEqual(t *testing.T) {
	targetStates := map[string]TargetStateEntry{
		"dir": &TargetStateDir{
			perm: 0o777 &^ chezmoitest.Umask,
		},
		"file": &TargetStateFile{
			perm: 0o666 &^ chezmoitest.Umask,
			lazyContents: &lazyContents{
				contents: []byte("# contents of file"),
			},
		},
		"file_empty": &TargetStateFile{
			perm: 0o666 &^ chezmoitest.Umask,
		},
		"file_executable": &TargetStateFile{
			perm: 0o777 &^ chezmoitest.Umask,
			lazyContents: &lazyContents{
				contents: []byte("#!/bin/sh\n"),
			},
		},
		"remove": &TargetStateRemove{},
		"symlink": &TargetStateSymlink{
			lazyLinkname: &lazyLinkname{
				linkname: "target",
			},
		},
	}

	actualStates := map[string]map[string]interface{}{
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
		"remove": {
			"/home/user": &vfst.Dir{Perm: 0o777},
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
				_, err = targetState.Apply(s, nil, actualStateEntry, chezmoitest.Umask)
				require.NoError(t, err)

				// Verify that the actual state entry matches the desired
				// state.
				vfst.RunTests(t, fs, "", vfst.TestPath("/home/user/target", targetStateTest(t, targetState)...))
			})
		})
	}
}

func targetStateTest(t *testing.T, ts TargetStateEntry) []vfst.PathTest {
	t.Helper()
	switch ts := ts.(type) {
	case *TargetStateRemove:
		return []vfst.PathTest{
			vfst.TestDoesNotExist,
		}
	case *TargetStateDir:
		return []vfst.PathTest{
			vfst.TestIsDir,
			vfst.TestModePerm(ts.perm &^ chezmoitest.Umask),
		}
	case *TargetStateFile:
		expectedContents, err := ts.Contents()
		require.NoError(t, err)
		return []vfst.PathTest{
			vfst.TestModeIsRegular,
			vfst.TestContents(expectedContents),
			vfst.TestModePerm(ts.perm &^ chezmoitest.Umask),
		}
	case *targetStateRenameDir:
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
