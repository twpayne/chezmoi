package chezmoi

import (
	"os"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/go-vfs/vfstest"
)

func TestDirName(t *testing.T) {
	for _, tc := range []struct {
		dirName string
		name    string
		mode    os.FileMode
	}{
		{dirName: "foo", name: "foo", mode: 0777},
		{dirName: "dot_foo", name: ".foo", mode: 0777},
		{dirName: "private_foo", name: "foo", mode: 0700},
		{dirName: "private_dot_foo", name: ".foo", mode: 0700},
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

func TestSourceFileName(t *testing.T) {
	for _, tc := range []struct {
		sourceFileName string
		psfn           parsedSourceFileName
	}{
		{
			sourceFileName: "foo",
			psfn: parsedSourceFileName{
				fileName: "foo",
				perm:     0666,
				empty:    false,
				template: false,
			},
		},
		{
			sourceFileName: "dot_foo",
			psfn: parsedSourceFileName{
				fileName: ".foo",
				perm:     0666,
				empty:    false,
				template: false,
			},
		},
		{
			sourceFileName: "private_foo",
			psfn: parsedSourceFileName{
				fileName: "foo",
				perm:     0600,
				empty:    false,
				template: false,
			},
		},
		{
			sourceFileName: "private_dot_foo",
			psfn: parsedSourceFileName{
				fileName: ".foo",
				perm:     0600,
				empty:    false,
				template: false,
			},
		},
		{
			sourceFileName: "empty_foo",
			psfn: parsedSourceFileName{
				fileName: "foo",
				perm:     0666,
				empty:    true,
				template: false,
			},
		},
		{
			sourceFileName: "executable_foo",
			psfn: parsedSourceFileName{
				fileName: "foo",
				perm:     0777,
				empty:    false,
				template: false,
			},
		},
		{
			sourceFileName: "foo.tmpl",
			psfn: parsedSourceFileName{
				fileName: "foo",
				perm:     0666,
				empty:    false,
				template: true,
			},
		},
		{
			sourceFileName: "private_executable_dot_foo.tmpl",
			psfn: parsedSourceFileName{
				fileName: ".foo",
				perm:     0700,
				empty:    false,
				template: true,
			},
		},
	} {
		t.Run(tc.sourceFileName, func(t *testing.T) {
			gotPSFN := parseSourceFileName(tc.sourceFileName)
			if diff, equal := messagediff.PrettyDiff(tc.psfn, gotPSFN); !equal {
				t.Errorf("parseSourceFileName(%q) == %+v, want %+v, diff:\n%s", tc.sourceFileName, gotPSFN, tc.psfn, diff)
			}
			if gotSourceFileName := tc.psfn.SourceFileName(); gotSourceFileName != tc.sourceFileName {
				t.Errorf("%+v.SourceFileName() == %q, want %q", tc.psfn, gotSourceFileName, tc.sourceFileName)
			}
		})
	}
}

func TestTargetStatePopulate(t *testing.T) {
	for _, tc := range []struct {
		name      string
		root      interface{}
		sourceDir string
		data      map[string]interface{}
		want      *TargetState
	}{
		{
			name: "simple_file",
			root: map[string]string{
				"/foo": "bar",
			},
			sourceDir: "/",
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &File{
						sourceName: "foo",
						Perm:       0666,
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
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Entries: map[string]Entry{
					".foo": &File{
						sourceName: "dot_foo",
						Perm:       0666,
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
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &File{
						sourceName: "private_foo",
						Perm:       0600,
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
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Entries: map[string]Entry{
					"foo": &Dir{
						sourceName: "foo",
						Perm:       0777,
						Entries: map[string]Entry{
							"bar": &File{
								sourceName: "foo/bar",
								Perm:       0666,
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
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Entries: map[string]Entry{
					".foo": &Dir{
						sourceName: "private_dot_foo",
						Perm:       0700,
						Entries: map[string]Entry{
							"bar": &File{
								sourceName: "private_dot_foo/bar",
								Perm:       0666,
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
			want: &TargetState{
				TargetDir: "/",
				Umask:     0,
				SourceDir: "/",
				Data: map[string]interface{}{
					"Email": "user@example.com",
				},
				Entries: map[string]Entry{
					".gitconfig": &File{
						sourceName: "dot_gitconfig.tmpl",
						Perm:       0666,
						Contents:   []byte("[user]\n\temail = user@example.com\n"),
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfstest.NewTempFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfstest.NewTempFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			ts := NewTargetState("/", 0, tc.sourceDir, tc.data)
			if err := ts.Populate(fs); err != nil {
				t.Fatalf("ts.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if diff, equal := messagediff.PrettyDiff(tc.want, ts); !equal {
				t.Errorf("ts.Populate(%+v) diff:\n%s\n", fs, diff)
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
			umask:     022,
			tests: []vfstest.Test{
				vfstest.TestPath("/home/user/.bashrc", vfstest.TestModeIsRegular, vfstest.TestContentsString("bar")),
				vfstest.TestPath("/home/user/.hgrc", vfstest.TestModeIsRegular, vfstest.TestContentsString("[ui]\nusername = John Smith <hello@example.com>\n")),
				vfstest.TestPath("/home/user/foo", vfstest.TestModeIsRegular, vfstest.TestContents(nil)),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fs, cleanup, err := vfstest.NewTempFS(tc.root)
			defer cleanup()
			if err != nil {
				t.Fatalf("vfstest.NewTempFS(_) == _, _, %v, want _, _, <nil>", err)
			}
			ts := NewTargetState(tc.targetDir, tc.umask, tc.sourceDir, tc.data)
			if err := ts.Populate(fs); err != nil {
				t.Fatalf("ts.Populate(%+v) == %v, want <nil>", fs, err)
			}
			if err := ts.Apply(fs, NewLoggingActuator(os.Stderr, NewFSActuator(fs, tc.targetDir))); err != nil {
				t.Fatalf("ts.Apply(fs, _) == %v, want <nil>", err)
			}
			vfstest.RunTests(t, fs, "", tc.tests)
		})
	}
}
