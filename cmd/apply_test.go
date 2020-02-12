package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	vfs "github.com/twpayne/go-vfs"
	"github.com/twpayne/go-vfs/vfst"
)

type scriptTestCase struct {
	name  string
	root  interface{}
	data  map[string]interface{}
	tests []vfst.Test
}

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
		{
			name: "define_template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl": `{{ define "foo" }}cont{{end}}{{ template "foo" }}ents`,
				},
			},
		},
		{
			name: "partial_template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl":         `{{ template "foo" }}ents`,
					".chezmoitemplates/foo": "{{ if true }}cont{{ end }}",
				},
			},
		},
		{
			name: "partial_template_in_subdir",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl":             `{{ template "foo" }}{{ template "bar/foo" }}`,
					".chezmoitemplates/foo":     "{{ if true }}cont{{ end }}",
					".chezmoitemplates/bar/foo": "{{ if true }}ents{{ end }}",
				},
			},
		},
		{
			name: "multiple_templates",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl":         `{{ template "foo" }}`,
					"dir/other.tmpl":        `{{ if true }}other stuff{{ end }}`,
					".chezmoitemplates/foo": "{{ if true }}contents{{ end }}",
				},
			},
		},
		{
			name: "multiple_associated",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"dir/file.tmpl":         `{{ template "foo" }}{{ template "bar" }}`,
					".chezmoitemplates/foo": "{{ if true }}cont{{ end }}",
					".chezmoitemplates/bar": "{{ if true }}ents{{ end }}",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.root["/home/user/.local/share/chezmoi/dir/file"] = "contents"
			tc.root["/home/user/.local/share/chezmoi/dir/other"] = "other stuff"
			tc.root["/home/user/.local/share/chezmoi/symlink_symlink"] = "target"
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := newTestConfig(fs)
			assert.NoError(t, c.runApplyCmd(nil, nil))
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
				vfst.TestPath("/home/user/dir/other",
					vfst.TestModeIsRegular,
					vfst.TestModePerm(0644),
					vfst.TestContentsString("other stuff"),
				),
				vfst.TestPath("/home/user/symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("target"),
				),
			)
		})
	}
}

func TestApplyFollow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		follow bool
		root   interface{}
		tests  []vfst.Test
	}{
		{
			name:   "follow",
			follow: true,
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					"bar": "baz",
					"foo": &vfst.Symlink{Target: "bar"},
				},
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": "qux",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/bar",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("qux"),
				),
				vfst.TestPath("/home/user/foo",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("bar"),
				),
			},
		},
		{
			name:   "nofollow",
			follow: false,
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					"bar": "baz",
					"foo": &vfst.Symlink{Target: "bar"},
				},
				"/home/user/.local/share/chezmoi": map[string]interface{}{
					"foo": "qux",
				},
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/bar",
					vfst.TestContentsString("baz"),
				),
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("qux"),
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := newTestConfig(
				fs,
				withFollow(tc.follow),
			)
			assert.NoError(t, c.runApplyCmd(nil, nil))
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestApplyRemove(t *testing.T) {
	for _, tc := range []struct {
		name     string
		noRemove bool
		root     interface{}
		data     map[string]interface{}
		tests    []vfst.Test
	}{
		{
			name: "simple",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "foo",
				"/home/user/foo": "# contents of foo\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name:     "no_remove",
			noRemove: true,
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "foo",
				"/home/user/foo": "# contents of foo\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of foo\n"),
				),
			},
		},
		{
			name: "pattern",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "f*",
				"/home/user/foo": "# contents of foo\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "template",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "{{ .bar }}",
				"/home/user/foo": "# contents of foo\n",
			},
			data: map[string]interface{}{
				"bar": "foo",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
		{
			name: "dont_remove_negative_pattern",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "f*\n!foo\n",
				"/home/user/foo": "# contents of foo\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of foo\n"),
				),
			},
		},
		{
			name: "dont_remove_ignored",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiignore": "foo",
				"/home/user/.local/share/chezmoi/.chezmoiremove": "f*",
				"/home/user/foo": "# contents of foo\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("# contents of foo\n"),
				),
			},
		},
		{
			name: "remove_subdirectory_first",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/.chezmoiremove": "foo\nfoo/bar\n",
				"/home/user/foo/bar":                             "# contents of bar\n",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo",
					vfst.TestDoesNotExist,
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := newTestConfig(
				fs,
				withData(tc.data),
				withRemove(!tc.noRemove),
			)
			assert.NoError(t, c.runApplyCmd(nil, nil))
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestApplyScript(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chezmoi")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()
	for _, tc := range getApplyScriptTestCases(tempDir) {
		t.Run(tc.name, func(t *testing.T) {
			fs := vfs.NewPathFS(vfs.OSFS, tempDir)
			require.NoError(t, vfst.NewBuilder().Build(fs, tc.root))
			defer func() {
				require.NoError(t, os.RemoveAll(tempDir))
				require.NoError(t, os.Mkdir(tempDir, 0700))
			}()
			apply := func() {
				c := newTestConfig(
					fs,
					withDestDir("/"),
					withData(tc.data),
				)
				require.NoError(t, c.runApplyCmd(nil, nil))
			}
			// Run apply three times. As chezmoi should be idempotent, the
			// result should be the same each time.
			for i := 0; i < 3; i++ {
				apply()
			}
			vfst.RunTests(t, vfs.OSFS, "", tc.tests)
		})
	}
}

func TestApplyRunOnce(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "chezmoi")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.RemoveAll(tempDir))
	}()
	tempFile := filepath.Join(tempDir, "foo")

	fs, cleanup, err := vfst.NewTestFS(
		[]interface{}{
			getRunOnceFiles(),
		},
	)
	require.NoError(t, err)
	defer cleanup()

	c := newTestConfig(
		fs,
		withDestDir("/"),
		withData(map[string]interface{}{
			"TempFile": tempFile,
		}),
	)

	require.NoError(t, c.runApplyCmd(nil, nil))
	vfst.RunTests(t, fs, "",
		vfst.TestPath("/home/user/.config/chezmoi/chezmoistate.boltdb",
			vfst.TestModeIsRegular,
		),
	)

	actualData, err := ioutil.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("bar\n"), actualData)

	require.NoError(t, c.runApplyCmd(nil, nil))
	actualData, err = ioutil.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, []byte("bar\n"), actualData)
}

func TestApplyRemoveEmptySymlink(t *testing.T) {
	for _, tc := range []struct {
		name  string
		root  interface{}
		tests []vfst.Test
	}{
		{
			name: "empty_symlink_remove_existing_symlink",
			root: map[string]interface{}{
				"/home/user/foo": &vfst.Symlink{Target: "bar"},
				"/home/user/.local/share/chezmoi/symlink_foo": "",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo", vfst.TestDoesNotExist),
			},
		},
		{
			name: "empty_symlink_remove_existing_dir",
			root: map[string]interface{}{
				"/home/user/foo/bar":                          "baz",
				"/home/user/.local/share/chezmoi/symlink_foo": "",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo", vfst.TestDoesNotExist),
			},
		},
		{
			name: "empty_symlink_no_target",
			root: map[string]interface{}{
				"/home/user/.local/share/chezmoi/symlink_foo": "",
			},
			tests: []vfst.Test{
				vfst.TestPath("/home/user/foo", vfst.TestDoesNotExist),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			c := newTestConfig(fs)
			assert.NoError(t, c.runApplyCmd(nil, nil))
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}
