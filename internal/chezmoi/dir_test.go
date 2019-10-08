package chezmoi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirAttributes(t *testing.T) {
	for _, tc := range []struct {
		sourceName string
		da         DirAttributes
	}{
		{
			sourceName: "foo",
			da: DirAttributes{
				Name: "foo",
				Perm: 0777,
			},
		},
		{
			sourceName: "dot_foo",
			da: DirAttributes{
				Name: ".foo",
				Perm: 0777,
			},
		},
		{
			sourceName: "private_foo",
			da: DirAttributes{
				Name: "foo",
				Perm: 0700,
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
				Name: ".foo",
				Perm: 0700,
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
			assert.Equal(t, tc.da, ParseDirAttributes(tc.sourceName))
			assert.Equal(t, tc.sourceName, tc.da.SourceName())
		})
	}
}
