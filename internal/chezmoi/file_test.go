package chezmoi

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
				Mode:     0o666,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "dot_foo",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0o666,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "private_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0o600,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "private_dot_foo",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0o600,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "empty_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0o666,
				Empty:    true,
				Template: false,
			},
		},
		{
			sourceName: "executable_foo",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0o777,
				Empty:    false,
				Template: false,
			},
		},
		{
			sourceName: "foo.tmpl",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     0o666,
				Empty:    false,
				Template: true,
			},
		},
		{
			sourceName: "private_executable_dot_foo.tmpl",
			fa: FileAttributes{
				Name:     ".foo",
				Mode:     0o700,
				Empty:    false,
				Template: true,
			},
		},
		{
			sourceName: "symlink_foo",
			fa: FileAttributes{
				Name: "foo",
				Mode: os.ModeSymlink | 0o666,
			},
		},
		{
			sourceName: "symlink_dot_foo",
			fa: FileAttributes{
				Name: ".foo",
				Mode: os.ModeSymlink | 0o666,
			},
		},
		{
			sourceName: "symlink_foo.tmpl",
			fa: FileAttributes{
				Name:     "foo",
				Mode:     os.ModeSymlink | 0o666,
				Template: true,
			},
		},
		{
			sourceName: "encrypted_private_dot_secret_file",
			fa: FileAttributes{
				Name:      ".secret_file",
				Mode:      0o600,
				Encrypted: true,
			},
		},
	} {
		t.Run(tc.sourceName, func(t *testing.T) {
			assert.Equal(t, tc.fa, ParseFileAttributes(tc.sourceName))
			assert.Equal(t, tc.sourceName, tc.fa.SourceName())
		})
	}
}
