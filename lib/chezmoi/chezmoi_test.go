package chezmoi

import (
	"os"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/aferot"
)

func TestDirName(t *testing.T) {
	for _, tc := range []struct {
		dirName string
		name    string
		mode    os.FileMode
	}{
		{dirName: "foo", name: "foo", mode: os.FileMode(0777)},
		{dirName: "dot_foo", name: ".foo", mode: os.FileMode(0777)},
		{dirName: "private_foo", name: "foo", mode: os.FileMode(0700)},
		{dirName: "private_dot_foo", name: ".foo", mode: os.FileMode(0700)},
	} {
		t.Run(tc.dirName, func(t *testing.T) {
			if gotName, gotMode := parseDirName(tc.dirName); gotName != tc.name || gotMode != tc.mode {
				t.Errorf("parseDirName(%q) == %q, %v, want %q, %v", tc.dirName, gotName, gotMode, tc.name, tc.mode)
			}
			if gotDirName := makeDirName(tc.name, tc.mode); gotDirName != tc.dirName {
				t.Errorf("makeDirName(%q, %v) == %q, want %q", tc.name, tc.mode, gotDirName, tc.dirName)
			}
		})
	}
}

func TestFileName(t *testing.T) {
	for _, tc := range []struct {
		fileName   string
		name       string
		mode       os.FileMode
		isEmpty    bool
		isTemplate bool
	}{
		{fileName: "foo", name: "foo", mode: os.FileMode(0666), isEmpty: false, isTemplate: false},
		{fileName: "dot_foo", name: ".foo", mode: os.FileMode(0666), isEmpty: false, isTemplate: false},
		{fileName: "private_foo", name: "foo", mode: os.FileMode(0600), isEmpty: false, isTemplate: false},
		{fileName: "private_dot_foo", name: ".foo", mode: os.FileMode(0600), isEmpty: false, isTemplate: false},
		{fileName: "empty_foo", name: "foo", mode: os.FileMode(0666), isEmpty: true, isTemplate: false},
		{fileName: "executable_foo", name: "foo", mode: os.FileMode(0777), isEmpty: false, isTemplate: false},
		{fileName: "foo.tmpl", name: "foo", mode: os.FileMode(0666), isEmpty: false, isTemplate: true},
		{fileName: "private_executable_dot_foo.tmpl", name: ".foo", mode: os.FileMode(0700), isEmpty: false, isTemplate: true},
	} {
		t.Run(tc.fileName, func(t *testing.T) {
			if gotName, gotMode, gotIsEmpty, gotIsTemplate := parseFileName(tc.fileName); gotName != tc.name || gotMode != tc.mode || gotIsEmpty != tc.isEmpty || gotIsTemplate != tc.isTemplate {
				t.Errorf("parseFileName(%q) == %q, %v, %v, %v want %q, %v, %v, %v", tc.fileName, gotName, gotMode, gotIsEmpty, gotIsTemplate, tc.name, tc.mode, tc.isEmpty, tc.isTemplate)
			}
			if gotFileName := makeFileName(tc.name, tc.mode, tc.isEmpty, tc.isTemplate); gotFileName != tc.fileName {
				t.Errorf("makeFileName(%q, %v, %v, %v) == %q, want %q", tc.name, tc.mode, tc.isEmpty, tc.isTemplate, gotFileName, tc.fileName)
			}
		})
	}
}

