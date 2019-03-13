package chezmoi

import (
	"os"
	"testing"

	"github.com/d4l3k/messagediff"
)

func TestFileAttributes(t *testing.T) {
	for _, tc := range []struct {
		sourceName string
		fa         FileAttributes
	}{
		{
			sourceName: "foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0666,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "dot_foo",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0666,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "private_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0600,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "private_dot_foo",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0600,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "empty_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0666,
				Empty:    true,
				Template: false,
			},
		},
		{
			sourceName: "executable_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0777,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "foo.tmpl",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0666,
				Empty:    false,
				Template: true,
			},
		},
		{
			sourceName: "private_executable_dot_foo.tmpl",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0700,
				Empty:    false,
				Template: true,
			},
		},
		{
			sourceName: "symlink_foo",
			fa: FileAttributes{
				Name: "foo",
				Mode: os.ModeSymlink | 0666,
			},
		},
		{
			sourceName: "symlink_dot_foo",
			fa: FileAttributes{
				Name: ".foo",
				Mode: os.ModeSymlink | 0666,
			},
		},
		{
			sourceName: "symlink_foo.tmpl",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     os.ModeSymlink | 0666,
				Template: true,
			},
		},
		{
			sourceName: "encrypted_private_dot_secret_file",
			fa: FileAttributes{
				Name:      ".secret_file",
				Mode:      0600,
				Encrypted: true,
			},
		},
	} {
		t.Run(tc.sourceName, func(t *testing.T) {
			gotFA := ParseFileAttributes(tc.sourceName)
			if diff, equal := messagediff.PrettyDiff(tc.fa, gotFA); !equal {
				t.Errorf("ParseFileAttributes(%q) == %+v, want %+v, diff:\n%s", tc.sourceName, gotFA, tc.fa, diff)
			}
			if gotSourceName := tc.fa.SourceName(); gotSourceName != tc.sourceName {
				t.Errorf("%+v.SourceName() == %q, want %q", tc.fa, gotSourceName, tc.sourceName)
			}
		})
	}
}
