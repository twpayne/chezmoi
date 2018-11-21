package chezmoi

import (
	"os"
	"testing"

	"github.com/d4l3k/messagediff"
	"github.com/twpayne/chezmoi/internal/absfstesting"
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
		if gotName, gotMode := parseDirName(tc.dirName); gotName != tc.name || gotMode != tc.mode {
			t.Errorf("parseDirName(%q) == %q, %v, want %q, %v", tc.dirName, gotName, gotMode, tc.name, tc.mode)
		}
		if gotDirName := makeDirName(tc.name, tc.mode); gotDirName != tc.dirName {
			t.Errorf("makeDirName(%q, %v) == %q, want %q", tc.name, tc.mode, gotDirName, tc.dirName)
		}
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
		if gotName, gotMode, gotIsEmpty, gotIsTemplate := parseFileName(tc.fileName); gotName != tc.name || gotMode != tc.mode || gotIsEmpty != tc.isEmpty || gotIsTemplate != tc.isTemplate {
			t.Errorf("parseFileName(%q) == %q, %v, %v, %v want %q, %v, %v, %v", tc.fileName, gotName, gotMode, gotIsEmpty, gotIsTemplate, tc.name, tc.mode, tc.isEmpty, tc.isTemplate)
		}
		if gotFileName := makeFileName(tc.name, tc.mode, tc.isEmpty, tc.isTemplate); gotFileName != tc.fileName {
			t.Errorf("makeFileName(%q, %v, %v, %v) == %q, want %q", tc.name, tc.mode, tc.isEmpty, tc.isTemplate, gotFileName, tc.fileName)
		}
	}
}

func TestRootStatePopulate(t *testing.T) {
	for _, tc := range []struct {
		fs        map[string]string
		sourceDir string
		data      map[string]interface{}
		want      *RootState
	}{
		{
			fs: map[string]string{
				"/foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Dirs:      map[string]*DirState{},
				Files: map[string]*FileState{
					"foo": {
						sourceName: "foo",
						Mode:       os.FileMode(0666),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/dot_foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Dirs:      map[string]*DirState{},
				Files: map[string]*FileState{
					".foo": {
						sourceName: "dot_foo",
						Mode:       os.FileMode(0666),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/private_foo": "bar",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Dirs:      map[string]*DirState{},
				Files: map[string]*FileState{
					"foo": {
						sourceName: "private_foo",
						Mode:       os.FileMode(0600),
						Contents:   []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/foo/bar": "baz",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Dirs: map[string]*DirState{
					"foo": {
						sourceName: "foo",
						Mode:       os.FileMode(0777),
						Dirs:       map[string]*DirState{},
						Files: map[string]*FileState{
							"bar": {
								sourceName: "foo/bar",
								Mode:       os.FileMode(0666),
								Contents:   []byte("baz"),
							},
						},
					},
				},
				Files: map[string]*FileState{},
			},
		},
		{
			fs: map[string]string{
				"/private_dot_foo/bar": "baz",
			},
			sourceDir: "/",
			want: &RootState{
				TargetDir: "/",
				Umask:     os.FileMode(0),
				SourceDir: "/",
				Dirs: map[string]*DirState{
					".foo": {
						sourceName: "private_dot_foo",
						Mode:       os.FileMode(0700),
						Dirs:       map[string]*DirState{},
						Files: map[string]*FileState{
							"bar": {
								sourceName: "private_dot_foo/bar",
								Mode:       os.FileMode(0666),
								Contents:   []byte("baz"),
							},
						},
					},
				},
				Files: map[string]*FileState{},
			},
		},
		{
			fs: map[string]string{
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
				Dirs: map[string]*DirState{},
				Files: map[string]*FileState{
					".gitconfig": {
						sourceName: "dot_gitconfig.tmpl",
						Mode:       os.FileMode(0666),
						Contents:   []byte("[user]\n\temail = user@example.com\n"),
					},
				},
			},
		},
	} {
		fs, err := absfstesting.MakeMemMapFs(tc.fs)
		if err != nil {
			t.Errorf("absfstesting.MakeMemMapFs(%v) == %v, %v, want !<nil>, <nil>", tc.fs, fs, err)
			continue
		}
		rs := NewRootState("/", 0, tc.sourceDir, tc.data)
		if err := rs.Populate(fs); err != nil {
			t.Errorf("rs.Populate(%+v) == %v, want <nil>", fs, err)
			continue
		}
		if diff, equal := messagediff.PrettyDiff(tc.want, rs); !equal {
			t.Errorf("rs.Populate(%+v) diff:\n%s\n", fs, diff)
		}
	}
}

func TestEndToEnd(t *testing.T) {
	for i, tc := range []struct {
		fsMap     map[string]string
		sourceDir string
		data      map[string]interface{}
		targetDir string
		umask     os.FileMode
		wantFsMap map[string]string
	}{
		{
			fsMap: map[string]string{
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
			wantFsMap: map[string]string{
				"/home/user/.bashrc":                "bar",
				"/home/user/.hgrc":                  "[ui]\nusername = John Smith <hello@example.com>\n",
				"/home/user/foo":                    "",
				"/home/user/.chezmoi/dot_bashrc":    "bar",
				"/home/user/.chezmoi/.git/HEAD":     "HEAD",
				"/home/user/.chezmoi/dot_hgrc.tmpl": "[ui]\nusername = {{ .name }} <{{ .email }}>\n",
				"/home/user/.chezmoi/empty.tmpl":    "{{ if false }}foo{{ end }}",
				"/home/user/.chezmoi/empty_foo":     "",
			},
		},
	} {
		fs, err := absfstesting.MakeMemMapFs(tc.fsMap)
		if err != nil {
			t.Errorf("case %d: absfstesting.MakeMemMapFs(%v) == %v, %v, want !<nil>, <nil>", i, tc.fsMap, fs, err)
			continue
		}
		rs := NewRootState(tc.targetDir, tc.umask, tc.sourceDir, tc.data)
		if err := rs.Populate(fs); err != nil {
			t.Errorf("rs.Populate(%+v) == %v, want <nil>", fs, err)
			continue
		}
		if err := rs.Apply(fs, NewLoggingActuator(NewFsActuator(fs, tc.targetDir))); err != nil {
			t.Errorf("case %d: rs.Apply(absfstesting.MakeMemMapFs(%v), _) == %v, want <nil>", i, tc.fsMap, err)
			continue
		}
		gotFsMap, err := absfstesting.MakeMapFs(fs)
		if err != nil {
			t.Errorf("case %d: absfstesting.MakeMapFs(%v) == %v, %v, want !<nil>, <nil>", i, fs, gotFsMap, err)
			continue
		}
		if diff, equal := messagediff.PrettyDiff(tc.wantFsMap, gotFsMap); !equal {
			t.Errorf("case %d:\n%s\n", i, diff)
		}
	}
}
