package chezmoi

import (
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twpayne/go-vfs/vfst"
)

func TestEndToEnd(t *testing.T) {
	for _, tc := range []struct {
		name          string
		root          interface{}
		sourceDir     string
		data          map[string]interface{}
		templateFuncs template.FuncMap
		destDir       string
		umask         os.FileMode
		tests         interface{}
	}{
		{
			name: "all",
			root: map[string]interface{}{
				"/home/user": map[string]interface{}{
					".bashrc": "foo",
					"dir": map[string]interface{}{
						"foo": "foo",
						"bar": "bar",
						"qux": "qux",
					},
					"replace_symlink": &vfst.Symlink{Target: "foo"},
				},
				"/home/user/.chezmoi": map[string]interface{}{
					".git/HEAD":                "HEAD",
					".chezmoiignore":           "{{ .ignore }} # comment\n",
					"README.md":                "contents of README.md\n",
					"dot_bashrc":               "bar",
					"dot_hgrc.tmpl":            "[ui]\nusername = {{ .name }} <{{ .email }}>\n",
					"empty.tmpl":               "{{ if false }}foo{{ end }}",
					"empty_foo":                "",
					"exact_dir/foo":            "foo",
					"exact_dir/.chezmoiignore": "qux\n",
					"whitespace":               " ",
					"symlink_bar":              "empty",
					"symlink_replace_symlink":  "bar",
				},
			},
			sourceDir: "/home/user/.chezmoi",
			data: map[string]interface{}{
				"name":   "John Smith",
				"email":  "john.smith@company.com",
				"ignore": "README.md",
			},
			destDir: "/home/user",
			umask:   022,
			tests: []vfst.Test{
				vfst.TestPath("/home/user/.bashrc",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("bar"),
				),
				vfst.TestPath("/home/user/.hgrc",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("[ui]\nusername = John Smith <john.smith@company.com>\n"),
				),
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContents(nil),
				),
				vfst.TestPath("/home/user/bar",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("empty"),
				),
				vfst.TestPath("/home/user/whitespace",
					vfst.TestDoesNotExist),
				vfst.TestPath("/home/user/replace_symlink",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("bar"),
				),
				vfst.TestPath("/home/user/dir/bar",
					vfst.TestDoesNotExist,
				),
				vfst.TestPath("/home/user/dir/qux",
					vfst.TestModeIsRegular,
					vfst.TestContentsString("qux"),
				),
				vfst.TestPath("/home/user/README.md",
					vfst.TestDoesNotExist,
				),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			ts := NewTargetState(tc.destDir, tc.umask, tc.sourceDir, tc.data, tc.templateFuncs, nil)
			assert.NoError(t, ts.Populate(fs, nil))
			applyOptions := &ApplyOptions{
				DestDir:           ts.DestDir,
				Ignore:            ts.TargetIgnore.Match,
				ScriptStateBucket: []byte("script"),
				Stdout:            os.Stdout,
				Umask:             022,
			}
			assert.NoError(t, ts.Apply(fs, NewLoggingMutator(os.Stderr, NewFSMutator(fs), false), applyOptions))
			vfst.RunTests(t, fs, "", tc.tests)
		})
	}
}

