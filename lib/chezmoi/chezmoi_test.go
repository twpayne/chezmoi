package chezmoi

import (
	"reflect"
	"testing"
)

func TestParseFileState(t *testing.T) {
	for _, tc := range []struct {
		filename string
		contents []byte
		data     interface{}
		wantErr  bool
		want     *FileState
	}{
		{
			filename: "foo",
			want: &FileState{
				Name: "foo",
				Mode: 0666,
			},
		},
		{
			filename: "foo",
			contents: []byte("bar"),
			want: &FileState{
				Name:     "foo",
				Mode:     0666,
				Contents: []byte("bar"),
			},
		},
		{
			filename: "foo.tmpl",
			contents: []byte("{{23 -}} < {{- 45}}"),
			want: &FileState{
				Name:     "foo",
				Mode:     0666,
				Contents: []byte("23<45"),
			},
		},
		{
			filename: "foo.tmpl",
			contents: []byte("{{.User}}"),
			data:     map[string]string{"User": "bar"},
			want: &FileState{
				Name:     "foo",
				Mode:     0666,
				Contents: []byte("bar"),
			},
		},
		{
			filename: "dot_bashrc",
			want: &FileState{
				Name: ".bashrc",
				Mode: 0666,
			},
		},
		{
			filename: "private_dot_netrc",
			want: &FileState{
				Name: ".netrc",
				Mode: 0600,
			},
		},
		{
			filename: "executable_foo",
			want: &FileState{
				Name: "foo",
				Mode: 0777,
			},
		},
		{
			filename: "foo.tmpl",
			want: &FileState{
				Name: "foo",
				Mode: 0666,
			},
		},
		{
			filename: "private_dot_bash_history.tmpl",
			want: &FileState{
				Name: ".bash_history",
				Mode: 0600,
			},
		},
	} {
		if got, gotErr := ParseFileState(tc.filename, tc.contents, tc.data); (gotErr != nil) != tc.wantErr || !reflect.DeepEqual(got, tc.want) {
			var wantErrStr string
			if tc.wantErr {
				wantErrStr = "!<nil>"
			} else {
				wantErrStr = "<nil>"
			}
			t.Errorf("ParseFileState(%q, %v, %v) == %v, %v, want %v, %v", tc.filename, tc.contents, tc.data, got, gotErr, tc.want, wantErrStr)
		}
	}
}
