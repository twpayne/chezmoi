package chezmoi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/absfs/afero"
	"github.com/d4l3k/messagediff"
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
		isTemplate bool
	}{
		{fileName: "foo", name: "foo", mode: os.FileMode(0666), isTemplate: false},
		{fileName: "dot_foo", name: ".foo", mode: os.FileMode(0666), isTemplate: false},
		{fileName: "private_foo", name: "foo", mode: os.FileMode(0600), isTemplate: false},
		{fileName: "private_dot_foo", name: ".foo", mode: os.FileMode(0600), isTemplate: false},
		{fileName: "executable_foo", name: "foo", mode: os.FileMode(0777), isTemplate: false},
		{fileName: "foo.tmpl", name: "foo", mode: os.FileMode(0666), isTemplate: true},
		{fileName: "private_executable_dot_foo.tmpl", name: ".foo", mode: os.FileMode(0700), isTemplate: true},
	} {
		if gotName, gotMode, gotIsTemplate := parseFileName(tc.fileName); gotName != tc.name || gotMode != tc.mode || gotIsTemplate != tc.isTemplate {
			t.Errorf("parseFileName(%q) == %q, %v, %v, want %q, %v, %v", tc.fileName, gotName, gotMode, gotIsTemplate, tc.name, tc.mode, tc.isTemplate)
		}
		if gotFileName := makeFileName(tc.name, tc.mode, tc.isTemplate); gotFileName != tc.fileName {
			t.Errorf("makeFileName(%q, %v, %v) == %q, want %q", tc.name, tc.mode, tc.isTemplate, gotFileName, tc.fileName)
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
		fs, err := makeMemMapFs(tc.fs)
		if err != nil {
			t.Errorf("makeMemMapFs(%v) == %v, %v, want !<nil>, <nil>", tc.fs, fs, err)
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
				"/home/user/.chezmoi/dot_bashrc":    "bar",
				"/home/user/.chezmoi/.git/HEAD":     "HEAD",
				"/home/user/.chezmoi/dot_hgrc.tmpl": "[ui]\nusername = {{ .name }} <{{ .email }}>\n",
			},
		},
	} {
		fs, err := makeMemMapFs(tc.fsMap)
		if err != nil {
			t.Errorf("case %d: makeMemMapFs(%v) == %v, %v, want !<nil>, <nil>", i, tc.fsMap, fs, err)
			continue
		}
		rs := NewRootState(tc.targetDir, tc.umask, tc.sourceDir, tc.data)
		if err := rs.Populate(fs); err != nil {
			t.Errorf("rs.Populate(%+v) == %v, want <nil>", fs, err)
			continue
		}
		if err := rs.Apply(fs, NewLoggingActuator(NewFsActuator(fs, tc.targetDir))); err != nil {
			t.Errorf("case %d: rs.Apply(makeMemMapFs(%v), _) == %v, want <nil>", i, tc.fsMap, err)
			continue
		}
		gotFsMap, err := makeMapFs(fs)
		if err != nil {
			t.Errorf("case %d: makeMapFs(%v) == %v, %v, want !<nil>, <nil>", i, fs, gotFsMap, err)
			continue
		}
		if diff, equal := messagediff.PrettyDiff(tc.wantFsMap, gotFsMap); !equal {
			t.Errorf("case %d:\n%s\n", i, diff)
		}
	}
}

func makeMemMapFs(fsMap map[string]string) (*afero.MemMapFs, error) {
	//fs := afero.NewMemMapFs()
	fs := &afero.MemMapFs{}
	for path, contents := range fsMap {
		if err := fs.MkdirAll(filepath.Dir(path), os.FileMode(0777)); err != nil {
			return nil, err
		}
		if err := afero.WriteFile(fs, path, []byte(contents), os.FileMode(0666)); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func makeMapFs(fs afero.Fs) (map[string]string, error) {
	mapFs := make(map[string]string)
	if err := afero.Walk(fs, "/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		contents, err := afero.ReadFile(fs, path)
		if err != nil {
			return err
		}
		mapFs[path] = string(contents)
		return nil
	}); err != nil {
		return nil, err
	}
	return mapFs, nil
}
