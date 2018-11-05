package chezmoi

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/absfs/afero"
	"github.com/d4l3k/messagediff"
)

func TestReadSourceDirState(t *testing.T) {
	for _, tc := range []struct {
		fs        map[string]string
		sourceDir string
		data      interface{}
		want      *DirState
	}{
		{
			fs: map[string]string{
				"/foo": "bar",
			},
			sourceDir: "/",
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{},
				Files: map[string]*FileState{
					"foo": &FileState{
						Name:     "foo",
						Mode:     os.FileMode(0666),
						Contents: []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/dot_foo": "bar",
			},
			sourceDir: "/",
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{},
				Files: map[string]*FileState{
					".foo": &FileState{
						Name:     ".foo",
						Mode:     os.FileMode(0666),
						Contents: []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/private_foo": "bar",
			},
			sourceDir: "/",
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{},
				Files: map[string]*FileState{
					"foo": &FileState{
						Name:     "foo",
						Mode:     os.FileMode(0600),
						Contents: []byte("bar"),
					},
				},
			},
		},
		{
			fs: map[string]string{
				"/foo/bar": "baz",
			},
			sourceDir: "/",
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{
					"foo": &DirState{
						Name: "foo",
						Mode: os.FileMode(0777),
						Dirs: map[string]*DirState{},
						Files: map[string]*FileState{
							"bar": &FileState{
								Name:     "bar",
								Mode:     os.FileMode(0666),
								Contents: []byte("baz"),
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
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{
					".foo": &DirState{
						Name: ".foo",
						Mode: os.FileMode(0700),
						Dirs: map[string]*DirState{},
						Files: map[string]*FileState{
							"bar": &FileState{
								Name:     "bar",
								Mode:     os.FileMode(0666),
								Contents: []byte("baz"),
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
			data: map[string]string{
				"Email": "user@example.com",
			},
			want: &DirState{
				Name: "",
				Mode: os.FileMode(0),
				Dirs: map[string]*DirState{},
				Files: map[string]*FileState{
					".gitconfig": &FileState{
						Name:     ".gitconfig",
						Mode:     os.FileMode(0666),
						Contents: []byte("[user]\n\temail = user@example.com\n"),
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
		if got, err := ReadSourceDirState(fs, tc.sourceDir, tc.data); err != nil || !reflect.DeepEqual(got, tc.want) {
			diff, _ := messagediff.PrettyDiff(tc.want, got)
			t.Errorf("ReadSourceDirState(makeMemMapFs(%v), %q, %v) == %+v, %v, want %+v, <nil>:\n%s", tc.fs, tc.sourceDir, tc.data, got, err, tc.want, diff)
		}
	}
}

func TestEndToEnd(t *testing.T) {
	for i, tc := range []struct {
		fsMap     map[string]string
		sourceDir string
		data      interface{}
		targetDir string
		wantFsMap map[string]string
	}{
		{
			fsMap: map[string]string{
				"/home/user/.bashrc":             "foo",
				"/home/user/.chezmoi/dot_bashrc": "bar",
			},
			sourceDir: "/home/user/.chezmoi",
			targetDir: "/home/user",
			wantFsMap: map[string]string{
				"/home/user/.bashrc":             "bar",
				"/home/user/.chezmoi/dot_bashrc": "bar",
			},
		},
	} {
		fs, err := makeMemMapFs(tc.fsMap)
		if err != nil {
			t.Errorf("case %d: makeMemMapFs(%v) == %v, %v, want !<nil>, <nil>", i, tc.fsMap, fs, err)
			continue
		}
		ds, err := ReadSourceDirState(fs, tc.sourceDir, tc.data)
		if err != nil {
			t.Errorf("case %d: ReadSourceDirState(makeMemMapFs(%v), %q, %v) == %v, %v, want !<nil>, <nil>", i, tc.fsMap, tc.sourceDir, tc.data, ds, err)
			continue
		}
		if err := ds.Apply(fs, tc.targetDir); err != nil {
			t.Errorf("case %d: %v.Apply(makeMemMapFs(%v), %q) == %v, want <nil>", i, ds, tc.fsMap, tc.targetDir, err)
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