func TestRootStatePopulate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		root      interface{}
		sourceDir string
		data      map[string]interface{}
		want      *RootState
	}{
		{
			name: "simple_file",
			root: map[string]string{
				"/foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &FileState{
						sourceName: "foo",
						Mode:       os.FileMode(0666),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "dot_file",
			root: map[string]string{
				"/dot_foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Entries: map[string]Entry{
					".foo": &FileState{
						sourceName: "dot_foo",
						Mode:       os.FileMode(0666),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "private_file",
			root: map[string]string{
				"/private_foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &FileState{
						sourceName: "private_foo",
						Mode:       os.FileMode(0600),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			name: "file_in_subdir",
			root: map[string]string{
				"/foo/bar": "baz",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &Dir{
						sourceName: "foo",
						Mode:       os.FileMode(0777),
						Entries: map[string]Entry{
							"bar": &FileState{
								sourceName: "foo/bar",
								Mode:       os.FileMode(0666),
								Contents:   []byte("baz"),
							},
						},
					},
				},
			},
		},
		{
			name: "file_in_private_dot_subdir",
			root: map[string]string{
				"/private_dot_foo/bar": "baz",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Entries: map[string]Entry{
					".foo": &Dir{
						sourceName: "private_dot_foo",
						Mode:       os.FileMode(0700),
						Entries: map[string]Entry{
							"bar": &FileState{
								sourceName: "private_dot_foo/bar",
								Mode:       os.FileMode(0666),
								Contents:   []byte("baz"),
							},
						},
					},
				},
			},
		},
		{
			name: "template_dot_file",
			root: map[string]string{
				"/dot_gitconfig.tmpl": "[user]\n\temail = {{.Email}}\n",
			},
			sourceDir: "/",
			data: map[string]interface{}{
				"Email": "user@example.com",
			},
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Data: map[string]interface{}{
					"Email": "user@example.com",
				},
				Entries: map[string]Entry{
					".gitconfig": &FileState{
						sourceName: "dot_gitconfig.tmpl",
						Mode:       os.FileMode(0666),
						Contents:   []byte("[user]\n\temail = user@example.com\n"),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := aferot.NewTempOsFs(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("aferot.NewTempOsFs(_) == _, _, %v, want _, _, <nil>", err)
			}
			rs := NewRootState("/", 0, tc.sourceDir, tc.data)
			if err := rs.Populate(fs); err != nil {
				t.Fatalf("rs.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if diff, equal := messagediff.PrettyDiff(tc.want, rs); !equal {
				t.Errorf("rs.Populate(%+v) diff:\n%s\n", fs, diff)
			}
		})
	}
}

func TestEndToEnd(t *testing.T) {
	for _, tc := range []struct {
		name      string
		root      map[string]string
		sourceDir string
		data      map[string]interface{}
		targetDir string
		umask     os.FileMode
		tests     interface{}
	}{
		{
			name: "all",
			root: map[string]string{
				"/home/user/.bashrc":                "foo",
				"/home/user/.chezmoi/dot_bashrc":    "bar",
				"/home/user/.chezmoi/.git/HEAD":     "HEAD",
				"/home/user/.chezmoi/dot_hgrc.tmpl": "[ui]\nusername = {{ .name }} <{{ .email }}>\n",
				"/home/user/.chezmoi/empty.tmpl":    "{{ if false }}foo{{ end }}",
				"/home/user/.chezmoi/empty_foo":     "",
			},
			sourceDir: "/home/user/.chezmoi",
			data: map[string]interface{}{
				"name":  "John Smith",
				"email": "hello@example.com",
			},
			targetDir: "/home/user",
			umask:     os.FileMode(022),
			tests: []aferot.Test{
				aferot.TestPath("/home/user/.bashrc", aferot.TestModeIsRegular, aferot.TestContentsString("bar")),
				aferot.TestPath("/home/user/.hgrc", aferot.TestModeIsRegular, aferot.TestContentsString("[ui]\nusername = John Smith <hello@example.com>\n")),
				aferot.TestPath("/home/user/foo", aferot.TestModeIsRegular, aferot.TestContents(nil)),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := aferot.NewTempOsFs(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("aferot.NewTempOsFs(_) == _, _, %v, want _, _, <nil>", err)
			}
			rs := NewRootState(tc.targetDir, tc.umask, tc.sourceDir, tc.data)
			if err := rs.Populate(fs); err != nil {
				t.Fatalf("rs.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if err := rs.Apply(fs, NewLoggingActuator(os.Stderr, NewFsActuator(fs, tc.targetDir))); err != nil {
				t.Fatalf("rs.Apply(fs, _) == %v, want <nil>", err)
			}
			aferot.RunTest(t, fs, "", tc.tests)
		})
	}
}
