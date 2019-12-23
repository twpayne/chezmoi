// +build !linux

package cmd

import vfs "github.com/twpayne/go-vfs"

func getOSRelease(fs vfs.FS) (map[string]string, error) {
	return nil, nil
}
