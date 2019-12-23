// +build !linux

package cmd

import "github.com/twpayne/go-vfs"

func getKernelInfo(fs vfs.FS) (map[string]string, error) {
	return nil, nil
}

func getOSRelease(fs vfs.FS) (map[string]string, error) {
	return nil, nil
}
