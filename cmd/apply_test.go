package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	"github.com/twpayne/go-vfs/vfst"
)

func TestApplyCommand(t *testing.T) {
	for _, tc := range []struct {
		name string
		root map[string]interface{}
	}{
		{
			name: "create",
			root: make(map[string]interface{}),
		},
		{
			name: "change_dir_permissions",
			root: map[string]interface{}{
				"/home/user/dir": &vfst.Dir{Perm: 0700},
			},
		},
		{
			name: "replace_file_with_dir",
			root: map[string]interface{}{
				"/home/user/dir": "file",
			},
		},
		{
			name: "replace_symlink_with_dir",
			root: map[string]interface{}{
				"/home/user/dir": &vfst.Symlink{Target: "target"},
			},
		},
		{
			name: "change_file_permissions",
			root: map[string]interface{}{
				"/home/user/dir/file": &vfst.File{
					Perm:     0755,
					Contents: []byte("contents"),
				},
			},
		},
		{
			name: "replace_dir_with_file",
			root: map[string]interface{}{
				"/home/user/dir/file": &vfst.Dir{Perm: 0755},
			},
		},
		{
			name: "replace_symlink_with_file",
			root: map[string]interface{}{
				"/home/user/dir/file": &vfst.Symlink{Target: "target"},
			},
		},
		{
			name: "replace_dir_with_symlink",
			root: map[string]interface{}{
				"/home/user/symlink": &vfst.Dir{Perm: 0755},
			},
		},
		{
			name: "replace_file_with_symlink",
			root: map[string]interface{}{
				"/home/user/symlink": "contents",
			},
		},
		{
			name: "change_symlink_target",
			root: map[string]interface{}{
				"/home/user/symlink": &vfst.Symlink{Target: "file"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.root["/home/user/.local/share/chezmoi/dir/file"] = "contents"
			tc.root["/home/user/.local/share/chezmoi/symlink_symlink"] = "target"
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := &Config{
				SourceDir: "/home/user/.local/share/chezmoi",
				DestDir:   "/home/user",
				Umask:     022,
			}
			assert.NoError(t, c.runApplyCmd(fs, nil))
			vfst.RunTests(t, fs, "",
				vfst.TestPath("/home/user/dir",
					vfst.TestIsDir,
					vfst.TestModePerm(0755),
				),
				vfst.TestPath("/home/user/dir/file",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0644),
					vfst.TestContentsString("contents"),
				),
				vfst.TestPath("/home/user/symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("target"),
				),
			)
		})
	}
}

func TestApplyScript(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chezmoi")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()
	for _, tc := range []struct {
		name     string
		root     interface{}
		data     map[string]interface{}
		evidence string
	}{
		{
			name: "simple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true": "#!/bin/sh\ntouch " + filepath.Join(tempDir, "simple") + "\n",
			},
			evidence: "simple",
		},
		{
			name: "simple_once",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_once_true": "#!/bin/sh\ntouch " + filepath.Join(tempDir, "simple_once") + "\n",
			},
			evidence: "simple_once",
		},
		{
			name: "template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true.tmpl": "#!/bin/sh\ntouch {{ .Evidence }}\n",
			},
			data: map[string]interface{}{
				"Evidence": filepath.Join(tempDir, "template"),
			},
			evidence: "template",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			persistentState, err := chezmoi.NewBoltPersistentState(fs, "/home/user/.config/chezmoi/chezmoistate.boltdb")
			require.NoError(t, err)
			c := &Config{
				SourceDir:         "/home/user/.local/share/chezmoi",
				DestDir:           "/",
				Umask:             022,
				Data:              tc.data,
				persistentState:   persistentState,
				scriptStateBucket: []byte("script"),
			}
			assert.NoError(t, c.runApplyCmd(fs, nil))
			evidencePath := filepath.Join(tempDir, tc.evidence)
			_, err = os.Stat(evidencePath)
			assert.NoError(t, err)
			assert.NoError(t, os.Remove(evidencePath))
		})
	}
}

func TestApplyRunOnce(t *testing.T) {
	statePath := "/home/user/.config/chezmoi/chezmoistate.boltdb"

	tempDir, err := ioutil.TempDir("", "chezmoi")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()
	tempFile := filepath.Join(tempDir, "foo")

	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		filepath.Dir(statePath):                             &vfst.Dir{Perm: 0755},
		"/home/user/.local/share/chezmoi/run_once_foo.tmpl": "#!/bin/sh\necho bar >> {{ .TempFile }}\n",
	})
	require.NoError(t, err)
	defer cleanup()

	persistentState, err := chezmoi.NewBoltPersistentState(fs, statePath)
	require.NoError(t, err)

	c := &Config{
		SourceDir: "/home/user/.local/share/chezmoi",
		DestDir:   "/",
		Umask:     022,
		Data: map[string]interface{}{
			"TempFile": tempFile,
		},
		persistentState:   persistentState,
		scriptStateBucket: []byte("script"),
	}

	require.NoError(t, c.runApplyCmd(fs, nil))
	vfst.RunTests(t, fs, "",
		vfst.TestPath(statePath,
			vfst.TestModeIsRegular,
		),
	)
	actualData, err := ioutil.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("bar\n"), actualData)

	require.NoError(t, c.runApplyCmd(fs, nil))
	actualData, err = ioutil.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("bar\n"), actualData)
}
