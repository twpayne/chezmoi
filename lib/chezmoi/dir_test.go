package chezmoi

import (
	"testing"

	"github.com/d4l3k/messagediff"
)

func TestDirAttributes(t *testing.T) {
	for _, tc := range []struct {
		sourceName string
		da         DirAttributes
	}{
		{
			sourceName: "foo",
			da: DirAttributes{
				Name:  "foo",
				Exact: false,
				Perm:  0777,
			},
		},
		{
			sourceName: "dot_foo",
			da: DirAttributes{
				Name:  ".foo",
				Exact: false,
				Perm:  0777,
			},
		},
		{
			sourceName: "private_foo",
			da: DirAttributes{
				Name:  "foo",
				Exact: false,
				Perm:  0700,
			},
		},
		{
			sourceName: "exact_foo",
			da: DirAttributes{
				Name:  "foo",
				Exact: true,
				Perm:  0777,
			},
		},
		{
			sourceName: "private_dot_foo",
			da: DirAttributes{
				Name:  ".foo",
				Exact: false,
				Perm:  0700,
			},
		},
		{
			sourceName: "exact_private_dot_foo",
			da: DirAttributes{
				Name:  ".foo",
				Exact: true,
				Perm:  0700,
			},
		},
	} {
		t.Run(tc.sourceName, func(t *testing.T) {
			gotDA := ParseDirAttributes(tc.sourceName)
			if diff, equal := messagediff.PrettyDiff(tc.da, gotDA); !equal {
				t.Errorf("ParseDirAttributes(%q) == %+v, want %+v, diff:\n%s", tc.sourceName, gotDA, tc.da, diff)
			}
			if gotSourceName := tc.da.SourceName(); gotSourceName != tc.sourceName {
				t.Errorf("%+v.SourceName() == %q, want %q", tc.da, gotSourceName, tc.sourceName)
			}
		})
	}
}
