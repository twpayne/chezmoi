package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/chezmoi/lib/chezmoi"
	vfs "github.com/twpayne/go-vfs"
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
		{
			name: "templates_dir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl":         `{{ template "foo" }}`,
					".chezmoitemplates/foo": "{{ if true }}contents{{ end }}",
				},
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
		name  string
		root  interface{}
		data  map[string]interface{}
		tests []vfst.Test
	}{
		{
			name: "simple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true": "#!/bin/sh\necho foo >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\nfoo\nfoo\n"),
				),
			},
		},
		{
			name: "simple_once",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_once_true": "#!/bin/sh\necho foo >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\n"),
				),
			},
		},
		{
			name: "template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/run_true.tmpl": "#!/bin/sh\necho {{ .Foo }} >>" + filepath.Join(tempDir, "evidence") + "\n",
			},
			data: map[string]interface{}{
				"Foo": "foo",
			},
			tests: []vfst.Test{
				vfst.TestPath(filepath.Join(tempDir, "evidence"),
					vfst.TestModeIsRegular,
					vfst.TestContentsString("foo\nfoo\nfoo\n"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.NewPathFS(vfs.OSFS, tempDir)
			defer func() {
				require.NoError(t, os.RemoveAll(tempDir))
				require.NoError(t, os.Mkdir(tempDir, 0700))
			}()
			require.NoError(t, vfst.NewBuilder().Build(fs, tc.root))
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
			// Run apply three times. As chezmoi should be idempotent, the
			// result should be the same each time.
			for i := 0; i < 3; i++ {
				assert.NoError(t, c.runApplyCmd(fs, nil))
			}
			vfst.RunTests(t, vfs.OSFS, "", tc.tests)
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
