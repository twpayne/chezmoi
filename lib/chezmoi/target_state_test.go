package chezmoi

import (
	"os"
	"testing"
	"text/template"

	"github.com/d4l3k/messagediff"
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
					"symlink_bar":              "empty",
					"symlink_replace_symlink":  "bar",
				},
			},
			sourceDir: "/home/user/.chezmoi",
			data: map[string]interface{}{
				"name":   "John Smith",
				"email":  "hello@example.com",
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
					vfst.TestContentsString("[ui]\nusername = John Smith <hello@example.com>\n"),
				),
				vfst.TestPath("/home/user/foo",
					vfst.TestModeIsRegular,
					vfst.TestContents(nil),
				),
				vfst.TestPath("/home/user/bar",
					vfst.TestModeType(os.ModeSymlink),
					vfst.TestSymlinkTarget("empty"),
				),
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
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			ts := NewTargetState(tc.destDir, tc.umask, tc.sourceDir, tc.data, tc.templateFuncs)
			if err := ts.Populate(fs); err != nil {
				t.Fatalf("ts.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if err := ts.Apply(fs, NewLoggingMutator(os.Stderr, NewFSMutator(fs, tc.destDir))); err != nil {
				t.Fatalf("ts.Apply(fs, _) == %v, want <nil>", err)
			}
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
								sourceName: "foo/bar",
								targetName: "foo/bar",
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
								sourceName: "private_dot_foo/bar",
								targetName: ".foo/bar",
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
				"Email": "user@example.com",
			},
			want: &TargetState{
				DestDir:      "/",
				TargetIgnore: NewPatternSet(),
				Umask:        0,
				SourceDir:    "/",
				Data: map[string]interface{}{
					"Email": "user@example.com",
				},
				Entries: map[string]Entry{
					".gitconfig": &File{
						sourceName: "dot_gitconfig.tmpl",
						targetName: ".gitconfig",
						Perm:       0666,
						Template:   true,
						contents:   []byte("[user]\n\temail = user@example.com\n"),
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
								sourceName: "exact_dir/foo",
								targetName: "dir/foo",
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
				"/.chezmoiignore": "f*\n",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir: "/",
				TargetIgnore: PatternSet(map[string]struct{}{
					"f*": {},
				}),
				Umask:     0,
				SourceDir: "/",
				Entries:   map[string]Entry{},
			},
		},
		{
			name: "ignore_subdir",
			root: map[string]interface{}{
				"/dir/.chezmoiignore": "foo\n",
			},
			sourceDir: "/",
			want: &TargetState{
				DestDir: "/",
				TargetIgnore: PatternSet(map[string]struct{}{
					"dir/foo": {},
				}),
				Umask:     0,
				SourceDir: "/",
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
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfst.NewTestFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfst.NewTestFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			ts := NewTargetState("/", 0, tc.sourceDir, tc.data, tc.templateFuncs)
			if err := ts.Populate(fs); err != nil {
				t.Fatalf("ts.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if err := ts.Evaluate(); err != nil {
				t.Errorf("ts.Evaluate() == %v, want <nil>", err)
			}
			if diff, equal := messagediff.PrettyDiff(tc.want, ts); !equal {
				t.Errorf("ts.Populate(%+v) diff:\n%s\n", fs, diff)
			}
		})
	}
}