func TestTargetStatePopulate(t *testing.T) {
	for _, tc := range []struct {
		name          string
		root          interface{}
		sourceDir     string
		data          map[string]interface{}
		templateFuncs template.FuncMap
		want          *TargetState
	}{
		{
			name: "simple_file",
			root: map[string]interface{}{
				"/foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"foo": &File{
						sourceName: "foo",
						targetName: "foo",
						Perm:       0666,
						contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "dot_file",
			root: map[string]interface{}{
				"/dot_foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					".foo": &File{
						sourceName: "dot_foo",
						targetName: ".foo",
						Perm:       0666,
						contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "private_file",
			root: map[string]interface{}{
				"/private_foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"foo": &File{
						sourceName: "private_foo",
						targetName: "foo",
						Perm:       0600,
						contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "file_in_subdir",
			root: map[string]interface{}{
				"/foo/bar": "baz",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"foo": &Dir{
						sourceName: "foo",
						targetName: "foo",
						Exact:      false,
						Perm:       0777,
						Entries: map[string]Entry{
							"bar": &File{
								sourceName: filepath.Join("foo", "bar"),
								targetName: filepath.Join("foo", "bar"),
								Perm:       0666,
								contents:   []byte("baz"),
							},
						},
					},
				},
			},
		},
		{
			name: "file_in_private_dot_subdir",
			root: map[string]interface{}{
				"/private_dot_foo/bar": "baz",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					".foo": &Dir{
						sourceName: "private_dot_foo",
						targetName: ".foo",
						Exact:      false,
						Perm:       0700,
						Entries: map[string]Entry{
							"bar": &File{
								sourceName: filepath.Join("private_dot_foo", "bar"),
								targetName: filepath.Join(".foo", "bar"),
								Perm:       0666,
								contents:   []byte("baz"),
							},
						},
					},
				},
			},
		},
		{
			name: "template_dot_file",
			root: map[string]interface{}{
				"/dot_gitconfig.tmpl": "[user]\n\temail = {{.Email}}\n",
			},
			sourceDir: "/",
			data: map[string]interface{}{
				"Email": "john.smith@company.com",
			},
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Data: map[string]interface{}{
					"Email": "john.smith@company.com",
				},
				Entries: map[string]Entry{
					".gitconfig": &File{
						sourceName: "dot_gitconfig.tmpl",
						targetName: ".gitconfig",
						Perm:       0666,
						Template:   true,
						contents:   []byte("[user]\n\temail = john.smith@company.com\n"),
					},
				},
			},
		},
		{
			name: "file_in_exact_dir",
			root: map[string]interface{}{
				"/exact_dir/foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"dir": &Dir{
						sourceName: "exact_dir",
						targetName: "dir",
						Exact:      true,
						Perm:       0777,
						Entries: map[string]Entry{
							"foo": &File{
								sourceName: filepath.Join("exact_dir", "foo"),
								targetName: filepath.Join("dir", "foo"),
								Perm:       0666,
								contents:   []byte("bar"),
							},
						},
					},
				},
			},
		},
		{
			name: "symlink",
			root: map[string]interface{}{
				"/symlink_foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"foo": &Symlink{
						sourceName: "symlink_foo",
						targetName: "foo",
						linkname:   "bar",
					},
				},
			},
		},
		{
			name: "symlink_dot_foo",
			root: map[string]interface{}{
				"/symlink_dot_foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					".foo": &Symlink{
						sourceName: "symlink_dot_foo",
						targetName: ".foo",
						linkname:   "bar",
					},
				},
			},
		},
		{
			name: "symlink_template",
			root: map[string]interface{}{
				"/symlink_foo.tmpl": "bar-{{ .host }}",
			},
			sourceDir: "/",
			data: map[string]interface{}{
				"host": "example.com",
			},
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Data: map[string]interface{}{
					"host": "example.com",
				},
				Entries: map[string]Entry{
					"foo": &Symlink{
						sourceName: "symlink_foo.tmpl",
						targetName: "foo",
						Template:   true,
						linkname:   "bar-example.com",
					},
				},
			},
		},
		{
			name: "ignore_pattern",
			root: map[string]interface{}{
				"/.chezmoiignore": "" +
					"f*\n" +
					"!g\n",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir: "/",
				TargetIgnore: &PatternSet{
					includes: map[string]struct{}{
						"f*": {},
					},
					excludes: map[string]struct{}{
						"g": {},
					},
				},
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries:      map[string]Entry{},
			},
		},
		{
			name: "remove_pattern",
			root: map[string]interface{}{
				"/.chezmoiremove": "" +
					"f*\n" +
					"!g\n",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: &PatternSet{
					includes: map[string]struct{}{
						"f*": {},
					},
					excludes: map[string]struct{}{
						"g": {},
					},
				},
				Umask:     0,
				SourceDir: "/",
				Entries:   map[string]Entry{},
			},
		},
		{
			name: "ignore_subdir",
			root: map[string]interface{}{
				"/dir/.chezmoiignore": "" +
					"foo\n" +
					"!bar\n",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir: "/",
				TargetIgnore: &PatternSet{
					includes: map[string]struct{}{
						filepath.Join("dir", "foo"): {},
					},
					excludes: map[string]struct{}{
						filepath.Join("dir", "bar"): {},
					},
				},
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"dir": &Dir{
						sourceName: "dir",
						targetName: "dir",
						Perm:       0777,
						Entries:    map[string]Entry{},
					},
				},
			},
		},
		{
			name: "min_version",
			root: map[string]interface{}{
				"/.chezmoiversion": "1.2.3\n",
				"/foo":             "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetRemove: NewPatternSet(),
				TargetIgnore: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries: map[string]Entry{
					"foo": &File{
						sourceName: "foo",
						targetName: "foo",
						Perm:       0666,
						contents:   []byte("bar"),
					},
				},
				MinVersion: semver.Must(semver.NewVersion("1.2.3")),
			},
		},
		{
			name: "empty_template_dir",
			root: map[string]interface{}{
				"/.chezmoitemplates": &vfst.Dir{Perm: 0755},
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetRemove: NewPatternSet(),
				TargetIgnore: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries:      map[string]Entry{},
			},
		},
		{
			name: "template_dir",
			root: map[string]interface{}{
				"/.chezmoitemplates/foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				TargetRemove: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Entries:      map[string]Entry{},
				Templates: map[string]*template.Template{
					"foo": template.Must(template.New("foo").Parse("bar")),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			require.NoError(t, err)
			defer cleanup()
			ts := NewTargetState("/", 0, tc.sourceDir, tc.data, tc.templateFuncs, nil)
			assert.NoError(t, ts.Populate(fs, nil))
			assert.NoError(t, ts.Evaluate())
			assert.Equal(t, tc.want, ts)
		})
	}
}
